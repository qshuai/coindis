package routers

import (
	"github.com/astaxie/beego"
	"github.com/qshuai/coindis/controllers"
)

func init() {
	beego.Router("/test", &controllers.IndexController{})
}
