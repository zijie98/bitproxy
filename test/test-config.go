package main

import (
	"encoding/json"
	"fmt"
	"rkproxy/manager"

	//"reflect"
	"rkproxy/utils"
)

func main() {
	config := []byte(`
		{
			"tcp": [
				{
					"local_port": 80,
					"remote_host": "127.0.01"
				}
			],
			"ss-client":
				{
					"local_port": 80,
					"remote_host": "127.0.0.1",
					"password": "123"
				},
			"ss-server":
			[]
		}
	`)
	handles := manager.Handles{
		HttpReproxys: make(map[int]*manager.HttpReproxyHandle),
		TcpProxys:    make(map[int]*manager.TcpProxyHandle),
		SsServers:    make(map[int]*manager.SsServerHandle),
	}

	var hash_config map[string][]map[string]interface{}

	err := json.Unmarshal(config, &hash_config)
	fmt.Println(err)

	fmt.Println(hash_config)

	for k, v := range hash_config {
		switch k {
		case "tcp":
			fmt.Println(v)
			fmt.Println("-=====--")
			for _, item := range v {

				//v[i]["local_port"]
				fmt.Println(item)
				fmt.Println("========")
				tcp_cfg := manager.TcpProxyConfig{}
				utils.FillStruct(item, &tcp_cfg)
				fmt.Println("---------")
				fmt.Println(tcp_cfg)
			}
		}
	}

	fmt.Println(handles.TcpProxys)

	if handles.TcpProxys == nil {
		fmt.Println("is nill")
	}
}
