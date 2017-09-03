package main

import (
	"fmt"
	"net"
	"time"

	"bufio"
	"bytes"
	"bitproxy/proxy"
)

func main() {
	pxy := proxy.NewHttpReproxy(1993, "douban.com", 80, "rkecloud.com")
	go pxy.Start()

	time.Sleep(1 * time.Second)
	fmt.Println("----")
	conn, err := net.Dial("tcp", "localhost:1993")
	if err != nil {
		fmt.Println("... dial err ", err)
		return
	}
	normal_test(conn)
}

func normal_test(conn net.Conn) {

	http_get_request_string := `GET / HTTP/1.1
Host: douban.com
Connection: keep-alive
User-Agent: HHH

`
	n, err := conn.Write([]byte(http_get_request_string))
	fmt.Println("request len ", n)
	if err != nil {
		fmt.Println("request baidu.com err  ", err)
		return
	}

	reader := bufio.NewReader(conn)
	for {
		line, _, _ := reader.ReadLine()
		if bytes.Compare(line, []byte("\r\n\r\n")) == 0 {
			break
		}
		fmt.Println(string(line))
	}
}
