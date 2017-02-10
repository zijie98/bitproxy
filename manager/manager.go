/*
	m = proxy.NewManager("xx.config")
	m.RunAll()

	m.StartTcpProxy(config)
	m.StartHttpReproxy(config)
	m.StartSsClient(config)
	m.StartSsServer(config)

	m.StopSsServer(port)
	m.StopSsClient()
	m.StopTcpProxy(port)
	m.StopHttpReproxy(port)
*/
package manager

import (
	"encoding/json"
	"io/ioutil"
	"rkproxy/log"
	"rkproxy/proxy"
	"rkproxy/proxy/ss"
	"rkproxy/utils"

	"os"
)

type Handles struct {
	HttpReproxys map[int]*HttpReproxyHandle `json:"http-reproxy"`
	TcpProxys    map[int]*TcpProxyHandle    `json:"tcp"`
	SsClient     *SsClientHandle            `json:"ss-client"`
	SsServers    map[int]*SsServerHandle    `json:"ss-server"`
}

type Manager struct {
	handles          Handles
	config_file_path string
}

func New(config_path string) (man *Manager) {
	man = &Manager{
		config_file_path: config_path,
		handles: Handles{
			HttpReproxys: make(map[int]*HttpReproxyHandle),
			TcpProxys:    make(map[int]*TcpProxyHandle),
			SsServers:    make(map[int]*SsServerHandle),
		},
	}
	return
}

//	将配置文件格式化到配置
//
func (this *Manager) ParseConfig() (err error) {
	file, err := os.Open(this.config_file_path)
	if err != nil {
		return err
	}
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	var temp_handles map[string][]map[string]interface{}
	err = json.Unmarshal(data, &temp_handles)
	if err != nil {
		return err
	}

	for k, v := range temp_handles {
		switch k {
		case "tcp":
			//fmt.Println(reflect.TypeOf(v).String())
			for _, val := range v {
				cfg := TcpProxyConfig{}
				utils.FillStruct(val, &cfg)
			}
		case "ss-client":
			for _, val := range v {
				cfg := SsClientConfig{}
				utils.FillStruct(val, &cfg)
			}
		case "ss-server":
			for _, val := range v {
				cfg := SsServerConfig{}
				utils.FillStruct(val, &cfg)
			}
		case "http-reproxy":
			for _, val := range v {
				cfg := HttpReproxyConfig{}
				utils.FillStruct(val, &cfg)
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
		log.Info("保存配置到配置文件{", this.config_file_path, "}出错:", err)
		return err
	}
	file, err := os.Create(this.config_file_path)
	_, err = file.Write(json_bytes)
	if err != nil {
		log.Info("写入配置到配置文件出错：", err)
	}
	file.Close()

	return err
}

//	关闭所有的代理服务
//
func (this *Manager) StopAll() {
	if this.handles.SsClient != nil {
		this.handles.SsClient.Proxy.Stop()
	}

	if this.handles.SsServers != nil {
		for _, handle := range this.handles.SsServers {
			handle.Proxy.Stop()
		}
	}
	if this.handles.TcpProxys != nil {
		for _, handle := range this.handles.TcpProxys {
			handle.Proxy.Stop()
		}
	}
	if this.handles.HttpReproxys != nil {
		for _, handle := range this.handles.HttpReproxys {
			handle.Proxy.Stop()
		}
	}
}

//	运行所有的代理服务
//
func (this *Manager) RunAll() {
	if this.handles.SsClient != nil {
		log.Info("MANAGER: Run ss client...")
		this.handles.SsClient.Proxy.Start()
	}
	if this.handles.SsServers != nil {
		for _, handle := range this.handles.SsServers {
			log.Info("MANAGER: Run ss server...")
			handle.Proxy.Start()
		}
	}
	if this.handles.TcpProxys != nil {
		for _, handle := range this.handles.TcpProxys {
			log.Info("MANAGER: Run tcp proxy...")
			handle.Proxy.Start()
		}
	}
	if this.handles.HttpReproxys != nil {
		for _, handle := range this.handles.HttpReproxys {
			log.Info("MANAGER: Run http reproxy...")
			handle.Proxy.Start()
		}
	}
}

func (this *Manager) StartSsClient(config *SsClientConfig) (*SsClientHandle, error) {
	proxy := ss.NewClient(
		config.LocalPort,
		utils.JoinHostPort(config.RemoteHost, config.RemotePort),
		config.Password,
		config.Crypt,
	)
	this.handles.SsClient = &SsClientHandle{
		Config: config,
		Proxy:  proxy,
	}
	err := proxy.Start()
	return this.handles.SsClient, err
}

func (this *Manager) StartSsServer(config *SsServerConfig) (*SsServerHandle, error) {
	proxy := ss.NewServer(
		config.LocalPort,
		config.Password,
		config.Crypt,
	)
	handle := &SsServerHandle{
		Config: config,
		Proxy:  proxy,
	}
	err := proxy.Start()
	if err == nil {
		this.handles.SsServers[config.LocalPort] = handle
	}
	return handle, err
}

func (this *Manager) StartTcpProxy(config *TcpProxyConfig) (*TcpProxyHandle, error) {
	proxy := proxy.NewTcpProxy(
		config.LocalPort,
		config.RemoteHost,
		config.RemotePort,
	)
	handle := &TcpProxyHandle{
		Config: config,
		Proxy:  proxy,
	}
	err := proxy.Start()
	if err == nil {
		this.handles.TcpProxys[config.LocalPort] = handle
	}
	return handle, err
}

func (this *Manager) StartHttpReproxy(config *HttpReproxyConfig) (*HttpReproxyHandle, error) {
	proxy := proxy.NewHttpReproxy(
		config.LocalPort,
		config.RemoteHost,
		config.RemotePort,
	)
	handle := &HttpReproxyHandle{
		Config: config,
		Proxy:  proxy,
	}
	err := proxy.Start()
	if err == nil {
		this.handles.HttpReproxys[config.LocalPort] = handle
	}
	return handle, err
}
