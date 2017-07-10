package api

import (
	"net/http"
	"os"

	"gopkg.in/gin-gonic/gin.v1"
	"rkproxy/utils"
)

var password string

func InitEngine() *gin.Engine {

	router := gin.Default()
	router.Use(validate)

	httpReproxy := router.Group("/HttpReproxy")
	{
		httpReproxy.POST("/", CreateHttpReproxy)
		httpReproxy.POST("/:action", ActionHttpReproxy)
	}
	ssServer := router.Group("/SsServer")
	{
		ssServer.POST("/", CreateSsServer)
		ssServer.POST("/:action", ActionSsServer)
	}
	stream := router.Group("/Stream")
	{
		stream.POST("/", CreateStream)
		stream.POST("/:action", ActionStream)
	}
	black := router.Group("/Black")
	{
		black.POST("/", CreateBlack)
		black.POST("/:action", ActionBlack)
	}
	return router
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
