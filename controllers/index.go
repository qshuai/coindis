package controllers

import (
	"time"

	"coindis/models"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/config"
	"github.com/astaxie/beego/orm"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcutil"
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

func (c *IndexController) Get() {
	datalist := models.GetHistoryLimit100()
	client = Client()
	amount, err := client.GetBalance("")
	if err != nil {
		amount = btcutil.Amount(balance)
	}

	c.Data["addr"] = conf.String("addr")
	c.Data["balance"] = amount.String()
	c.Data["list"] = datalist
	c.TplName = "index.html"
}

func (c *IndexController) Post() {
	addr := c.GetString("address")
	amount, err := c.GetFloat("amount")
	if err != nil {
		r := Response{1, "Get Amount error"}
		c.Data["json"] = r
		c.ServeJSON()
	}
	if amount > limit {
		r := Response{1, "Amount is too big"}
		c.Data["json"] = r
		c.ServeJSON()
	}

	address, err := btcutil.DecodeAddress(addr, &chaincfg.TestNet3Params)
	if err != nil {
		r := Response{1, "Decode Address error"}
		c.Data["json"] = r
		c.ServeJSON()
	}

	ip := c.Ctx.Input.IP()
	hisrecoder, ok := models.ReturnTimeIfExist(addr, ip)
	if !ok {
		r := Response{1, "Create Transaction error"}
		c.Data["json"] = r
		c.ServeJSON()
	}

	if hisrecoder != nil {
		updated := hisrecoder.Updated
		if updated.Sub(time.Now()) < time.Duration(interval) {
			r := Response{1, "Request Interval less than one day"}
			c.Data["json"] = r
			c.ServeJSON()
		}
	}

	client = Client()
	txid, err := client.SendToAddress(address, btcutil.Amount(amount*1e8))
	if err != nil {
		r := Response{1, "Create Transaction error"}
		c.Data["json"] = r
		c.ServeJSON()
	}

	o := orm.NewOrm()
	if hisrecoder.Address == addr && hisrecoder.IP == ip {
		his := models.History{
			Amount: int64(amount * 1e8),
		}
		o.Update(his, "amount", "updated")
	} else {
		his := models.History{
			Address: address.String(),
			IP:      ip,
			Amount:  34,
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
}
