package api

import (
	"net/http"

	"gopkg.in/gin-gonic/gin.v1"

	"rkproxy/manager"
	"rkproxy/utils"
)

var password string

func InitEngine() *gin.Engine {

	router := gin.Default()
	router.Use(validate)

	httpReproxy := router.Group("/HttpReproxy")
	{
		httpReproxy.POST("/", CreateHttpReproxy, afterFunc)
		httpReproxy.POST("/:action", ActionHttpReproxy, afterFunc)
	}
	ssServer := router.Group("/SsServer")
	{
		ssServer.POST("/", CreateSsServer, afterFunc)
		ssServer.POST("/:action", ActionSsServer, afterFunc)
	}
	stream := router.Group("/Stream")
	{
		stream.POST("/", CreateStream, afterFunc)
		stream.POST("/:action", ActionStream, afterFunc)
	}
	black := router.Group("/Black")
	{
		black.POST("/", CreateBlack)
		black.POST("/:action", ActionBlack)
	}
	return router
}

func afterFunc(ctx *gin.Context) {
	manager.Man.SaveToConfig()
}

func Start(pwd string, port uint) error {
	password = pwd

	gin.DefaultWriter = utils.NewLogger("api.log")
	gin.DefaultErrorWriter = utils.NewLogger("api.error.log")

	return InitEngine().Run(utils.JoinHostPort("", port))
}

func validate(ctx *gin.Context) {
	pwd := ctx.Request.Header.Get("USER-PWD")
	if password == "test" {
		return
	}
	if len(pwd) == 0 || pwd != password {
		ctx.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		ctx.Abort()
		return
	}
}
