package models

import (
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
)

type History struct {
	Id      int
	Address string
	IP      string
	Amount  int64
	Updated time.Time `orm:"type(datetime);auto_now"`
	Created time.Time `orm:"auto_now_add;type(datetime)"`
}

func init() {
	orm.Debug = true

	//获取配置信息
	username := beego.AppConfig.String("mysql::username")
	password := beego.AppConfig.String("mysql::password")
	host := beego.AppConfig.String("mysql::host")
	port := beego.AppConfig.String("mysql::port")
	database := beego.AppConfig.String("mysql::database")
	orm.RegisterDataBase("default", "mysql", username+":"+password+"@tcp("+host+":"+port+")/"+database+"?charset=utf8&loc=Asia%2FShanghai")

	orm.RegisterModel(new(History))
	orm.RunSyncdb("default", false, true)
}
