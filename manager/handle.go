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
	LocalPort() uint
}

//	句柄接口
//
type ProxyHandler interface {
	ListeningPort() uint
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

func (this *Handle) ListeningPort() uint {
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

type StreamProxyHandle struct {
	Handle
}

type SsClientHandle struct {
	Handle
}

type SsServerHandle struct {
	Handle
}

func NewStreamProxy(config *StreamProxyConfig) *StreamProxyHandle {
	pxy := proxy.NewStremProxy(
		ss.NetProtocol(config.LocalNet),
		config.LocalPort,
		config.RemoteHost,
		config.RemotePort,
	)
	handle := &StreamProxyHandle{
		Handle: Handle{
			Config: config,
			Proxy:  pxy,
		},
	}
	return handle
}

func NewSsClient(config *SsClientConfig) *SsClientHandle {
	pxy := ss.NewClient(
		ss.NetProtocol(config.LocalNet),
		config.LocalPort,
		utils.JoinHostPort(config.ServerHost, config.ServerPort),
		ss.NetProtocol(config.ChannelNet),
		config.Password,
		config.Crypt,
	)
	handle := &SsClientHandle{
		Handle: Handle{
			Config: config,
			Proxy:  pxy,
		},
	}
	return handle
}

func NewSsServer(config *SsServerConfig) *SsServerHandle {
	pxy := ss.NewServer(
		ss.NetProtocol(config.ChannelNet),
		config.LocalPort,
		config.Password,
		config.Crypt,
	)
	handle := &SsServerHandle{
		Handle: Handle{
			Config: config,
			Proxy:  pxy,
		},
	}
	return handle
}

func NewHttpReproxy(config *HttpReproxyConfig) *HttpReproxyHandle {
	pxy := proxy.NewHttpReproxy(
		config.LocalPort,
		config.RemoteHost,
		config.RemotePort,
	)
	handle := &HttpReproxyHandle{
		Handle: Handle{
			Config: config,
			Proxy:  pxy,
		},
	}
	return handle
}
