package main

import (
	"github.com/astaxie/beego"
	_ "github.com/qshuai/coindis/routers"
)

func main() {
	beego.Run(":8000")
}
