/*
	Http反向代理
*/

package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"rkproxy/log"
	"rkproxy/utils"
)

type HttpReproxy struct {
	local_port  int
	remote_host string
	remote_port int

	log *log.Logger
}

func (this *HttpReproxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	remote, err := url.Parse("http://" + utils.JoinHostPort(this.remote_host, this.remote_port))
	if err != nil {
		panic(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)
	this.log.Info("HttpReverseProxy: http proxy request : ", r.Host, " ", r.RequestURI)
	proxy.ServeHTTP(w, r)
}

func (this *HttpReproxy) Start() error {
	err := http.ListenAndServe(utils.JoinHostPort("", this.local_port), this)
	if err != nil {
		this.log.Info("HttpReverseProxy: ListenAndServe: ", err)
		return err
	}
	return nil
}

func (this *HttpReproxy) Stop() error {
	return nil
}

func (this *HttpReproxy) LocalPort() int {
	return this.local_port
}

func NewHttpReproxy(local_port int, remote_host string, remote_port int) *HttpReproxy {
	return &HttpReproxy{
		local_port:  local_port,
		remote_host: remote_host,
		remote_port: remote_port,
		log:         log.NewLogger("HttpReverseProxy"),
	}
}
