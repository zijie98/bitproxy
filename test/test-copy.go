package main

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"rkproxy/utils"
)

func main() {
	fmt.Println("Start..")
	limit := &utils.Limiter{Rate: 0} // 20KB/s
	s := ""
	for i := 0; i < 20480; i++ {
		s += "helloworld"
	}
	fmt.Printf("begin .. %v\n", time.Now())
	fmt.Printf("size %dbyte %dKB\n", len(s), len(s)/1024)
	reader := strings.NewReader(s)
	var writer = &bytes.Buffer{}

	utils.Copy(writer, reader, limit)

	fmt.Printf("end .. %v\n", time.Now())
}
