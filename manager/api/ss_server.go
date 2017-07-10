package api

import (
	"net/http"

	"github.com/kataras/go-errors"
	"gopkg.in/gin-gonic/gin.v1"
	"rkproxy/manager"
	"time"
)

func CreateSsServer(ctx *gin.Context) {
	var server_config manager.SsServerConfig
	err := ctx.BindJSON(&server_config)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "bind json error " + err.Error()})
		return
	}
	e := make(chan error)
	go func() {
		e <- manager.Man.CreateSsServer(&server_config).Start()
	}()
	select {
	case err = <-e:
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "create ss server error ：" + err.Error()})
		return
	case <-time.After(time.Second * 1):
		ctx.JSON(http.StatusOK, gin.H{"message": "ok"})
	}
}

func ActionSsServer(ctx *gin.Context) {
	var server_config manager.SsServerConfig
	err := ctx.BindJSON(&server_config)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "bind json error " + err.Error()})
		return
	}
	switch ctx.Param("action") {
	case "stop":
		err = stopSs(&server_config)
	case "start":
		err = startSs(&server_config)
	case "traffic":
		var t uint64
		t, err = trafficSs(&server_config)
		ctx.JSON(http.StatusOK, gin.H{
			"traffic": t,
			"unit":    "byte",
			"message": "ok",
		})
		return
	case "remove":
		err = removeSs(&server_config)
	case "modify":
		err = modifySs(&server_config)
	default:
		err = errors.New("not found action")
	}
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
	} else {
		ctx.JSON(http.StatusOK, gin.H{"message": "ok"})
	}
}

func stopSs(config *manager.SsServerConfig) error {
	proxy := manager.Man.FindProxyByPort(config.Port)
	if proxy == nil {
		return errors.New("无法找到该ss代理")
	}
	return proxy.Stop()
}

func startSs(config *manager.SsServerConfig) error {
	proxy := manager.Man.FindProxyByPort(config.Port)
	err := make(chan error)
	go func() {
		err <- proxy.Start()
	}()
	select {
	case e := <-err:
		return e
	case <-time.After(time.Second * 1):
		return nil
	}
}

func trafficSs(config *manager.SsServerConfig) (uint64, error) {
	proxy := manager.Man.FindProxyByPort(config.Port)
	return proxy.GetTraffic()
}

func removeSs(config *manager.SsServerConfig) error {
	err := stopSs(config)
	if err != nil {
		return err
	}
	manager.Man.DeleteByPort(config.Port)
	return nil
}

func modifySs(config *manager.SsServerConfig) error {
	err := removeSs(config)
	if err != nil {
		return err
	}
	return manager.Man.CreateSsServer(config).Start()
}
