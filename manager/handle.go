/*
	代理程序句柄
*/
package manager

import (
	"github.com/molisoft/bitproxy/proxy"
	"github.com/molisoft/bitproxy/proxy/ss"
	"github.com/molisoft/bitproxy/utils"
)

type ProxyStatus int

const (
	STOP ProxyStatus = iota
	RUNING
)

//	句柄
//
type Handle struct {
	Config interface{}
	Proxy  proxy.Proxyer

	status ProxyStatus
}

func (this *Handle) Port() uint {
	return this.Proxy.LocalPort()
}

func (this *Handle) Start() error {
	if this.status == RUNING {
		return nil
	} else {
		this.status = RUNING
	}
	return this.Proxy.Start()
}

func (this *Handle) Stop() error {
	if this.status == STOP {
		return nil
	} else {
		this.status = STOP
	}
	return this.Proxy.Stop()
}

func (this *Handle) GetConfig() interface{} {
	return this.Config
}

func (this *Handle) GetTraffic() (uint64, error) {
	return this.Proxy.Traffic()
}

func NewStreamProxy(config *StreamProxyConfig) *Handle {
	pxy := proxy.NewStreamProxy(
		proxy.NetProtocol(config.LocalNet),
		config.LocalPort,
		config.ServerHost,
		config.ServerPort,
		config.Rate,
		config.EnableTraffic,
	)
	handle := &Handle{
		Config: config,
		Proxy:  pxy,
	}
	return handle
}

func NewSsClient(config *SsClientConfig) *Handle {
	pxy := ss.NewClient(
		proxy.NetProtocol(config.LocalNet),
		config.LocalPort,
		utils.JoinHostPort(config.ServerHost, config.ServerPort),
		proxy.NetProtocol(config.ChannelNet),
		config.Password,
		config.Crypt,
	)
	handle := &Handle{
		Config: config,
		Proxy:  pxy,
	}
	return handle
}

func NewSsServer(config *SsServerConfig) *Handle {
	pxy := ss.NewServer(
		proxy.NetProtocol(config.ChannelNet),
		config.Port,
		config.Password,
		config.Crypt,
		config.Rate,
		config.EnableTraffic,
	)
	handle := &Handle{
		Config: config,
		Proxy:  pxy,
	}
	return handle
}

func NewHttpReproxy(config *HttpReproxyConfig) *Handle {
	pxy := proxy.NewHttpReproxy(
		config.LocalPort,
		config.ServerHost,
		config.ServerPort,
		config.Name,
	)
	handle := &Handle{
		Config: config,
		Proxy:  pxy,
	}
	return handle
}

func NewFtpProxy(config *FtpProxyConfig) *Handle {
	pxy := proxy.NewFtpProxy(
		config.LocalPort,
		config.ServerHost,
		config.ServerPort,
	)
	handle := &Handle{
		Config: config,
		Proxy:  pxy,
	}
	return handle
}
