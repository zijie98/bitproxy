package blacklist

import (
	"testing"
	"time"
)

func TestBlackContainer_Black(t *testing.T) {
	var i int = 0

	for i != 100 {
		BlackFilter <- RequestAt{
			Ip: "127.0.0.1",
			At: time.Now(),
		}
		i++
	}

	//t.Fatalf("%#v", BlackAts.blacks)
	//t.Fatal("return ", BlackAts.blacks["127.0.0.1"])
	time.Sleep(1 * time.Second)
	if BlackAts.IsBlack("127.0.0.1") {
		t.Log("is black .. 127.0.0.1")
	}

}
