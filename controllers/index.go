package controllers

import (
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/config"
	"github.com/astaxie/beego/context"
	"github.com/astaxie/beego/orm"
	"github.com/bcext/cashutil"
	"github.com/bcext/gcash/chaincfg"
	"github.com/bcext/gcash/rpcclient"
	"github.com/gin-gonic/gin"
	"github.com/qshuai/coindis/models"
	"github.com/qshuai/coindis/utils"
	"github.com/sirupsen/logrus"
)

type IndexController struct {
	beego.Controller

	conf     config.Configer
	client   *rpcclient.Client
	balance  int64
	limit    float64
	interval int64
	ic       *utils.InfoCache
	token    string
}

type Response struct {
	Code    int
	Message string
}

func Home(ctx *gin.Context) {
	dataList, err := models.GetHistoryLimit100()
	if err != nil {
		// todo<qshuai> ErrorPage html
		return
	}

	ctx.HTML(http.StatusOK, "./views/index.html", gin.H{
		"addr":    c.conf.String("addr"),
		"limit":   limit,
		"balance": cashutil.Amount(atomic.LoadInt64(&c.balance)).String(),
		"list":    dataList,
	})
}

func FetchCoin(ctx *gin.Context) {
	ip := ctx.ClientIP()
	logrus.WithFields(logrus.Fields{
		"address": ctx.GetString("address"),
		"amount":  ctx.GetString("amount"),
		"ip":      ip,
	}).Debug("received post request")

	postToken := ctx.GetString("token")
	if postToken != "" && postToken != token {
		r := Response{1, "invalid token"}
		ctx.JSON(http.StatusOK, gin.H{
			"json": r,
		})

		return
	}

	isValid := postToken != "" && token == postToken
	// get bitcoin address and ip from request body
	addr := ctx.GetString("address")
	address, err := cashutil.DecodeAddress(addr, &chaincfg.TestNet3Params)
	if err != nil {
		logrus.Debugf("the input address: %s", addr)
		r := Response{1, "invalid bitcoin cash address"}
		ctx.JSON(http.StatusOK, gin.H{
			"json": r,
		})

		return
	}

	bech32Address := address.EncodeAddress(true)
	if !isValid && c.ic.IsExit(bech32Address) {
		r := Response{1, "Do not request repeat!"}
		ctx.JSON(http.StatusOK, gin.H{
			"json": r,
		})

		return
	}

	if !isValid && c.ic.IsExit(ip) {
		r := Response{1, "Do not request repeat!"}
		ctx.JSON(http.StatusOK, gin.H{
			"json": r,
		})

		return
	}

	// get and parse amount
	amount, err := strconv.ParseFloat(ctx.GetString("amount"), 10)
	if err != nil {
		logrus.Debugf("the input amount invalid: %s", amount)
		r := Response{1, "Get Amount error"}
		ctx.JSON(http.StatusOK, gin.H{
			"json": r,
		})

		return
	}
	if !isValid && amount > c.limit {
		r := Response{1, "Amount is too big"}
		ctx.JSON(http.StatusOK, gin.H{
			"json": r,
		})

		return
	}

	// view database, refused when meeting less than one day's request
	existed := true
	hisrecoder, err := models.ReturnTimeIfExist(bech32Address, ip)
	if err != nil && err != orm.ErrNoRows {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"json": "deny to service",
		})

		return
	}

	if err == orm.ErrNoRows {
		existed = false
	} else {
		existed = true

		now := time.Now()
		diff := now.Sub(hisrecoder.Updated).Hours()
		if diff < float64(interval) {
			r := Response{1, "request frequently"}
			ctx.JSON(http.StatusOK, gin.H{
				"json": r,
			})

			return
		}
	}

	// insert to cache， because it will be successful mostly!
	if !isValid {
		c.ic.InsertNew(bech32Address, ip)
	}

	c.client = Client(c)
	txid, err := c.client.SendToAddress(address, cashutil.Amount(amount*1e8))
	if err != nil {
		logrus.Debugf("create transaction error: %v", err)
		// unlucky, to remove the cache for request again.
		c.ic.RemoveOne(bech32Address, ip)
		r := Response{1, "Create Transaction error"}
		c.Data["json"] = r
		c.ServeJSON()
		return
	}

	// update balance at this time
	atomic.SwapInt64(&c.balance, atomic.LoadInt64(&c.balance)-int64(amount*1e8))

	o := orm.NewOrm()
	if existed {
		his := models.History{
			Id:      hisrecoder.Id,
			Address: bech32Address,
			IP:      ip,
			Amount:  amount + hisrecoder.Amount,
		}
		_, err = o.Update(&his, "amount", "address", "ip", "updated")
		if err != nil {
			logrus.Errorf("update entry error: %s", err)
		}
	} else {
		his := models.History{
			Address: bech32Address,
			IP:      ip,
			Amount:  amount,
		}
		_, err = o.Insert(&his)
		if err != nil {
			logrus.Errorf("insert entry error: %s", err)
		}
	}

	r := Response{0, txid.String()}
	ctx.JSON(http.StatusOK, gin.H{
		"json": r,
	})
}

func updateBalance(c *IndexController) {
	client := Client(c)
	amount, err := client.GetBalance("")
	if err != nil {
		logrus.Error("update balance via rpc failed: " + err.Error())
		return
	}

	logrus.Info("update balance via rpc request")

	atomic.SwapInt64(&c.balance, int64(amount))
}

// Deprecated function, instead of gin.Context.ClientIP()
func getClientIP(ctx *context.Context) string {
	ip := ctx.Request.Header.Get("X-Forwarded-For")
	if strings.Contains(ip, "127.0.0.1") || ip == "" {
		ip = ctx.Request.Header.Get("X-real-ip")
	}

	if ip == "" {
		return "127.0.0.1"
	}

	return ip
}
