/**
在服务器端限制客户端的数量
*/

package ss

import (
	"container/list"
	"net"
	"time"
)

type Client struct {
	conns map[net.Conn]int8 // 连接者的连接
	ip    string            // 连接者的ip
	at    time.Time         // 时间
}

type ClientLimit struct {
	clients    *list.List // list适合做堆栈
	limitCount int
}

func NewClientLimit(n int) *ClientLimit {
	return &ClientLimit{
		clients:    list.New(),
		limitCount: n,
	}
}

func (this *ClientLimit) findElementByIp(ip string) *list.Element {
	for e := this.clients.Front(); e != nil; e = e.Next() {
		if e.Value.(*Client).ip == ip {
			return e
		}
	}
	return nil
}

// 根据ip查找到ClientLimit对象
func (this *ClientLimit) findByIp(ip string) *Client {
	e := this.findElementByIp(ip)
	if e != nil {
		return e.Value.(*Client)
	}
	return nil
}

// 检查某个ip是否存在
func (this *ClientLimit) isExist(ip string) bool {
	return this.findByIp(ip) != nil
}

// 释放ip下的连接
func (this *ClientLimit) release() bool {
	if this.clients.Len() == 0 {
		return false
	}
	ele := this.clients.Back()
	client := ele.Value.(*Client)
	for conn, _ := range client.conns {
		conn.Close()
	}
	client.conns = nil
	this.clients.Remove(ele)
	return true
}

func (this *ClientLimit) RemoveConn(client net.Conn) {
	c := this.findByIp(getIp(client))
	if c != nil {
		delete(c.conns, client)
	}
}

func (this *ClientLimit) isExceed() bool {
	return this.clients.Len() > this.limitCount
}

func (this *ClientLimit) Add(client net.Conn) {
	// ip如果已存在则记录链接，同时跳过
	ip := getIp(client)
	c := this.findByIp(ip)
	if c != nil {
		c.conns[client] = 0
		return
	}
	// 新ip添加到列表
	c = &Client{
		ip:    ip,
		at:    time.Now(),
		conns: make(map[net.Conn]int8, 10),
	}
	c.conns[client] = 0
	this.clients.PushFront(c)

	// 限制列表超过限制数，则释放最后一个连接的ip
	if this.isExceed() {
		this.release()
	}

}

func getIp(conn net.Conn) string {
	ip := conn.RemoteAddr().String()
	ip, _, _ = net.SplitHostPort(ip)
	return ip
}
