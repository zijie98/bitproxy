/*
	Author moli
	2016-12-03 00：27
	ss 客户端, 使用kcp（udp）协议，加速网络环境不好的情况下提高网络传输质量。如果网络质量好，直接TCP ……
	ss client <-> ss server

	ss.StartSSClient(local_port)
*/

package ss

import (
	kcp "github.com/xtaci/kcp-go"
	//
	"crypto/sha1"
	"fmt"
	"github.com/kataras/go-errors"
	"github.com/xtaci/smux"
	"golang.org/x/crypto/pbkdf2"
	"io"
	"net"
	"rkproxy/log"
	"rkproxy/utils"
)

type SSClient struct {
	crypt      string
	pwd        string
	local_port int

	server_addr string // 8.8.8.8:1990

	ln net.Listener
}

func clog(msg ...interface{}) {
	log.Info("SSClient", msg)
}

func (this *SSClient) handle(client io.ReadWriter) {

	buf := make([]byte, 32) // 32 byte
	_, err := client.Read(buf)
	if err != nil {
		clog("Client online err ", err)
		return
	}
	//clog("1 - ", buf)
	if buf[0] != socks5Version {
		clog("Only suppert socks5")
		return
	}

	buf = buf[:0] // reset
	buf = append(buf, socks5Version, 0)
	if _, err = client.Write(buf); err != nil {
		clog("Send msg to client err ", err)
		return
	}

	raw_addr, err := this.getRequestRemoteAddr(client)
	if err != nil {
		clog("RemoteAddrHandle err ", err, raw_addr)
		return
	}

	buf = buf[:0] // reset
	buf = append(buf, socks5Version, 0, 0, 0x1, 0, 0, 0, 0, 0, 0)
	if _, err := client.Write(buf); err != nil {
		clog("Send established to client err ", err)
		return
	}

	//clog("3 - ", raw_addr)
	// connect to server
	server_conn, err := this.createServerConn()
	if err != nil {
		clog("Connect to server fail ", err)
		return
	}

	// ss协议中，将把浏览器的请求发给服务器
	server_conn.Write(raw_addr)

	//server_conn , err := net.Dial("tcp", this.host(raw_addr))

	go utils.Copy(client, server_conn)
	utils.Copy(server_conn, client)
	//go io.Copy(client, server_conn)
	//io.Copy(server_conn, client)
}

// 创建Server连接
func (this *SSClient) createServerConn() (io.ReadWriteCloser, error) {

	pass := pbkdf2.Key([]byte(this.pwd), []byte(SALT), 4096, 32, sha1.New)
	var block kcp.BlockCrypt

	switch this.crypt {
	case CryptXor:
		block, _ = kcp.NewSimpleXORBlockCrypt(pass)
	case CryptSalsa20:
		block, _ = kcp.NewSalsa20BlockCrypt(pass)
	default:
		log.Info("SSCLIENT: No suppert ", this.crypt)
		return nil, errors.New("SSCLIENT: No suppert " + this.crypt)
	}

	smuxConfig := smux.DefaultConfig()

	conn, err := kcp.DialWithOptions(this.server_addr, block, 10, 3)
	if err != nil {
		log.Info("SSCLIENT: kcp.DialWithOptions error", err)
		return nil, errors.New("SSCLIENT: kcp.DialWithOptions error " + err.Error())
	}
	conn.SetStreamMode(true)
	conn.SetNoDelay(1, 20, 1, 0)
	conn.SetMtu(1350)
	conn.SetWindowSize(1024, 1024)
	conn.SetACKNoDelay(true)
	conn.SetKeepAlive(10)

	session, err := smux.Client(conn, smuxConfig)
	if err != nil {
		log.Info("SSCLIENT: kcp smux.Client error ", err)
		return nil, errors.New("SSCLIENT: kcp smux.Client error " + err.Error())
	}
	stream, err := session.OpenStream()
	if err != nil {
		log.Info("SSCLIENT: kcp session.OpenStream error ", err.Error())
		return nil, errors.New("SSCLIENT: kcp session.OpenStream error " + err.Error())
	}
	return stream, nil
}

func (this *SSClient) getRequestRemoteAddr(client io.ReadWriter) (buf []byte, err error) {
	const (
		LenTypeIpv4       = 3 + 1 + net.IPv4len + 2
		LenTypeIpv6       = 3 + 1 + net.IPv6len + 2
		LenTypeDoaminBase = 3 + 1 + 1 + 2

		rawType = 3
		rawAddr = 4
	)

	buf = make([]byte, 260) // 260 byte

	n, err := client.Read(buf)
	if err != nil || n < rawType {
		clog("Read error ", err)
		return nil, err
	}

	reqLen := -1
	switch buf[rawType] {
	case typeIpv4:
		reqLen = LenTypeIpv4
	case typeIpv6:
		reqLen = LenTypeIpv6
	case typeDomain:
		reqLen = int(buf[rawAddr]) + LenTypeDoaminBase
	default:
		clog("Raw addr err", &client)
		return nil, errors.New("Raw addr err")
	}

	if n < reqLen {
		if _, err := io.ReadFull(client, buf[n:reqLen]); err != nil {
			clog("ReadFull err..", err)
			return nil, err
		}
	}
	buf = buf[rawType:reqLen]
	return
}

func (this *SSClient) Start() error {
	var err error
	this.ln, err = net.Listen("tcp", fmt.Sprintf(":%d", this.local_port))
	if err != nil {
		clog("Listen err ", err)
		return err
	}
	clog("Listen port", this.local_port)
	for {
		conn, err := this.ln.Accept()
		if err != nil {
			clog("Accept err ", err)
			break
		}
		clog("Accept ip ", conn.RemoteAddr(), "client ", &conn)
		go this.handle(conn)
	}
	return nil
}

func (this *SSClient) Stop() error {
	return this.ln.Close()
}

func (this *SSClient) LocalPort() int {
	return this.local_port
}

func NewClient(local_port int, server_addr, pwd, crypt string) *SSClient {
	return &SSClient{
		crypt:       crypt,
		pwd:         pwd,
		local_port:  local_port,
		server_addr: server_addr,
	}
}
