package main

import (
	//"rkproxy/proxy"
	"rkproxy/proxy/ss"
)

func main() {

	go ss.StartSSClient(4000, "127.0.0.1:4001", "123", "xor")

	go ss.StartSSServer(4001, "123", "xor")

	select {}
}
