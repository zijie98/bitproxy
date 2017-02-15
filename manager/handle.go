/*
	定义Handle
*/
package manager

import (
	"rkproxy/proxy"
	"rkproxy/proxy/ss"
	"rkproxy/utils"
)

type Proxyer interface {
	Start() error
	Stop() error
	LocalPort() int
}

type HttpReproxyHandle struct {
	Config *HttpReproxyConfig
	Proxy  Proxyer
}

type TcpProxyHandle struct {
	Config *TcpProxyConfig
	Proxy  Proxyer
}

type SsClientHandle struct {
	Config *SsClientConfig
	Proxy  Proxyer
}

type SsServerHandle struct {
	Config *SsServerConfig
	Proxy  Proxyer
}

type Handles struct {
	HttpReproxys map[int]*HttpReproxyHandle `json:"http-reproxy"`
	TcpProxys    map[int]*TcpProxyHandle    `json:"tcp"`
	SsClient     *SsClientHandle            `json:"ss-client"`
	SsServers    map[int]*SsServerHandle    `json:"ss-server"`
}

func NewTcpProxy(config *TcpProxyConfig) *TcpProxyHandle {
	proxy := proxy.NewTcpProxy(
		config.LocalPort,
		config.RemoteHost,
		config.RemotePort,
	)
	handle := &TcpProxyHandle{
		Config: config,
		Proxy:  proxy,
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
		Config: config,
		Proxy:  proxy,
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
		Config: config,
		Proxy:  proxy,
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
		Config: config,
		Proxy:  proxy,
	}
	return handle
}
