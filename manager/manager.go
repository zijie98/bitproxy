/*
	Manger负责管理所有的代理程序

	m = proxy.NewManager("xx.config")
	m.ParseConfig()
	m.RunAll()
*/
package manager

import (
	"encoding/json"
	"io/ioutil"
	"rkproxy/log"
	"rkproxy/utils"

	"github.com/kataras/go-errors"
	"os"
	"strings"
)

type Manager struct {
	handles          map[uint]ProxyHandler
	config_file_path string
	config_content   []byte

	log *log.Logger
}

func New(config_path string) (man *Manager) {
	man = &Manager{
		config_file_path: config_path,
		handles:          make(map[uint]ProxyHandler),
		log:              log.NewLogger("Manager"),
	}
	man.readConfigByFilePath()
	return man
}

func NewByConfigContent(content []byte) (man *Manager) {
	man = &Manager{
		config_content: content,
		handles:        make(map[uint]ProxyHandler),
		log:            log.NewLogger("Manager"),
	}
	return man
}

func (this *Manager) readConfigByFilePath() (err error) {
	file, err := os.Open(this.config_file_path)
	if err != nil {
		return err
	}
	defer file.Close()

	this.config_content, err = ioutil.ReadAll(file)
	if err != nil {
		return err
	}
	this.config_content = []byte(strings.TrimSpace(string(this.config_content)))
	if len(this.config_content) == 0 {
		return errors.New("config.json中没有内容")
	}
	return nil
}

func (this *Manager) GetHandles() map[uint]ProxyHandler {
	return this.handles
}

func (this *Manager) SetConfigContent(content []byte) {
	this.config_content = content
}

//	将配置文件格式化到配置
//
func (this *Manager) ParseConfig() (err error) {
	var temp map[string][]map[string]interface{}
	err = json.Unmarshal(this.config_content, &temp)
	if err != nil {
		this.log.Info("Unmarshal出错：", err)
		return err
	}

	for k, v := range temp {
		switch k {
		case "tcp":
			for _, val := range v {
				cfg := StreamProxyConfig{}
				utils.FillStruct(val, &cfg)
				this.CreateTcpProxy(&cfg)
			}
		case "ss-client":
			for _, val := range v {
				cfg := SsClientConfig{}
				utils.FillStruct(val, &cfg)
				this.CreateSsClient(&cfg)
			}
		case "ss-server":
			for _, val := range v {
				cfg := SsServerConfig{}
				utils.FillStruct(val, &cfg)
				this.CreateSsServer(&cfg)
			}
		case "http-reproxy":
			for _, val := range v {
				cfg := HttpReproxyConfig{}
				utils.FillStruct(val, &cfg)
				this.CreateHttpReproxy(&cfg)
			}
		}
	}
	return nil
}

//	将配置格式化为json字符串
//
func (this *Manager) formatConfig() (json_bytes []byte, err error) {
	json_bytes, err = json.Marshal(this.handles)
	return
}

//	保存配置到配置文件
//
func (this *Manager) SaveToConfig() error {
	json_bytes, err := this.formatConfig()
	if err != nil {
		this.log.Info("格式化配置{", this.config_file_path, "}出错:", err)
		return err
	}
	file, err := os.Create(this.config_file_path)
	if err != nil {
		this.log.Info("创建配置文件失败: ", err)
		return err
	}
	defer file.Close()

	_, err = file.Write(json_bytes)
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
			handle.Start()
		}
	}
}

//	创建Tcp代理
//
func (this *Manager) CreateTcpProxy(config *StreamProxyConfig) ProxyHandler {
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
	this.handles[config.LocalPort] = handle
	return handle
}

//	创建Http反向代理
//
func (this *Manager) CreateHttpReproxy(config *HttpReproxyConfig) ProxyHandler {
	handle := NewHttpReproxy(config)
	this.handles[config.LocalPort] = handle
	return handle
}
