package routers

import (
	"coindis/controllers"

	"github.com/astaxie/beego"
)

func init() {
	beego.Router("/", &controllers.IndexController{})
}
