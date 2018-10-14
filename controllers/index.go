package controllers

import (
	"sync/atomic"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/config"
	"github.com/astaxie/beego/orm"
	"github.com/bcext/cashutil"
	"github.com/bcext/gcash/chaincfg"
	"github.com/bcext/gcash/rpcclient"
	"github.com/qshuai/coindis/models"
)

type IndexController struct {
	beego.Controller
}

type Response struct {
	Code    int
	Message string
}

var conf config.Configer
var client *rpcclient.Client
var balance int64
var limit float64
var interval int64
var ic = newInfoCache()

func (c *IndexController) Get() {
	dataList := models.GetHistoryLimit100()

	old := atomic.LoadInt64(&balance)
	if old <= 1 {
		client = Client()
		amount, err := client.GetBalance("")
		if err != nil {
			amount = cashutil.Amount(old)
		}

		// cache balance
		atomic.SwapInt64(&balance, int64(amount))
	}

	c.Data["addr"] = conf.String("addr")
	c.Data["balance"] = cashutil.Amount(atomic.LoadInt64(&balance)).String()
	c.Data["limit"] = limit
	c.Data["list"] = dataList
	c.TplName = "index.html"
}

func (c *IndexController) Post() {
	// get bitcoin address and ip from request body
	addr := c.GetString("address")
	if ic.isExit(addr) {
		r := Response{1, "Do not request repeat!"}
		c.Data["json"] = r
		c.ServeJSON()
		return
	}

	// get and parse amount
	amount, err := c.GetFloat("amount")
	if err != nil {
		r := Response{1, "Get Amount error"}
		c.Data["json"] = r
		c.ServeJSON()
		return
	}
	if amount > limit {
		r := Response{1, "Amount is too big"}
		c.Data["json"] = r
		c.ServeJSON()
		return
	}

	// view database, refused for less than one day's request
	hisrecoder := models.ReturnTimeIfExist(addr)

	address, err := cashutil.DecodeAddress(addr, &chaincfg.TestNet3Params)
	if err != nil {
		r := Response{1, "Decode Address error"}
		c.Data["json"] = r
		c.ServeJSON()
		return
	}

	if hisrecoder != nil {
		now := time.Now()
		if now.Sub(hisrecoder.Updated) < time.Duration(interval) {
			r := Response{1, "Request Interval less than one day"}
			c.Data["json"] = r
			c.ServeJSON()
			return
		}
	}

	// insert to cache， because it will be successful mostly!
	ic.insertNew(addr)

	client = Client()
	txid, err := client.SendToAddress(address, cashutil.Amount(amount*1e8))
	if err != nil {
		// unlucky, to remove the cache for request again.
		ic.removeOne(addr)
		r := Response{1, "Create Transaction error"}
		c.Data["json"] = r
		c.ServeJSON()
		return
	}

	// update balance at this time
	atomic.SwapInt64(&balance, atomic.LoadInt64(&balance)-int64(amount*1e8))

	o := orm.NewOrm()
	if hisrecoder != nil && hisrecoder.Address == addr {
		his := models.History{
			Amount: amount,
		}
		o.Update(&his, "amount", "updated")
	} else {
		his := models.History{
			Address: address.EncodeAddress(true),
			Amount:  amount,
		}
		o.Insert(&his)
	}

	r := Response{0, txid.String()}
	c.Data["json"] = r
	c.ServeJSON()
}

func Client() *rpcclient.Client {
	if client != nil {
		return client
	}
	// acquire configure item
	link := conf.String("rpc::url") + ":" + conf.String("rpc::port")
	user := conf.String("rpc::user")
	passwd := conf.String("rpc::passwd")

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
	c, err := rpcclient.New(connCfg, nil)
	if err != nil {
		panic(err)
	}

	client = c
	return c
}

func init() {
	var err error
	conf, err = config.NewConfig("ini", "conf/app.conf")
	if err != nil {
		panic(err)
	}

	limit, err = conf.Float("limit")
	if err != nil {
		panic(err)
	}

	interval, err = conf.Int64("interval")
	if err != nil {
		panic(err)
	}

	go func() {
		ticker := time.NewTicker(time.Minute * 10)
		for _ = range ticker.C {
			updateBalance()
			ic.clean()
		}
	}()
}
