/*
	By moli molisoft@qq.com
	Blog huoxr.com

*/
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"syscall"

	"github.com/molisoft/bitproxy/manager"
	"github.com/molisoft/bitproxy/manager/api"
	"github.com/molisoft/bitproxy/services"
	"github.com/molisoft/bitproxy/utils"
	"io"
)

const ETC_PATH = "/etc"
const CONFIG_FILENAME = "config.json"
const PID_FILENAME = "bitproxy.pid"

var AppPath string
var AppPidPath string
var ConfigPath string
var log *utils.Logger

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

func initEnv() {
	path, _ := exec.LookPath(os.Args[0])
	AppPath = filepath.Dir(path)
}

func initLog() {
	utils.LogPath = AppPath
	log = utils.NewLogger("Main")
}

func initPid() {
	file, err := os.Open(manager.Man.PidPath)
	if err == nil {
		pid, err := bufio.NewReader(file).ReadString(' ')
		if err == io.EOF {
			id, err := strconv.Atoi(pid)
			if err == nil {
				fmt.Println("kill old process ", id)
				syscall.Kill(id, syscall.SIGINT)
			}
		}
		file.Close()
		os.Remove(manager.Man.PidPath)
	}

	file, err = os.OpenFile(manager.Man.PidPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("写pid文件错误：", err)
		return
	}
	defer file.Close()
	file.WriteString(fmt.Sprintf("%d", os.Getpid()))
}

func initApi() error {
	if manager.Man.Config.Api != nil {
		config := manager.Man.Config.Api
		return api.Start(config.Password, config.Port)
	}
	return nil
}

func initRedis() {
	if manager.Man.Config.Redis != nil {
		config := manager.Man.Config.Redis
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

func initFlag() {
	pidPath := filepath.Join(AppPath, PID_FILENAME)
	configPath := filepath.Join(AppPath, CONFIG_FILENAME)

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		configPath = filepath.Join(ETC_PATH, CONFIG_FILENAME)
	}
	if _, err := os.Stat(pidPath); os.IsNotExist(err) {
		pidPath = filepath.Join(ETC_PATH, PID_FILENAME)
	}

	flag.StringVar(&ConfigPath, "c", configPath, "配置文件")
	flag.StringVar(&AppPidPath, "p", pidPath, "进程id路径")
	flag.Parse()
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	initEnv()
	initLog()
	initFlag()

	manager.New(AppPath, AppPidPath, ConfigPath)

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
