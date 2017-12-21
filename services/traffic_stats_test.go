package services

import (
	"testing"
	"time"
)

func TestAddTrafficStats(t *testing.T) {

	InitRedis("127.0.0.1", 6379)
	StartStats()

	DeleteTrafficStats(10001)

	time.Sleep(1 * time.Second)

	AddTrafficStats(10001, 1000)
	AddTrafficStats(10001, 1000)

	time.Sleep(1 * time.Second)

	n, err := GetTraffic(10001)
	if err != nil {
		t.Error(err)
	}
	if n <= 0 {
		t.Error("无法读取到流量统计 ", n)
	}
	if n == 2000 {
		t.Log("good")
	}
}
