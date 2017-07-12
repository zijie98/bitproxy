package api

import (
	"net/http"
	"os"

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
		black.POST("/", CreateBlack, afterFunc)
		black.POST("/:action", ActionBlack, afterFunc)
	}
	return router
}

func afterFunc(ctx *gin.Context) {
	manager.Man.SaveToConfig()
}

func Start(pwd string, port uint) error {
	password = pwd

	logfile, err := logfile()
	if err == nil {
		gin.DefaultWriter = logfile
	}

	return InitEngine().Run(utils.JoinHostPort("", port))
}

func logfile() (file *os.File, err error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	logPath := wd + "/api.log"
	file, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return
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
