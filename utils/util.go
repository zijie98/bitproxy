package utils

import (
	"fmt"
	"net"
	"reflect"
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
