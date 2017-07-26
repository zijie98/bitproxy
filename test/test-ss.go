package main

import (
	"fmt"
	"golang.org/x/net/proxy"
	"net/http"
	"rkproxy/proxy/ss"
	"time"
)

func main() {
	net := ss.CryptXor
	server := ss.NewServer(ss.KCP_PROTOCOL, 1990, "hellopwd", net, 0)
	go server.Start()

	client := ss.NewClient(ss.TCP_PROTOCOL, 1991, "localhost:1990", ss.KCP_PROTOCOL, "hellopwd", net)
	go client.Start()

	time.Sleep(1 * time.Second)
	if httpRequestTest() {
		fmt.Println("---------测试成功！-----------", net)
	} else {
		fmt.Println("---------测试失败！-----------", net)
	}
}

func httpRequestTest() bool {

	dialer, err := proxy.SOCKS5("tcp", "localhost:1991", nil, proxy.Direct)
	if err != nil {
		return false
	}

	// http
	httpTransport := &http.Transport{Dial: dialer.Dial}
	httpclient := http.Client{Transport: httpTransport}

	req, err := http.NewRequest("GET", "http://www.baidu.com", nil)

	if err != nil {
		fmt.Println("httpRequestTest", err)
		return false
	}
	resp, err := httpclient.Do(req)
	fmt.Println("resp ..is ok ..")
	if err != nil {
		fmt.Println("httpRequestTest", err)
		return false
	}

	if resp.StatusCode == 200 {
		fmt.Println("httpRequestTest test is ok")
		return true
	}
	return false
}
