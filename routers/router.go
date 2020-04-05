package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/qshuai/coindis/controllers"
)

func RegisterApi(r *gin.Engine) {
	r.GET("/", controllers.Home)
	r.POST("/", controllers.FetchCoin)
}
