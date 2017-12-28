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
	"time"

	"github.com/kataras/go-errors"
	"github.com/molisoft/bitproxy/proxy"
	"github.com/molisoft/bitproxy/utils"
	"github.com/xtaci/kcp-go"
)

type SSClient struct {
	crypt      string
	pwd        string
	localPort  uint
	localNet   proxy.NetProtocol // tcp/udp， 浏览器 -> ss客户端 之间的通信方式
	channelNet proxy.NetProtocol // tcp/udp/kcp， ss客户端 -> ss服务器 之间的通信方式
	serverAddr string            // 8.8.8.8:1990
	ln         net.Listener
	log        *utils.Logger
}

func (this *SSClient) handle(client io.ReadWriteCloser) {
	defer client.Close()

	buf := make([]byte, 32) // 32 byte
	_, err := client.Read(buf)
	if err != nil {
		this.log.Info("Client online err ", err)
		return
	}
	//clog("1 - ", buf)
	if buf[0] != SOCKS5_VERSION {
		this.log.Info("Only support socks5")
		return
	}

	buf = buf[:0] // reset
	buf = append(buf, SOCKS5_VERSION, 0)
	if _, err = client.Write(buf); err != nil {
		this.log.Info("Send msg to client err ", err)
		return
	}

	rawAddr, err := this.getRequestRemoteAddr(client)
	if err != nil {
		this.log.Info("RemoteAddrHandle err ", err, rawAddr)
		return
	}

	buf = buf[:0] // reset
	buf = append(buf, SOCKS5_VERSION, 0, 0, 0x1, 0, 0, 0, 0, 0, 0)
	if _, err := client.Write(buf); err != nil {
		this.log.Info("Send established to client err ", err)
		return
	}

	// connect to server
	serverConn, err := this.getServerConn()
	if err != nil {
		this.log.Info("Connect to server fail ", err)
		return
	}

	server, err := NewCryptConn(serverConn, this.pwd, this.crypt)
	if err != nil {
		this.log.Info("NewCryptConn err ", err)
		return
	}
	defer server.Close()

	// ss协议中，将把浏览器的请求发给服务器
	server.Write(rawAddr)

	go utils.CopyWithTimeout(server, client, nil, 60*time.Second)
	utils.CopyWithTimeout(client, server, nil, 60*time.Second)

	this.log.Info("handle is closed")
}

func (this *SSClient) getServerConn() (net.Conn, error) {
	if this.channelNet == proxy.UDP_PROTOCOL {
		return net.Dial("udp", this.serverAddr)

	} else if this.channelNet == proxy.TCP_PROTOCOL {
		return net.Dial("tcp", this.serverAddr)

	} else if this.channelNet == proxy.KCP_PROTOCOL {
		conn, err := kcp.DialWithOptions(this.serverAddr, nil, 10, 3)
		if err != nil {
			this.log.Info("SSCLIENT: kcp.DialWithOptions error", err)
			return nil, errors.New("SSCLIENT: kcp.DialWithOptions error " + err.Error())
		}
		conn.SetStreamMode(true)
		conn.SetNoDelay(1, 20, 2, 1)
		conn.SetMtu(1350)
		conn.SetWindowSize(1024, 1024)
		conn.SetACKNoDelay(true)
		return conn, nil
	} else {
		return nil, errors.New("Not found protocol: " + string(this.channelNet) + "?")
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
	if this.localNet == proxy.KCP_PROTOCOL {
		return errors.New("浏览器（或其他软件）连接到本客户端是不支持kcp协议的")
	}
	var err error
	this.ln, err = net.Listen(string(this.localNet), fmt.Sprintf(":%d", this.localPort))
	if err != nil {
		this.log.Info("Listen err ", err)
	}
	return nil
}

func (this *SSClient) Start() error {
	this.initListen()

	this.log.Info("Listen port", this.localPort)
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
	return this.localPort
}

func NewClient(localNet proxy.NetProtocol, localPort uint, serverAddr string, channelNet proxy.NetProtocol, pwd, crypt string) proxy.Proxyer {
	return &SSClient{
		crypt:      crypt,
		pwd:        pwd,
		localNet:   localNet,
		localPort:  localPort,
		channelNet: channelNet,
		serverAddr: serverAddr,
		log:        utils.NewLogger("SSClient"),
	}
}
