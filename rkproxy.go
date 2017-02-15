package main

import (
	logger "rkproxy/log"
	"rkproxy/manager"
)

const config_path = "config.json"

var man *manager.Manager = manager.New(config_path)
var log *logger.Logger = logger.NewLogger("Main")

func main() {
	err := man.ParseConfig(false)
	if err != nil {
		log.Info(err.Error())
		return
	}

	//man.RunAll()

	select {}
}
