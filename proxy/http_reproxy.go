/*
	Http反向代理
*/

package proxy

import (
	"time"

	"rkproxy/log"
	"rkproxy/utils"

	"github.com/valyala/fasthttp"
)

type HttpReproxy struct {
	local_port  uint
	remote_host string
	remote_port uint
	from_name   string
	log         *log.Logger

	proxyClient *fasthttp.HostClient
}

func (this *HttpReproxy) reverseProxyHandler(ctx *fasthttp.RequestCtx) {
	this.log.Info("request : ", ctx.URI().String())
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
			this.log.Info("Request 502..")
			time.Sleep(2 * time.Second)
			retry++
		} else {
			break
		}
	}

	this.postprocessResponse(resp)
}

func (this *HttpReproxy) prepareRequest(req *fasthttp.Request) {
	// do not proxy "Connection" header.
	req.Header.Del("Connection")
	if len(req.Header.UserAgent()) <= 1 {
		req.Header.Set("User-Agent", "RKProxy1.0")
	}
	req.Header.Set("From", this.from_name)
}

func (this *HttpReproxy) postprocessResponse(resp *fasthttp.Response) {
	// do not proxy "Connection" header
	resp.Header.Del("Connection")

	resp.Header.Set("Server", "RKProxy1.0")
	resp.Header.Set("From", this.from_name)
}

func (this *HttpReproxy) Start() error {
	this.proxyClient = &fasthttp.HostClient{
		Addr: utils.JoinHostPort(this.remote_host, this.remote_port),
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

func (this *HttpReproxy) LocalPort() uint {
	return this.local_port
}

func NewHttpReproxy(local_port uint, remote_host string, remote_port uint, from_name string) *HttpReproxy {
	return &HttpReproxy{
		local_port:  local_port,
		remote_host: remote_host,
		remote_port: remote_port,
		from_name:   from_name,
		log:         log.NewLogger("HttpReverseProxy"),
	}
}
