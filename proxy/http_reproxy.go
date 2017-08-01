/*
	Http反向代理
*/

package proxy

import (
	"net"
	"time"

	"rkproxy/libs"
	"rkproxy/utils"

	"github.com/valyala/fasthttp"
)

type HttpReproxy struct {
	local_port  uint
	remote_host string
	remote_port uint
	from_name   string
	log         *utils.Logger

	enable_black bool

	proxyClient *fasthttp.HostClient
}

var ReproxyUserAgent = "RKProxy"

func (this *HttpReproxy) reverseProxyHandler(ctx *fasthttp.RequestCtx) {
	this.log.Info(ctx.RemoteIP().String(), " - ", string(ctx.Method()), " - ", ctx.URI().String(), " - ", string(ctx.UserAgent()))

	if this.isBlack(ctx.RemoteAddr()) {
		this.log.Info("Blacked ", ctx.RemoteIP())
		return
	}

	retry := 0
	retry_count := 2
	req := &ctx.Request
	resp := &ctx.Response

	this.prepareRequest(req)

	for retry < retry_count {
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
	defer req.ResetBody()
	this.postprocessResponse(resp)
}

func (this *HttpReproxy) isBlack(addr net.Addr) bool {
	if !this.enable_black {
		return false
	}
	ip, _, _ := net.SplitHostPort(addr.String())
	if libs.Wall.IsBlack(ip) {
		return true
	}
	libs.Filter <- libs.RequestAt{
		Ip: ip,
		At: time.Now(),
	}
	return false
}

func (this *HttpReproxy) prepareRequest(req *fasthttp.Request) {
	// do not proxy "Connection" header.
	req.Header.Del("Connection")
	if len(req.Header.UserAgent()) <= 1 {
		req.Header.Set("User-Agent", ReproxyUserAgent)
	}
	req.Header.Set("From", this.from_name)
}

func (this *HttpReproxy) postprocessResponse(resp *fasthttp.Response) {
	// do not proxy "Connection" header
	resp.Header.Del("Connection")

	resp.Header.Set("Server", ReproxyUserAgent)
	resp.Header.Set("From", this.from_name)
}

func (this *HttpReproxy) Start() error {
	this.log.Info("Listen port", this.local_port)

	this.proxyClient = &fasthttp.HostClient{
		Addr:            utils.JoinHostPort(this.remote_host, this.remote_port),
		MaxConns:        512,
		MaxConnDuration: 60 * time.Second,
		ReadTimeout:     60 * time.Second,
		WriteTimeout:    60 * time.Second,
	}
	err := fasthttp.ListenAndServe(utils.JoinHostPort("", this.local_port), this.reverseProxyHandler)
	if err != nil {
		this.log.Info("HttpReverseProxy: ListenAndServe: ", err)
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
	return this.local_port
}

func NewHttpReproxy(local_port uint, remote_host string, remote_port uint, from_name string) *HttpReproxy {
	return &HttpReproxy{
		local_port:  local_port,
		remote_host: remote_host,
		remote_port: remote_port,
		from_name:   from_name,
		log:         utils.NewLogger("HttpReverseProxy"),
	}
}
