package main

import (
	"fmt"
	"rkproxy/manager"
	//"reflect"
)

func main() {
	config := []byte(`
		{
			"tcp": [
				{
					"local_port": 8080,
					"remote_host": "baidu.com",
					"remote_port": 80
				}
			],
			"ss-client":[
				{
					"local_port": 8081,
					"remote_host": "127.0.0.1",
					"remote_port": 8082,
					"password": "123",
					"crypt": "XOR"
				}
			],
			"ss-server": [
				{
					"server_port": 8082,
					"password": "123",
					"crypt": "XOR"
				}
			]
		}
	`)
	man := manager.NewByConfigContent(config)
	err := man.ParseConfig()
	if err != nil {
		fmt.Println("ParseConfig .. ", err)
		return
	}
	handles := man.GetHandles()
	port := handles[8080].ListeningPort()
	fmt.Println(port == 8080)
}
