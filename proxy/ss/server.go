/*
	Author moli
	基于kcp协议的shadowsocks服务器端
	* 代码有参考学习shadowsocks-go、kcptun的源码
*/

package ss

import (
	"crypto/sha1"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/xtaci/kcp-go"
	"github.com/xtaci/smux"
	"golang.org/x/crypto/pbkdf2"
	"io"
	"net"
	"rkproxy/log"
	"rkproxy/utils"
	"time"
)

type SSServer struct {
	crypt string
	pwd   string
	port  int

	ln  *kcp.Listener
	log *log.Logger
}

func (this *SSServer) getRequest(client io.Reader) (host string, extra []byte, err error) {
	const (
		rawType  = 0
		rawDmLen = 1

		LenTypeIpv4       = 1 + net.IPv4len + 2 // type + ipv4 len + port
		LenTypeIpv6       = 1 + net.IPv6len + 2
		LenTypeDomainBase = 1 + 1 + 2 // type + len + port

		ipv4Index   = 1
		ipv6Index   = 1
		domainIndex = 2
	)
	buf := make([]byte, 260)
	var n int
	if n, err = client.Read(buf); err != nil {
		this.log.Info("Read request buf err ", err)
		err = errors.New("Read request buf err " + err.Error())
		return
	}

	reqLen := 0
	switch buf[rawType] {
	case typeIpv4:
		reqLen = LenTypeIpv4
	case typeIpv6:
		reqLen = LenTypeIpv6
	case typeDomain:
		reqLen = int(buf[rawDmLen]) + LenTypeDomainBase
	default:
		this.log.Info("Raw addr err")
		err = errors.New("Raw addr err")
		return
	}

	if n < reqLen {
		// n 小于 reqLen 的长度，则认为传送过来的数据不完成，必须等去完整才能继续
		if _, err = io.ReadFull(client, buf[n:reqLen]); err != nil {
			return
		}
	} else if n > reqLen {
		// 如果传过来的内容超出（一般认为是客户端发的数据包过多）
		extra = buf[reqLen:n]
	}
	switch buf[rawType] {
	case typeIpv4:
		host = net.IP(buf[ipv4Index : ipv4Index+net.IPv4len]).String()
	case typeIpv6:
		host = net.IP(buf[ipv6Index : ipv6Index+net.IPv6len]).String()
	case typeDomain:
		host = string(buf[domainIndex : domainIndex+buf[rawDmLen]])
	}

	port := binary.BigEndian.Uint16(buf[reqLen-2 : reqLen])
	host = net.JoinHostPort(host, fmt.Sprintf("%d", port))
	return
}

func (this *SSServer) swapData(dst io.ReadWriter, src io.ReadWriter) {
	go utils.Copy(dst, src)
	utils.Copy(src, dst)
	//go io.Copy(dst, src)
	//io.Copy(src, dst)
}

func (this *SSServer) handle(client io.ReadWriteCloser) {

	smuxConfig := smux.DefaultConfig()

	session, err := smux.Server(client, smuxConfig)
	if err != nil {
		this.log.Info("kcp smux.Client error ", err)
		return
	}
	defer session.Close()

	for {
		client_stream, err := session.AcceptStream()
		if err != nil {
			this.log.Info("kcp session.OpenStream error ", err.Error())
			return
		}
		raw_host_addr, extra, err := this.getRequest(client_stream)
		if err != nil {
			this.log.Info("Request err ", err)
			return
		}

		raw_host, err := net.DialTimeout("tcp", raw_host_addr, 5*time.Second)
		if err != nil {
			this.log.Info("Dial to raw host err ", err, raw_host_addr)
			return
		}

		if extra != nil {
			//slog("Write extra : ", extra)
			if _, err = raw_host.Write(extra); err != nil {
				this.log.Info("Write extra date err ", err)
				return
			}
		}
		go this.swapData(client_stream, raw_host)
	}

}

func (this *SSServer) Start() error {
	var err error
	pass := pbkdf2.Key([]byte(this.pwd), []byte(SALT), 4096, 32, sha1.New)
	var block kcp.BlockCrypt

	switch this.crypt {
	case CryptXor:
		block, _ = kcp.NewSimpleXORBlockCrypt(pass)
	case CryptSalsa20:
		block, _ = kcp.NewSalsa20BlockCrypt(pass)
	default:
		this.log.Info("No suppert ", this.crypt)
		return errors.New("SSCLIENT: No suppert " + this.crypt)
	}

	this.ln, err = kcp.ListenWithOptions(fmt.Sprintf(":%d", this.port), block, 10, 3)
	if err != nil {
		this.log.Info("kcp.DialWithOptions error", err)
		return errors.New("SSCLIENT: kcp.DialWithOptions error " + err.Error())
	}
	this.log.Info("listen ", this.port)

	for {
		if conn, err := this.ln.AcceptKCP(); err == nil {
			conn.SetStreamMode(true)
			conn.SetNoDelay(1, 20, 1, 0)
			conn.SetMtu(1350)
			conn.SetWindowSize(1024, 1024)
			conn.SetACKNoDelay(true)
			conn.SetKeepAlive(10)

			this.log.Info("Accept address:", conn.RemoteAddr())
			go this.handle(conn)
		} else {
			this.log.Info("Accept err ", this.port, " ", err)
		}
	}

	return nil
}

func (this *SSServer) Stop() error {
	return this.ln.Close()
}

func (this *SSServer) LocalPort() int {
	return this.port
}

func NewServer(port int, pwd, crypt string) *SSServer {
	return &SSServer{
		crypt: crypt,
		pwd:   pwd,
		port:  port,
		log:   log.NewLogger("SSServer"),
	}
}
