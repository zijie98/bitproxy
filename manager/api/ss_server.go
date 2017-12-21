package api

import (
	"net/http"
	"time"

	"github.com/kataras/go-errors"
	"github.com/molisoft/bitproxy/manager"
	"gopkg.in/gin-gonic/gin.v1"
)

func CreateSsServer(ctx *gin.Context) {
	server_config := new(manager.SsServerConfig)
	err := ctx.BindJSON(server_config)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	removeSs(server_config)
	proxy := manager.Man.CreateProxy(server_config, true)
	e := make(chan error)
	go func() {
		e <- proxy.Start()
	}()
	select {
	case err = <-e:
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		}
		return
	case <-time.After(2 * time.Second):
		ctx.JSON(http.StatusOK, gin.H{"message": "ok"})
		return
	}
}

func ActionSsServer(ctx *gin.Context) {
	var serverConfig manager.SsServerConfig
	err := ctx.BindJSON(&serverConfig)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	switch ctx.Param("action") {
	case "stop":
		err = stopSs(&serverConfig)
	case "start":
		err = startSs(&serverConfig)
	case "traffic":
		var t uint64
		t, err = trafficSs(&serverConfig)
		ctx.JSON(http.StatusOK, gin.H{
			"traffic": t,
			"unit":    "byte",
			"message": "ok",
		})
		return
	case "remove":
		err = removeSs(&serverConfig)
	default:
		err = errors.New("not found action")
	}
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"message": err.Error()})
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
	if proxy == nil {
		return errors.New("无法找到该ss代理")
	}
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
	if proxy == nil {
		return 0, errors.New("无法找到该ss代理")
	}
	return proxy.GetTraffic()
}

func removeSs(config *manager.SsServerConfig) error {
	proxy := manager.Man.FindProxyByPort(config.Port)
	if proxy == nil {
		return errors.New("无法找到该ss代理")
	}
	manager.Man.DeleteByPort(config.Port)
	return nil
}
