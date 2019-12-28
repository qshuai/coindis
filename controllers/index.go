package controllers

import (
	"github.com/qshuai/coindis/conf"
	"github.com/qshuai/coindis/utils"
	"os"
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
	"github.com/qshuai/coindis/models"
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

func (c *IndexController) Get() {
	dataList := models.GetHistoryLimit100()

	c.Data["addr"] = c.conf.String("addr")
	c.Data["limit"] = c.limit
	c.Data["balance"] = cashutil.Amount(atomic.LoadInt64(&c.balance)).String()
	c.Data["list"] = dataList
	c.TplName = "index.html"
}

func (c *IndexController) Post() {
	ip := getClientIP(c.Ctx)

	logrus.WithFields(logrus.Fields{
		"address": c.GetString("address"),
		"amount":  c.GetString("amount"),
		"ip":      ip,
	}).Debug("received post request")

	postToken := c.GetString("token")
	if postToken != "" && postToken != c.token {
		r := Response{1, "invalid token"}
		c.Data["json"] = r
		c.ServeJSON()
		return
	}
	isValid := postToken != "" && c.token == postToken

	// get bitcoin address and ip from request body
	addr := c.GetString("address")
	address, err := cashutil.DecodeAddress(addr, &chaincfg.TestNet3Params)
	if err != nil {
		logrus.Debugf("the input address: %s", addr)
		r := Response{1, "The address not correct"}
		c.Data["json"] = r
		c.ServeJSON()
		return
	}

	bech32Address := address.EncodeAddress(true)
	if !isValid && c.ic.IsExit(bech32Address) {
		r := Response{1, "Do not request repeat!"}
		c.Data["json"] = r
		c.ServeJSON()
		return
	}

	if !isValid && c.ic.IsExit(ip) {
		r := Response{1, "Do not request repeat!"}
		c.Data["json"] = r
		c.ServeJSON()
		return
	}

	// get and parse amount
	amount, err := c.GetFloat("amount")
	if err != nil {
		logrus.Debugf("the input amount: %s", amount)
		r := Response{1, "Get Amount error"}
		c.Data["json"] = r
		c.ServeJSON()
		return
	}
	if !isValid && amount > c.limit {
		r := Response{1, "Amount is too big"}
		c.Data["json"] = r
		c.ServeJSON()
		return
	}

	// view database, refused for less than one day's request
	hisrecoder := models.ReturnTimeIfExist(bech32Address, ip)
	if !isValid && hisrecoder != nil {
		now := time.Now()
		diff := now.Sub(hisrecoder.Updated).Hours()
		if diff < float64(c.interval) {
			r := Response{1, "request frequently"}
			c.Data["json"] = r
			c.ServeJSON()
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
	if hisrecoder != nil && hisrecoder.Address == bech32Address {
		his := models.History{
			Id:      hisrecoder.Id,
			Address: bech32Address,
			IP:      ip,
			Amount:  amount + hisrecoder.Amount,
		}
		o.Update(&his, "amount", "address", "ip", "updated")
	} else {
		his := models.History{
			Address: bech32Address,
			IP:      ip,
			Amount:  amount,
		}
		o.Insert(&his)
	}

	r := Response{0, txid.String()}
	c.Data["json"] = r
	c.ServeJSON()
}

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

func Client(c *IndexController) *rpcclient.Client {
	if c.client != nil {
		return c.client
	}

	// acquire configure item
	link := c.conf.String("rpc::url") + ":" + c.conf.String("rpc::port")
	user := c.conf.String("rpc::user")
	passwd := c.conf.String("rpc::passwd")

	// rpc client instance
	connCfg := &rpcclient.ConnConfig{
		Host:         link,
		User:         user,
		Pass:         passwd,
		HTTPPostMode: true, // Bitcoin core only supports HTTP POST mode
		DisableTLS:   true, // Bitcoin core does not provide TLS by default
	}
	// Notice the notification parameter is nil since notifications are
	// not supported in HTTP POST mode.
	var err error
	c.client, err = rpcclient.New(connCfg, nil)
	if err != nil {
		logrus.Error(err.Error())
		panic(err)
	}

	err = c.client.Ping()
	if err != nil {
		logrus.Error("rpc connection error: " + err.Error())
		os.Exit(1)
	}

	return c.client
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

func init() {
	var err error
	var controller IndexController
	controller.conf = conf.GetConfig()

	level, err := logrus.ParseLevel(controller.conf.String("log::level"))
	if err != nil {
		panic(err)
	}
	logrus.SetLevel(level)

	_, err = os.Stat("./logs/")
	if os.IsNotExist(err) {
		os.MkdirAll("./logs", 0766)
	}

	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true, TimestampFormat: "2006-01-02 15:04:05",
		DisableTimestamp: false})

	file, err := os.OpenFile("./logs/"+controller.conf.String("log::filename")+".log",
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.FileMode(0644))

	if err != nil {
		panic(err)
	}
	logrus.SetOutput(file)

	// init balance
	updateBalance(&controller)

	controller.limit, err = controller.conf.Float("limit")
	if err != nil {
		panic(err)
	}

	// should justify whether the token configuration is empty or not
	controller.token = controller.conf.String("token")

	controller.interval, err = controller.conf.Int64("interval")
	if err != nil {
		panic(err)
	}

	go func() {
		ticker := time.NewTicker(time.Second * 30)
		defer func() {
			ticker.Stop()
			logrus.Error("the goroutine holds the ticker for updating balance existed")
		}()

		for _ = range ticker.C {
			updateBalance(&controller)
		}
	}()

	go func() {
		ticker := time.NewTicker(time.Minute * 10)
		defer func() {
			ticker.Stop()
			logrus.Error("the goroutine holds the ticker for cache clean existed")
		}()

		for _ = range ticker.C {
			controller.ic.Clean()
		}
	}()

	logrus.Info("the coindis program started!")
}
