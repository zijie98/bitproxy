package ss

import (
	"container/list"
	"net"
	"testing"
)

func TestClientLimit(t *testing.T) {
	limitCount := 2
	clientLimit := &ClientLimit{
		limitCount: limitCount,
		clients:    list.New(),
	}
	conn, err := net.Dial("tcp", "baidu.com:80")
	if err != nil {
		t.Error(err)
		return
	}
	conn2, _ := net.Dial("tcp", "163.com:80")
	conn3, _ := net.Dial("tcp", "qq.com:80")

	clientLimit.Add(conn)
	clientLimit.Add(conn2)
	clientLimit.Add(conn3)
	if clientLimit.isExist(getIp(conn.RemoteAddr().String())) {
		t.Error("conn应被删除")
	}
}

func getIp(str string) string {
	ip, _, _ := net.SplitHostPort(str)
	return ip
}
