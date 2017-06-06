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
	"rkproxy/log"
)

type Manager struct {
	handles    map[uint]ProxyHandler
	configPath string

	log *log.Logger
}

func New(configPath string) (man *Manager) {
	man = &Manager{
		configPath: configPath,
		handles:    make(map[uint]ProxyHandler),
		log:        log.NewLogger("Manager"),
	}
	return man
}

func (this *Manager) GetHandles() map[uint]ProxyHandler {
	return this.handles
}

//	将配置文件格式化到配置
//
func (this *Manager) ParseConfig() (err error) {
	err = configor.Load(&Config, this.configPath)
	if Config.HttpReproxy != nil {
		for _, cfg := range Config.HttpReproxy {
			this.CreateHttpReproxy(&cfg)
		}
	}
	if Config.Stream != nil {
		for _, cfg := range Config.Stream {
			this.CreateStreamProxy(&cfg)
		}
	}
	if Config.SsClient != nil {
		this.CreateSsClient(Config.SsClient)
	}
	if Config.SsServer != nil {
		for _, cfg := range Config.SsServer {
			this.CreateSsServer(&cfg)
		}
	}
	return
}

//	保存配置到配置文件
//
func (this *Manager) SaveToConfig() error {
	b, err := Config.toBytes()
	if err != nil {
		this.log.Info("格式化配置{", this.configPath, "}出错:", err)
		return err
	}
	file, err := os.OpenFile(this.configPath, os.O_RDWR|os.O_CREATE, 0666)
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

//	创建Tcp代理
//
func (this *Manager) CreateStreamProxy(config *StreamProxyConfig) ProxyHandler {
	handle := NewStreamProxy(config)
	this.handles[config.LocalPort] = handle
	return handle
}

//	创建SS客户端
//
func (this *Manager) CreateSsClient(config *SsClientConfig) ProxyHandler {
	handle := NewSsClient(config)
	this.handles[config.LocalPort] = handle
	return handle
}

//	创建ss服务器端
//
func (this *Manager) CreateSsServer(config *SsServerConfig) ProxyHandler {
	handle := NewSsServer(config)
	this.handles[config.Port] = handle
	return handle
}

//	创建Http反向代理
//
func (this *Manager) CreateHttpReproxy(config *HttpReproxyConfig) ProxyHandler {
	handle := NewHttpReproxy(config)
	this.handles[config.LocalPort] = handle
	return handle
}
