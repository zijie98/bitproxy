/*
	Manger负责管理所有的代理程序

	m = proxy.NewManager("xx.config")
	m.ParseConfig()
	m.RunAll()
*/
package manager

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/jinzhu/configor"
	"github.com/molisoft/bitproxy/utils"
)

var Man *Manager

const ETC_PATH = "/etc"
const CONFIG_FILENAME = "config.json"
const PID_FILENAME = "bitproxy.pid"

type Manager struct {
	handles    map[uint]ProxyHandler
	ConfigPath string
	PidPath string
	WorkspacePath string

	log *utils.Logger
}

func New() {
	path, _ := exec.LookPath(os.Args[0])
	workspacePath := filepath.Dir(path)
	pidPath := filepath.Join(workspacePath, PID_FILENAME)

	configPath := filepath.Join(workspacePath, CONFIG_FILENAME)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		configPath = filepath.Join(ETC_PATH, CONFIG_FILENAME)
	}

	Man = &Manager{
		ConfigPath: configPath,
		PidPath: pidPath,
		WorkspacePath: workspacePath,

		handles:    make(map[uint]ProxyHandler),
		log:        utils.NewLogger("Manager"),
	}
}

func (this *Manager) Config() *ProxyConfig {
	return &Config
}

func (this *Manager) GetHandles() map[uint]ProxyHandler {
	return this.handles
}

func (this *Manager) FindProxyByPort(port uint) ProxyHandler {
	return this.handles[port]
}

func (this *Manager) DeleteByPort(port uint) {
	delete(this.handles, port)
}

//	将配置文件格式化到配置
//
func (this *Manager) ParseConfig() (err error) {
	err = configor.Load(&Config, this.ConfigPath)
	if err != nil {
		return err
	}
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
	if Config.FtpProxy != nil {
		for _, cfg := range Config.FtpProxy {
			this.CreateFtpProxy(&cfg)
		}
	}
	return
}

//	保存配置到配置文件
//
func (this *Manager) SaveToConfig() error {
	b, err := Config.toBytes()
	if err != nil {
		this.log.Info("格式化配置{", this.ConfigPath, "}出错:", err)
		return err
	}
	file, err := os.OpenFile(this.ConfigPath, os.O_RDWR|os.O_CREATE, 0666)
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

// 创建ftp代理
func (this *Manager) CreateFtpProxy(config *FtpProxyConfig) ProxyHandler {
	handle := NewFtpProxy(config)
	this.handles[config.LocalPort] = handle
	return handle
}
