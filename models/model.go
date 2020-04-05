package models

import (
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/spf13/viper"

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

func SyncDataSource() error {
	orm.Debug = true

	//get mysql configuration
	username := viper.GetString("mysql.username")
	password := viper.GetString("mysql.password")
	host := viper.GetString("mysql.host")
	port := viper.GetString("mysql.port")
	database := viper.GetString("mysql.database")
	err := orm.RegisterDataBase("default", "mysql", username+":"+password+"@tcp("+host+":"+port+")/"+database+"?charset=utf8mb4&loc=Asia%2FShanghai")
	if err != nil {
		return err
	}

	orm.RegisterModel(new(History))
	err = orm.RunSyncdb("default", false, true)
	return err
}
