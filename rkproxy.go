package main

import (
	"rkproxy/log"
	"rkproxy/manager"
)

const config_path = "config.json"

var man *manager.Manager = manager.New(config_path)

func main() {
	err := man.ParseConfig()
	if err != nil {
		log.Info(err.Error())
		return
	}

	//man.RunAll()

	select {}
}
