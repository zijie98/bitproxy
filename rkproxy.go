/*
	By moli molisoft@qq.com
	Blog huoxr.com

*/
package main

import (
	logger "rkproxy/log"
	"rkproxy/manager"
	"runtime"
)

const config_path = "config.json"

var man *manager.Manager = manager.New(config_path)
var log *logger.Logger = logger.NewLogger("Main")

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	err := man.ParseConfig()
	if err != nil {
		log.Info(err.Error())
		return
	}

	man.RunAll()

	select {}
}
