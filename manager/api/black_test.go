package api

import (
	"bytes"
	"net/http"
	"strings"
	"testing"
	"time"
	
	"bitproxy/services"
)

func TestCreateBlack(t *testing.T) {
	go Start("test", 8888)

	time.Sleep(time.Second * 1)

	buf := strings.NewReader(`
		{
			"ips":["222.222.222.222", "111.1.1.1"]
		}
		`)

	r, err := http.Post("http://localhost:8888/Black/", "application/json", buf)

	var b bytes.Buffer
	n, err := b.ReadFrom(r.Body)
	if err != nil {
		t.Log("Read error ", err)
		return
	}
	//t.Error(err)
	//t.Error(string(b.Bytes()))
	t.Log("read buf count ", n)
	t.Log("buf ", string(b.Bytes()))

	if blacklist.Wall.IsBlack("222.222.222.222") == false {
		t.Fatal("添加到屏蔽墙失败！")
	}
}
