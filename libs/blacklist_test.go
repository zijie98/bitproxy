package blacklist

import (
	"testing"
	"time"
)

func TestBlackContainer_Black(t *testing.T) {
	var i int = 0

	for i != 100 {
		Filter <- RequestAt{
			Ip: "127.0.0.1",
			At: time.Now(),
		}
		i++
	}

	//t.Fatalf("%#v", Wall.blacks)
	//t.Fatal("return ", Wall.blacks["127.0.0.1"])
	time.Sleep(1 * time.Second)
	if Wall.IsBlack("127.0.0.1") {
		t.Log("is black .. 127.0.0.1")
	}

}
