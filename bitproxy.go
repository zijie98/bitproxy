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

	"github.com/molisoft/bitproxy/manager"
	"github.com/molisoft/bitproxy/manager/api"
	"github.com/molisoft/bitproxy/services"
	"github.com/molisoft/bitproxy/utils"
)

var log *utils.Logger = utils.NewLogger("Main")

func listenSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		select {
		case sig := <-c:
			fmt.Printf("ctrl+c (%v)\n", sig)
			if manager.Man != nil {
				manager.Man.StopAll()
			}
			os.Remove(manager.Man.PidPath)
			os.Exit(0)
		}
	}()
}

func initFlag() {
	flag.StringVar(&manager.Man.ConfigPath, "c", manager.Man.ConfigPath, "配置文件")
	flag.StringVar(&manager.Man.PidPath, "p", manager.Man.PidPath, "进程id路径")
	flag.Parse()
}

func initPid() {
	if _, err := os.Stat(manager.Man.PidPath); os.IsExist(err) {
		os.Remove(manager.Man.PidPath)
	}
	file, err := os.OpenFile(manager.Man.PidPath, os.O_CREATE|os.O_WRONLY, 0644)
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
