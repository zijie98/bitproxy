/**
屏蔽IP

- 实现思路：保存IP的最近x次访问的时间，最新与最后一次对比，如果时间间隔少于n秒，则视为恶意IP(一秒钟超过10个请求则屏蔽)
- 目的：屏蔽一定程度的CC攻击
*/
package blacklist

import (
	"container/list"
	"time"
)

const (
	MAX_LIMIT   = 15
	MAX_BETWEEN = 1 // 最后一次请求和最近的请求的间隔多少秒
)

type BlackContainer struct {
	blacks   map[string]bool
	runtimes map[string]*list.List

	max_limit   int
	max_between int64
}

var BlackWall *BlackContainer

func init() {
	if BlackWall == nil {
		BlackWall = &BlackContainer{
			max_limit:   MAX_LIMIT,
			max_between: MAX_BETWEEN,
			runtimes:    make(map[string]*list.List),
			blacks:      make(map[string]bool),
		}
	}
}

func (r *BlackContainer) AddIp(ip string) {
	if r.IsBlack(ip) {
		return
	}
	if r.runtimes[ip] == nil {
		r.runtimes[ip] = list.New()
	}
	r.runtimes[ip].PushFront(ip)

	r.check(&ip)
}

func (r *BlackContainer) Black(ip string) {
	r.blacks[ip] = true
}

func (r *BlackContainer) IsBlack(ip string) bool {
	return r.blacks[ip]
}

func (r *BlackContainer) check(ip *string) {
	if r.runtimes[*ip].Len() < r.max_limit {
		return
	}
	back := r.runtimes[*ip].Back().Value.(time.Time)
	front := r.runtimes[*ip].Front().Value.(time.Time)
	if (front.Unix() - back.Unix()) <= r.max_between { // 最新和最后一次请求时间小于等于max_between则为非法请求
		r.Black(*ip)
	}
	r.runtimes[*ip].Init() // 超过10个请求清空
}

func (r *BlackContainer) Remove(ip string) {
	delete(r.blacks, ip)
}
