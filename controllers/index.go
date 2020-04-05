package controllers

import (
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/gin-gonic/gin"
	"github.com/qshuai/coindis/models"
	"github.com/qshuai/coindis/pb"
	"github.com/qshuai/coindis/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
)

var (
	balance int64
	ic      *utils.InfoCache
)

func Home(ctx *gin.Context) {
	dataList, err := models.GetHistoryLimit100()
	if err != nil {
		// todo<qshuai> ErrorPage html
		return
	}

	ctx.HTML(http.StatusOK, "index.html", gin.H{
		"addr":    viper.GetString("faucet.addr"),
		"limit":   viper.GetString("faucet.limit"),
		"balance": btcutil.Amount(atomic.LoadInt64(&balance)).String(),
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
	if postToken != "" && postToken != viper.GetString("faucet.token") {
		r := Response{1, "invalid token"}
		ctx.JSON(http.StatusOK, gin.H{
			"json": r,
		})

		return
	}

	isValid := postToken != "" && viper.GetString("faucet.token") == postToken
	// get bitcoin address and ip from request body
	addr := strings.Trim(ctx.GetString("address"), " ")
	_, err := btcutil.DecodeAddress(addr, &chaincfg.TestNet3Params)
	if err != nil {
		logrus.Debugf("the input address: %s", addr)
		r := Response{1, "invalid bitcoin cash address"}
		ctx.JSON(http.StatusOK, gin.H{
			"json": r,
		})

		return
	}

	if !isValid && ic.IsExit(addr) {
		r := Response{1, "Do not request repeat!"}
		ctx.JSON(http.StatusOK, gin.H{
			"json": r,
		})

		return
	}

	if !isValid && ic.IsExit(ip) {
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
	if !isValid && amount > viper.GetFloat64("faucet.limit") {
		r := Response{1, "Amount is too big"}
		ctx.JSON(http.StatusOK, gin.H{
			"json": r,
		})

		return
	}

	// view database, refused when meeting less than one day's request
	existed := true
	his, err := models.ReturnTimeIfExist(addr, ip)
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
		diff := now.Sub(his.Updated).Hours()
		if diff < float64(viper.GetInt64("faucet.interval")) {
			r := Response{1, "request frequently"}
			ctx.JSON(http.StatusOK, gin.H{
				"json": r,
			})

			return
		}
	}

	// insert to cache， because it will be successful mostly!
	if !isValid {
		ic.InsertNew(addr, ip)
	}

	txid, err := SendCoin(addr, int64(amount*1e8))
	if err != nil {
		logrus.Errorf("create transaction error: %s", err)
		// unlucky, to remove the cache for request again.
		ic.RemoveOne(addr, ip)

		ctx.JSON(http.StatusOK, gin.H{
			"json": Response{1, "Create Transaction error"},
		})

		return
	}

	// update balance at this time
	atomic.SwapInt64(&balance, atomic.LoadInt64(&balance)-int64(amount*1e8))

	o := orm.NewOrm()
	if existed {
		his := models.History{
			Id:      his.Id,
			Address: addr,
			IP:      ip,
			Amount:  amount + his.Amount,
		}
		_, err = o.Update(&his, "amount", "address", "ip", "updated")
		if err != nil {
			logrus.Errorf("update entry error: %s", err)
		}
	} else {
		his := models.History{
			Address: addr,
			IP:      ip,
			Amount:  amount,
		}
		_, err = o.Insert(&his)
		if err != nil {
			logrus.Errorf("insert entry error: %s", err)
		}
	}

	r := Response{0, txid}
	ctx.JSON(http.StatusOK, gin.H{
		"json": r,
	})
}

func UpdateBalance() error {
	client, err := utils.WalletClient(viper.GetString("wallet.host"), viper.GetString("wallet.port"))
	if err != nil {
		return err
	}

	amount, err := client.Balance(context.Background(), &pb.Empty{})
	if err != nil {
		return err
	}

	logrus.Info("update balance via rpc request")

	atomic.SwapInt64(&balance, int64(amount.Confirmed))
	return nil
}

func SendCoin(address string, amount int64) (string, error) {
	client, err := utils.WalletClient(viper.GetString("wallet.host"), viper.GetString("wallet.port"))
	if err != nil {
		return "", err
	}

	txId, err := client.Spend(context.Background(), &pb.SpendInfo{
		Address:  address,
		Amount:   uint64(amount),
		FeeLevel: pb.FeeLevel_ECONOMIC,
	})
	if err != nil {
		return "", err
	}

	return txId.String(), nil
}

func InitCache() {
	ic = utils.New(120)
}

func CleanCache() {
	ic.Clean()
}
