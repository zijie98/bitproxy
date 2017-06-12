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

var (
	config_path = "config.json"
	pid_path    = "rkproxy.pid"
)

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

func init_flag() {
	flag.StringVar(&config_path, "c", "config.json", "配置文件")
	flag.StringVar(&pid_path, "p", "rkproxy.pid", "进程id路径")
	flag.Parse()
}

func init_pid() {
	file, err := os.OpenFile(pid_path, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("写pid文件错误：", err)
		return
	}
	defer file.Close()
	file.Write([]byte(fmt.Sprintf("%d", os.Getpid())))
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	init_flag()
	init_pid()
	listen_signal()

	man = manager.New(config_path)

	err := man.ParseConfig()
	if err != nil {
		log.Info(err.Error())
		return
	}

	man.RunAll()

	select {}
}
