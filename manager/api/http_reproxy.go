package api

import (
	"net/http"

	"gopkg.in/gin-gonic/gin.v1"
)

func CreateHttpReproxy(ctx *gin.Context) {

	ctx.String(http.StatusOK, "hello")
}

func ActionHttpReproxy(ctx *gin.Context) {

	ctx.String(http.StatusOK, "action")
}
