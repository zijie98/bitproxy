package log

import "fmt"

func Info(prefix string, msg ...interface{}) {
	fmt.Print(prefix)
	fmt.Print(" ")
	fmt.Println(msg...)
}
