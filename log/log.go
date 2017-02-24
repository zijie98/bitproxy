package log

import "fmt"

type Logger struct {
	prefix string
}

func NewLogger(prefix string) *Logger {
	return &Logger{
		prefix: prefix,
	}
}

func (this *Logger) Info(msg ...interface{}) {
	fmt.Print(this.prefix)
	fmt.Print(" - ")
	fmt.Println(msg...)
}
