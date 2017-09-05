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
	"runtime"

	"bitproxy/manager"
	"bitproxy/manager/api"
	"bitproxy/services"
	"bitproxy/utils"
)

var man *manager.Manager
var log *utils.Logger = utils.NewLogger("Main")

func listenSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		select {
		case sig := <-c:
			fmt.Printf("ctrl+c (%v)\n", sig)
			if man != nil {
				man.StopAll()
			}
			os.Remove(man.PidPath)
			os.Exit(0)
		}
	}()
}

func initFlag() {
	flag.StringVar(&man.ConfigPath, "c", man.ConfigPath, "配置文件")
	flag.StringVar(&man.PidPath, "p", man.PidPath, "进程id路径")
	flag.Parse()
}

func initPid() {
	if _, err := os.Stat(man.PidPath); os.IsExist(err) {
		os.Remove(man.PidPath)
	}
	file, err := os.OpenFile(man.PidPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("写pid文件错误：", err)
		return
	}
	defer file.Close()
	file.WriteString(fmt.Sprintf("%d", os.Getpid()))
}

func initApi() error {
	if manager.Man.Config().Api != nil {
		config := manager.Man.Config().Api
		return api.Start(config.Password, config.Port)
	}
	return nil
}

func initRedis() {
	if manager.Man.Config().Redis != nil {
		config := manager.Man.Config().Redis
		services.InitRedis(config.Host, config.Port)
	}
}

func initStats() {
	services.StartStats()
}

func initBlackList() {
	services.StartBlackList()
}

func initPublicIp() {
	_, err := utils.PublicIp()
	if err != nil {
		fmt.Println("初始化公网IP出错（用于ftp代理）")
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	manager.New()

	initFlag()

	err := manager.Man.ParseConfig()
	if err != nil {
		log.Info(err.Error())
		return
	}

	initPid()
	listenSignal()

	initRedis()
	initStats()
	initBlackList()
	initPublicIp()

	manager.Man.RunAll()

	initApi()
	select {}
}
