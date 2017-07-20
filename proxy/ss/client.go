/*
	Author moli
	2016-12-03 00：27
	ss 客户端, 使用kcp（udp）协议，加速网络环境不好的情况下提高网络传输质量。如果网络质量好，直接TCP ……
	ss client <-> ss server

	ss.StartSSClient(local_port)
*/

package ss

import (
	"fmt"
	"io"
	"net"

	"github.com/kataras/go-errors"
	"github.com/xtaci/kcp-go"
	//"github.com/xtaci/smux"

	"rkproxy/utils"
)

type SSClient struct {
	crypt       string
	pwd         string
	local_port  uint
	local_net   NetProtocol // tcp/udp， 浏览器 -> ss客户端 之间的通信方式
	channel_net NetProtocol // tcp/udp/kcp， ss客户端 -> ss服务器 之间的通信方式
	server_addr string      // 8.8.8.8:1990
	ln          net.Listener
	log         *utils.Logger
}

func (this *SSClient) handle(client io.ReadWriter) {
	buf := make([]byte, 32) // 32 byte
	_, err := client.Read(buf)
	if err != nil {
		this.log.Info("Client online err ", err)
		return
	}
	//clog("1 - ", buf)
	if buf[0] != SOCKS5_VERSION {
		this.log.Info("Only suppert socks5")
		return
	}

	buf = buf[:0] // reset
	buf = append(buf, SOCKS5_VERSION, 0)
	if _, err = client.Write(buf); err != nil {
		this.log.Info("Send msg to client err ", err)
		return
	}

	raw_addr, err := this.getRequestRemoteAddr(client)
	if err != nil {
		this.log.Info("RemoteAddrHandle err ", err, raw_addr)
		return
	}

	buf = buf[:0] // reset
	buf = append(buf, SOCKS5_VERSION, 0, 0, 0x1, 0, 0, 0, 0, 0, 0)
	if _, err := client.Write(buf); err != nil {
		this.log.Info("Send established to client err ", err)
		return
	}

	// connect to server
	server_conn, err := this.getServerConn()
	if err != nil {
		this.log.Info("Connect to server fail ", err)
		return
	}

	server, err := NewCryptConn(server_conn, this.pwd, this.crypt)
	if err != nil {
		this.log.Info("NewCryptConn err ", err)
		return
	}

	// ss协议中，将把浏览器的请求发给服务器
	server.Write(raw_addr)

	go utils.Copy(client, server, nil, nil)
	utils.Copy(server, client, nil, nil)
}

func (this *SSClient) getServerConn() (net.Conn, error) {
	if this.channel_net == UDP_PROTOCOL {
		return net.Dial("udp", this.server_addr)

	} else if this.channel_net == TCP_PROTOCOL {
		return net.Dial("tcp", this.server_addr)

	} else if this.channel_net == KCP_PROTOCOL {
		conn, err := kcp.DialWithOptions(this.server_addr, nil, 10, 3)
		if err != nil {
			this.log.Info("SSCLIENT: kcp.DialWithOptions error", err)
			return nil, errors.New("SSCLIENT: kcp.DialWithOptions error " + err.Error())
		}
		conn.SetStreamMode(true)
		conn.SetNoDelay(1, 20, 1, 0)
		conn.SetMtu(1350)
		conn.SetWindowSize(1024, 1024)
		conn.SetACKNoDelay(true)
		conn.SetKeepAlive(10)
		return conn, nil
	} else {
		return nil, errors.New("Not found protocol: " + string(this.channel_net) + "?")
	}
}

func (this *SSClient) getRequestRemoteAddr(client io.ReadWriter) (addr []byte, err error) {

	addr = make([]byte, 260) // 260 byte

	n, err := client.Read(addr)
	if err != nil || n < CLIENT_RAW_TYPE {
		this.log.Info("Read error ", err)
		return nil, err
	}

	reqLen := -1
	switch addr[CLIENT_RAW_TYPE] {
	case TYPE_IPV4:
		reqLen = CLIENT_LEN_TYPE_IPV4
	case TYPE_IPV6:
		reqLen = CLIENT_LEN_TYPE_IPV6
	case TYPE_DOMAIN:
		reqLen = int(addr[CLIENT_RAW_ADDR]) + CLIENT_LEN_TYPE_DOMAIN_BASE
	default:
		this.log.Info("Raw addr err", &client)
		return nil, errors.New("Raw addr err")
	}

	if n < reqLen {
		if _, err := io.ReadFull(client, addr[n:reqLen]); err != nil {
			this.log.Info("ReadFull err..", err)
			return nil, err
		}
	}
	addr = addr[CLIENT_RAW_TYPE:reqLen]
	return
}

func (this *SSClient) initListen() error {
	if this.local_net == KCP_PROTOCOL {
		return errors.New("浏览器（或其他软件）连接到本客户端是不支持kcp协议的")
	}
	var err error
	this.ln, err = net.Listen(string(this.local_net), fmt.Sprintf(":%d", this.local_port))
	if err != nil {
		this.log.Info("Listen err ", err)
	}
	return nil
}

func (this *SSClient) Start() error {
	this.initListen()

	this.log.Info("Listen port", this.local_port)
	for {
		conn, err := this.ln.Accept()
		if err != nil {
			this.log.Info("Accept err ", err)
			continue
		}
		this.log.Info("Accept ip ", conn.RemoteAddr(), "client ", &conn)
		go this.handle(conn)
	}
	return nil
}

func (this *SSClient) Stop() error {
	if this.ln == nil {
		return nil
	}
	return this.ln.Close()
}

func (this *SSClient) Traffic() (uint64, error) {
	return 0, nil
}

func (this *SSClient) LocalPort() uint {
	return this.local_port
}

func NewClient(local_net NetProtocol, local_port uint, server_addr string, channel_net NetProtocol, pwd, crypt string) *SSClient {
	return &SSClient{
		crypt:       crypt,
		pwd:         pwd,
		local_net:   local_net,
		local_port:  local_port,
		channel_net: channel_net,
		server_addr: server_addr,
		log:         utils.NewLogger("SSClient"),
	}
}
