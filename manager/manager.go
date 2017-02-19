/*
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

	"os"
)

type Manager struct {
	handles          Handles
	config_file_path string

	log *log.Logger
}

func New(config_path string) (man *Manager) {
	man = &Manager{
		config_file_path: config_path,
		handles: Handles{
			HttpReproxys: make(map[int]*HttpReproxyHandle),
			TcpProxys:    make(map[int]*TcpProxyHandle),
			SsServers:    make(map[int]*SsServerHandle),
		},
		log: log.NewLogger("Manager"),
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
			for _, val := range v {
				cfg := TcpProxyConfig{}
				utils.FillStruct(val, &cfg)
				this.handles.TcpProxys[cfg.LocalPort] = NewTcpProxy(&cfg)
			}
		case "ss-client":
			for _, val := range v {
				cfg := SsClientConfig{}
				utils.FillStruct(val, &cfg)
				this.handles.SsClient = NewSsClient(&cfg)
			}
		case "ss-server":
			for _, val := range v {
				cfg := SsServerConfig{}
				utils.FillStruct(val, &cfg)
				this.handles.SsServers[cfg.LocalPort] = NewSsServer(&cfg)
			}
		case "http-reproxy":
			for _, val := range v {
				cfg := HttpReproxyConfig{}
				utils.FillStruct(val, &cfg)
				this.handles.HttpReproxys[cfg.LocalPort] = NewHttpReproxy(&cfg)
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
		this.log.Info("保存配置到配置文件{", this.config_file_path, "}出错:", err)
		return err
	}
	file, err := os.Create(this.config_file_path)
	_, err = file.Write(json_bytes)
	if err != nil {
		this.log.Info("写入配置到配置文件出错：", err)
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
	this.StopAll()

	if this.handles.SsClient != nil {
		this.log.Info("Run ss client...")
		this.handles.SsClient.Proxy.Start()
	}
	if this.handles.SsServers != nil {
		for _, handle := range this.handles.SsServers {
			this.log.Info("Run ss server...")
			handle.Proxy.Start()
		}
	}
	if this.handles.TcpProxys != nil {
		for _, handle := range this.handles.TcpProxys {
			this.log.Info("Run tcp proxy...")
			handle.Proxy.Start()
		}
	}
	if this.handles.HttpReproxys != nil {
		for _, handle := range this.handles.HttpReproxys {
			this.log.Info("Run http reproxy...")
			handle.Proxy.Start()
		}
	}
}
