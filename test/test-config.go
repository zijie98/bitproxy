package main

import (
	"fmt"
	"rkproxy/manager"
	//"reflect"
	"os"
)

func main() {
	filepath := "config.json"
	file, err := os.OpenFile(filepath, os.O_RDWR, 0755)
	if err != nil {
		fmt.Println("open test.json fail:", err)
		return
	}
	defer file.Close()

	man := manager.New(filepath)
	err = man.ParseConfig()
	if err != nil {
		fmt.Println("ParseConfig .. ", err)
		return
	}
	handles := man.GetHandles()
	port := handles[1081].ListeningPort()
	fmt.Println(port == 1081)
}
