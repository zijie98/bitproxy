package log

import (
	"fmt"
	"time"
)

type Logger struct {
	prefix string
}

func NewLogger(prefix string) *Logger {
	return &Logger{
		prefix: prefix,
	}
}

func (this *Logger) Info(msg ...interface{}) {
	fmt.Print(time.Now().Format("2006-01-02 15:04:05"), " - ")
	fmt.Print(this.prefix)
	fmt.Print("-")
	fmt.Println(msg...)
}
