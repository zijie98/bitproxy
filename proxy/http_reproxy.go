/*
	Http反向代理
*/

package proxy

import (
	"net"
	"time"

	"github.com/molisoft/bitproxy/services"
	"github.com/molisoft/bitproxy/utils"
	"github.com/valyala/fasthttp"
)

type HttpReproxy struct {
	localPort  uint
	remoteHost string
	remotePort uint
	fromName   string
	log        *utils.Logger

	enableBlack bool

	proxyClient *fasthttp.HostClient
}

var ReproxyUserAgent = "BitProxy"

func (this *HttpReproxy) reverseProxyHandler(ctx *fasthttp.RequestCtx) {
	this.log.Info(ctx.RemoteIP().String(), " - ", string(ctx.Method()), " - ", ctx.URI().String(), " - ", string(ctx.UserAgent()))

	//if this.isBlack(ctx.RemoteAddr()) {
	//	this.log.Info("Blocked ", ctx.RemoteIP())
	//	return
	//}

	retry := 0
	retryCount := 2
	req := &ctx.Request
	resp := &ctx.Response

	this.prepareRequest(req, ctx)

	for retry < retryCount {
		if err := this.proxyClient.Do(req, resp); err != nil {
			this.log.Info("error when proxying the request: %s", err)
		}
		if resp.StatusCode() == fasthttp.StatusBadGateway {
			this.log.Info("Response status code 502.. retrying")
			time.Sleep(2 * time.Second)
			resp.ResetBody()
			retry++
		} else {
			break
		}
	}
	this.postprocessResponse(resp)
}

func (this *HttpReproxy) isBlack(addr net.Addr) bool {
	if !this.enableBlack {
		return false
	}
	ip, _, _ := net.SplitHostPort(addr.String())
	if services.Wall.IsBlack(ip) {
		return true
	}
	services.Filter <- services.RequestAt{
		Ip: ip,
		At: time.Now(),
	}
	return false
}

func (this *HttpReproxy) prepareRequest(req *fasthttp.Request, ctx *fasthttp.RequestCtx) {
	req.Header.Del("Connection")
	if len(req.Header.UserAgent()) <= 1 {
		req.Header.Set("User-Agent", ReproxyUserAgent)
	}
	req.Header.Set("From", this.fromName)
	req.Header.Set("X-Forwarded-For", ctx.RemoteIP().String())
}

func (this *HttpReproxy) postprocessResponse(resp *fasthttp.Response) {
	//resp.Header.Del("Connection")
	resp.Header.Set("Server", ReproxyUserAgent)
	resp.Header.Set("From", this.fromName)
}

func (this *HttpReproxy) Start() error {
	this.log.Info("Listen port", this.localPort)

	this.proxyClient = &fasthttp.HostClient{
		Addr:            utils.JoinHostPort(this.remoteHost, this.remotePort),
		MaxConns:        512,
		MaxConnDuration: 60 * time.Second,
		ReadTimeout:     60 * time.Second,
		WriteTimeout:    60 * time.Second,
	}

	s := &fasthttp.Server{
		Handler:              this.reverseProxyHandler,
		MaxKeepaliveDuration: 60 * time.Second,
		ReadTimeout:          60 * time.Second,
		WriteTimeout:         60 * time.Second,
		MaxConnsPerIP:        64,
		Logger:               utils.NewLogger("HttpReverseProxy"),
	}
	err := s.ListenAndServe(utils.JoinHostPort("", this.localPort))
	if err != nil {
		this.log.Info("ListenAndServe: ", err)
		return err
	}
	return nil
}

func (this *HttpReproxy) Stop() error {
	return nil
}

func (this *HttpReproxy) Traffic() (uint64, error) {
	return 0, nil
}

func (this *HttpReproxy) LocalPort() uint {
	return this.localPort
}

func NewHttpReproxy(localPort uint, remoteHost string, remotePort uint, fromName string) Proxyer {
	return &HttpReproxy{
		localPort:  localPort,
		remoteHost: remoteHost,
		remotePort: remotePort,
		fromName:   fromName,
		log:        utils.NewLogger("HttpReverseProxy"),
	}
}
