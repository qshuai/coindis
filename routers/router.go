package routers

import (
	"coindis/controllers"

	"github.com/astaxie/beego"
)

func init() {
	beego.Router("/test", &controllers.IndexController{})
}
