package api

import (
	"fmt"
	"net/http"

	"github.com/molisoft/bitproxy/utils"
	"gopkg.in/gin-gonic/gin.v1"
)

var password string

func initEngine() *gin.Engine {

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
	//manager.Man.SaveToConfig()
}

func Start(pwd string, port uint) error {
	fmt.Println("Start Api service.. port ", port)
	password = pwd

	gin.DefaultWriter = utils.NewLogger("api")
	gin.DefaultErrorWriter = utils.NewLogger("api.error")

	server := initEngine()
	return server.Run(utils.JoinHostPort("", port))
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
