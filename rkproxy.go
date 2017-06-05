/*
	By moli molisoft@qq.com
	Blog huoxr.com

*/
package main

import (
	"fmt"
	"os"
	"os/signal"
	logger "rkproxy/log"
	"rkproxy/manager"
	"runtime"
)

const config_path = "config.json"

var man *manager.Manager = manager.New(config_path)
var log *logger.Logger = logger.NewLogger("Main")

func listen_signal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			fmt.Printf("ctrl+c(%v)\n", sig)
			man.StopAll()
			os.Exit(0)
		}
	}()
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	listen_signal()

	err := man.ParseConfig()
	if err != nil {
		log.Info(err.Error())
		return
	}

	man.RunAll()

	select {}
}
