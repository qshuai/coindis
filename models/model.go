package models

import (
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/qshuai/coindis/conf"

	_ "github.com/go-sql-driver/mysql"
)

type History struct {
	Id      int
	Address string
	IP      string
	Amount  float64
	Updated time.Time `orm:"type(datetime);auto_now"`
	Created time.Time `orm:"auto_now_add;type(datetime)"`
}

func init() {
	orm.Debug = true
	config := conf.GetConfig()

	//get mysql configuration
	username := config.String("mysql::username")
	password := config.String("mysql::password")
	host := config.String("mysql::host")
	port := config.String("mysql::port")
	database := config.String("mysql::database")
	err := orm.RegisterDataBase("default", "mysql", username+":"+password+"@tcp("+host+":"+port+")/"+database+"?charset=utf8mb4&loc=Asia%2FShanghai")
	if err != nil {
		panic(err)
	}

	orm.RegisterModel(new(History))
	//err = orm.RunSyncdb("default", false, true)
	//if err != nil {
	//	panic(err)
	//}
}
