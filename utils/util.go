package utils

import (
	"fmt"
	"io/ioutil"
	"net"
	"reflect"
	"time"
)

func JoinHostPort(host string, port uint) string {
	return net.JoinHostPort(host, fmt.Sprintf("%d", port))
}

//	map初始化struct
//
func FillStruct(data map[string]interface{}, result interface{}) {
	value := reflect.ValueOf(result).Elem()
	types := reflect.TypeOf(result).Elem()

	for k, v := range data {
		for i := 0; i < value.NumField(); i++ {
			key := types.Field(i).Tag.Get("json")
			if key == k {
				switch reflect.TypeOf(v).Kind().String() {
				case "float64":
					v = int(v.(float64))
				}
				value.Field(i).Set(reflect.ValueOf(v))
			}
		}
	}
}

var publicIp string

func PublicIp() (ip string, err error) {
	//publicIp = "127.0.0.1"
	if len(publicIp) > 0 {
		return publicIp, nil
	}
	conn, err := net.DialTimeout("tcp", "ns1.dnspod.net:6666", 10*time.Second)
	if err != nil {
		return
	}
	defer func() {
		conn.Close()
	}()
	err = conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	if err != nil {
		return
	}
	buff, err := ioutil.ReadAll(conn)
	if err != nil {
		return
	}
	ip = string(buff)
	publicIp = ip
	return
}
