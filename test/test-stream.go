package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"time"

	proxy "github.com/molisoft/bitproxy/proxy"
)

func main() {
	pxy := proxy.NewStreamProxy(proxy.TCP_PROTOCOL, 1992, "www.baidu.com", 80, 0, false)
	go pxy.Start()
	time.Sleep(1 * time.Second)
	fmt.Println("----")
	conn, err := net.Dial("tcp", "localhost:1992")
	if err != nil {
		fmt.Println("... dial err ", err)
		return
	}
	http_get_request_string := `GET / HTTP/1.1
Host: www.baidu.com
Connection: keep-alive
User-Agent: HHH

`
	n, err := conn.Write([]byte(http_get_request_string))
	fmt.Println("request len ", n)
	if err != nil {
		fmt.Println("request baidu.com err  ", err)
		return
	}

	buff := make([]byte, 512)
	n, err = io.ReadFull(conn, buff)

	if n == 512 {
		//pxy.Stop()
		fmt.Println("stream test is ok")
	} else {
		fmt.Println("stream test is fail")
		os.Exit(1)
	}
}
