/*
	By moli molisoft@qq.com
	Blog huoxr.com

*/
package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	logger "rkproxy/log"
	"rkproxy/manager"
	"runtime"
)

var config_path = "config.json"

var man *manager.Manager
var log *logger.Logger = logger.NewLogger("Main")

func listen_signal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		select {
		case sig := <-c:
			fmt.Printf("ctrl+c (%v)\n", sig)
			if man != nil {
				man.StopAll()
			}
			os.Exit(0)
		}
	}()
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	listen_signal()

	flag.StringVar(&config_path, "c", "config.json", "配置文件")
	flag.Parse()

	man = manager.New(config_path)

	err := man.ParseConfig()
	if err != nil {
		log.Info(err.Error())
		return
	}

	man.RunAll()

	select {}
}
