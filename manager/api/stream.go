package api

import (
	"net/http"

	"github.com/kataras/go-errors"
	"gopkg.in/gin-gonic/gin.v1"

	"rkproxy/manager"
)

func CreateStream(ctx *gin.Context) {
	var config manager.StreamProxyConfig
	err := ctx.BindJSON(&config)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "bind json error " + err.Error()})
		return
	}
	err = manager.Man.CreateStreamProxy(&config).Start()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "create stream error " + err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func ActionStream(ctx *gin.Context) {
	var config manager.StreamProxyConfig
	err := ctx.BindJSON(&config)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "bind json error " + err.Error()})
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
	case "modify":
		err = modifyStream(&config)
	default:
		err = errors.New("not found action")
	}
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
	} else {
		ctx.JSON(http.StatusOK, gin.H{"message": "ok"})
	}
}

func stopStream(config *manager.StreamProxyConfig) error {
	proxy := manager.Man.FindProxyByPort(config.LocalPort)
	return proxy.Stop()
}

func startStream(config *manager.StreamProxyConfig) error {
	proxy := manager.Man.FindProxyByPort(config.LocalPort)
	return proxy.Start()
}

func trafficStream(config *manager.StreamProxyConfig) (uint64, error) {
	proxy := manager.Man.FindProxyByPort(config.LocalPort)
	return proxy.GetTraffic()
}

func removeStream(config *manager.StreamProxyConfig) error {
	err := stopStream(config)
	if err != nil {
		return err
	}
	manager.Man.DeleteByPort(config.LocalPort)
	return nil
}

func modifyStream(config *manager.StreamProxyConfig) error {
	err := removeStream(config)
	if err != nil {
		return err
	}
	return manager.Man.CreateStreamProxy(config).Start()
}
