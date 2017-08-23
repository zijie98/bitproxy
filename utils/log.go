package utils

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
	at := time.Now().Format("2006-01-02 15:04:05")
	line := fmt.Sprintf("%s - %s -", at, this.prefix)
	fmt.Println(line, msg)
}
