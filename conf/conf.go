package conf

import (
	"github.com/astaxie/beego/config"
	"sync"
)

var (
	once sync.Once
	conf config.Configer
)

func GetConfig() config.Configer {
	if conf != nil {
		return conf
	}

	once.Do(func() {
		var err error
		conf, err = config.NewConfig("ini", "/Users/qshuai/project/go/src/github.com/qshuai/coindis/app.conf")
		if err != nil {
			panic(err)
		}
	})

	return conf
}
