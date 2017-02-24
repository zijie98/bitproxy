/*
	代理程序句柄
*/
package manager

import (
	"rkproxy/proxy"
	"rkproxy/proxy/ss"
	"rkproxy/utils"
)

//	代理程序的接口
//
type Proxyer interface {
	Start() error
	Stop() error
	LocalPort() int
}

//	句柄接口
//
type ProxyHandler interface {
	ListeningPort() int
	Start() error
	Stop() error
	GetConfig() interface{}
}

//	句柄
//
type Handle struct {
	Config interface{}
	Proxy  Proxyer
}

func (this *Handle) ListeningPort() int {
	return this.Proxy.LocalPort()
}
func (this *Handle) Start() error {
	return this.Proxy.Start()
}
func (this *Handle) Stop() error {
	return this.Proxy.Stop()
}
func (this *Handle) GetConfig() interface{} {
	return this.Config
}

type HttpReproxyHandle struct {
	Handle
}

type TcpProxyHandle struct {
	Handle
}

type SsClientHandle struct {
	Handle
}

type SsServerHandle struct {
	Handle
}

func NewTcpProxy(config *TcpProxyConfig) ProxyHandler {
	proxy := proxy.NewTcpProxy(
		config.LocalPort,
		config.RemoteHost,
		config.RemotePort,
	)
	handle := &TcpProxyHandle{
		Handle: Handle{
			Config: config,
			Proxy:  proxy,
		},
	}
	return handle
}

func NewSsClient(config *SsClientConfig) *SsClientHandle {
	proxy := ss.NewClient(
		config.LocalPort,
		utils.JoinHostPort(config.RemoteHost, config.RemotePort),
		config.Password,
		config.Crypt,
	)
	handle := &SsClientHandle{
		Handle: Handle{
			Config: config,
			Proxy:  proxy,
		},
	}
	return handle
}

func NewSsServer(config *SsServerConfig) *SsServerHandle {
	proxy := ss.NewServer(
		config.LocalPort,
		config.Password,
		config.Crypt,
	)
	handle := &SsServerHandle{
		Handle: Handle{
			Config: config,
			Proxy:  proxy,
		},
	}
	return handle
}

func NewHttpReproxy(config *HttpReproxyConfig) *HttpReproxyHandle {
	proxy := proxy.NewHttpReproxy(
		config.LocalPort,
		config.RemoteHost,
		config.RemotePort,
	)
	handle := &HttpReproxyHandle{
		Handle: Handle{
			Config: config,
			Proxy:  proxy,
		},
	}
	return handle
}
