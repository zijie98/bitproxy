package api

import (
	"errors"
	"net/http"

	"github.com/molisoft/bitproxy/services"
	"gopkg.in/gin-gonic/gin.v1"
)

type BlackPost struct {
	Ips []string `json:"ips"`
}

func CreateBlack(ctx *gin.Context) {
	var ips BlackPost
	err := ctx.BindJSON(&ips)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "bind json error " + err.Error()})
		return
	}
	for _, ip := range ips.Ips {
		services.Wall.Black(ip)
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func ActionBlack(ctx *gin.Context) {
	var ips BlackPost
	err := ctx.BindJSON(&ips)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "bind json error " + err.Error()})
		return
	}

	action := ctx.Param("action")
	switch action {
	case "remove":
		for _, ip := range ips.Ips {
			services.Wall.Remove(ip)
		}
	default:
		err = errors.New("not found action")
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
	} else {
		ctx.JSON(http.StatusOK, gin.H{"message": "ok"})
	}
}
