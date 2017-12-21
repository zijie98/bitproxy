package api

import (
	"net/http"
	"time"

	"github.com/kataras/go-errors"
	"github.com/molisoft/bitproxy/manager"
	"gopkg.in/gin-gonic/gin.v1"
)

func CreateStream(ctx *gin.Context) {
	var config manager.StreamProxyConfig
	err := ctx.BindJSON(&config)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"message": "bind json error " + err.Error()})
		return
	}
	removeStream(&config)
	proxy := manager.Man.CreateProxy(&config, true)
	e := make(chan error)
	go func() {
		e <- proxy.Start()
	}()
	select {
	case err = <-e:
		ctx.JSON(http.StatusOK, gin.H{"message": "create stream error " + err.Error()})
		return
	case <-time.After(2 * time.Second):
		ctx.JSON(http.StatusOK, gin.H{"message": "ok"})
		return
	}
}

func ActionStream(ctx *gin.Context) {
	var config manager.StreamProxyConfig
	err := ctx.BindJSON(&config)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"message": "bind json error " + err.Error()})
		return
	}
	switch ctx.Param("action") {
	case "stop":
		err = stopStream(&config)
	case "start":
		err = startStream(&config)
	case "traffic":
		var t uint64
		t, err = trafficStream(&config)
		ctx.JSON(http.StatusOK, gin.H{
			"traffic": t,
			"unit":    "byte",
			"message": "ok",
		})
		return
	case "remove":
		err = removeStream(&config)
	default:
		err = errors.New("not found action")
	}
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"message": err.Error()})
	} else {
		ctx.JSON(http.StatusOK, gin.H{"message": "ok"})
	}
}

func stopStream(config *manager.StreamProxyConfig) error {
	proxy := manager.Man.FindProxyByPort(config.LocalPort)
	if proxy == nil {
		return errors.New("not found")
	}
	return proxy.Stop()
}

func startStream(config *manager.StreamProxyConfig) error {
	proxy := manager.Man.FindProxyByPort(config.LocalPort)
	if proxy == nil {
		return errors.New("not found")
	}
	return proxy.Start()
}

func trafficStream(config *manager.StreamProxyConfig) (uint64, error) {
	proxy := manager.Man.FindProxyByPort(config.LocalPort)
	if proxy == nil {
		return 0, errors.New("not found")
	}
	return proxy.GetTraffic()
}

func removeStream(config *manager.StreamProxyConfig) error {
	proxy := manager.Man.FindProxyByPort(config.LocalPort)
	if proxy == nil {
		return errors.New("not found")
	}
	manager.Man.DeleteByPort(config.LocalPort)
	return nil
}
