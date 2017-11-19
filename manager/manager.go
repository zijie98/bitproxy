/*
	Manger负责管理所有的代理程序

	m = proxy.NewManager("xx.config")
	m.ParseConfig()
	m.RunAll()
*/
package manager

import (
	"os"

	"github.com/jinzhu/configor"
	"github.com/molisoft/bitproxy/proxy"
	"github.com/molisoft/bitproxy/utils"
)

var Man *Manager

type Manager struct {
	handles       map[uint]proxy.ProxyHandler
	Config        *ProxyConfig
	ConfigPath    string
	PidPath       string
	WorkspacePath string

	log *utils.Logger
}

func New(appPath, pidPath, configPath string) {
	Man = &Manager{
		ConfigPath:    configPath,
		PidPath:       pidPath,
		WorkspacePath: appPath,
		Config:        new(ProxyConfig),

		handles: make(map[uint]proxy.ProxyHandler),
		log:     utils.NewLogger("Manager"),
	}
}

func (this *Manager) GetHandles() map[uint]proxy.ProxyHandler {
	return this.handles
}

func (this *Manager) FindProxyByPort(port uint) proxy.ProxyHandler {
	h, ok := this.handles[port]
	if ok {
		return h
	}
	return nil
}

func (this *Manager) DeleteByPort(port uint) {
	h := this.FindProxyByPort(port)
	if h == nil {
		return
	}
	h.Stop()
	delete(this.handles, port)

	if this.Config.HttpReproxy != nil {
		for i, cfg := range this.Config.HttpReproxy {
			if cfg.LocalPort == port {
				this.Config.HttpReproxy = append(this.Config.HttpReproxy[:i], this.Config.HttpReproxy[i+1:]...)
				return
			}
		}
	}
	if this.Config.FtpProxy != nil {
		for i, cfg := range this.Config.FtpProxy {
			if cfg.LocalPort == port {
				this.Config.FtpProxy = append(this.Config.FtpProxy[:i], this.Config.FtpProxy[i+1:]...)
				return
			}
		}
	}
	if this.Config.SsServer != nil {
		for i, cfg := range this.Config.SsServer {
			if cfg.Port == port {
				this.Config.SsServer = append(this.Config.SsServer[:i], this.Config.SsServer[i+1:]...)
				return
			}
		}
	}
	if this.Config.Stream != nil {
		for i, cfg := range this.Config.Stream {
			if cfg.LocalPort == port {
				this.Config.Stream = append(this.Config.Stream[:i], this.Config.Stream[i+1:]...)
				return
			}
		}
	}
	if this.Config.SsClient != nil && this.Config.SsClient.LocalPort == port {
		this.Config.SsClient = nil
		return
	}

	this.SaveToConfig()
}

//	将配置文件格式化到配置
//
func (this *Manager) ParseConfig() (err error) {
	err = configor.Load(this.Config, this.ConfigPath)
	if err != nil {
		return err
	}
	if this.Config.HttpReproxy != nil {
		for _, cfg := range this.Config.HttpReproxy {
			this.CreateProxy(&cfg, false)
		}
	}
	if this.Config.Stream != nil {
		for _, cfg := range this.Config.Stream {
			this.CreateProxy(&cfg, false)
		}
	}
	if this.Config.SsClient != nil {
		this.CreateProxy(this.Config.SsClient, false)
	}
	if this.Config.SsServer != nil {
		for _, cfg := range this.Config.SsServer {
			this.CreateProxy(&cfg, false)
		}
	}
	if this.Config.FtpProxy != nil {
		for _, cfg := range this.Config.FtpProxy {
			this.CreateProxy(&cfg, false)
		}
	}
	return
}

func (this *Manager) CreateProxy(config interface{}, appendToConfig bool) (handler proxy.ProxyHandler) {
	switch config.(type) {
	case *StreamProxyConfig:
		handler = NewStreamProxy(config.(*StreamProxyConfig))
		if appendToConfig {
			this.Config.Stream = append(this.Config.Stream, *config.(*StreamProxyConfig))
		}
	case *SsClientConfig:
		handler = NewSsClient(config.(*SsClientConfig))
		if appendToConfig {
			this.Config.SsClient = config.(*SsClientConfig)
		}
	case *SsServerConfig:
		handler = NewSsServer(config.(*SsServerConfig))
		if appendToConfig {
			this.Config.SsServer = append(this.Config.SsServer, *config.(*SsServerConfig))
		}
	case *HttpReproxyConfig:
		handler = NewHttpReproxy(config.(*HttpReproxyConfig))
		if appendToConfig {
			this.Config.HttpReproxy = append(this.Config.HttpReproxy, *config.(*HttpReproxyConfig))
		}
	case *FtpProxyConfig:
		handler = NewFtpProxy(config.(*FtpProxyConfig))
		if appendToConfig {
			this.Config.FtpProxy = append(this.Config.FtpProxy, *config.(*FtpProxyConfig))
		}
	}
	if handler != nil {
		this.handles[handler.Port()] = handler
	}
	return
}

//	保存配置到配置文件
//
func (this *Manager) SaveToConfig() error {
	b, err := this.Config.toBytes()
	if err != nil {
		this.log.Info("格式化配置{", this.ConfigPath, "}出错:", err)
		return err
	}
	os.Remove(this.ConfigPath)

	file, err := os.OpenFile(this.ConfigPath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		this.log.Info("创建配置文件失败: ", err)
		return err
	}
	defer file.Close()

	_, err = file.Write(b)
	if err != nil {
		this.log.Info("写入配置到配置文件出错：", err)
	}
	return err
}

//	关闭所有的代理服务
//
func (this *Manager) StopAll() {
	if len(this.handles) > 0 {
		for _, handle := range this.handles {
			handle.Stop()
		}
	}
}

//	运行所有的代理服务
//
func (this *Manager) RunAll() {
	this.StopAll()

	if len(this.handles) > 0 {
		for _, handle := range this.handles {
			go handle.Start()
		}
	}
}

func (this *Manager) Stop(port uint) {
	handler := this.FindProxyByPort(port)
	if handler != nil {
		handler.Stop()
	}
}

func (this *Manager) Start(port uint) {
	handler := this.FindProxyByPort(port)
	if handler != nil {
		go handler.Start()
	}
}
