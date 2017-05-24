package proxy

import (
	"net"
	"time"

	"rkproxy/log"
	"rkproxy/proxy/ss"
	"rkproxy/utils"
)

const (
	DefaultDialTimeout = 5
)

type TcpProxy struct {
	local_port  int
	local_net   ss.NetProtocol
	remote_host string
	remote_port int

	ln  net.Listener
	log *log.Logger
}

func (this *TcpProxy) Start() error {
	var err error
	this.ln, err = net.Listen(string(this.local_net), utils.JoinHostPort("", this.local_port))
	if err != nil {
		this.log.Info("Can't Listen port: ", this.local_port, " ", err)
		return err
	}
	this.log.Info("Start tcp/udp proxy.")
	for {
		conn, err := this.ln.Accept()
		if err != nil {
			this.log.Info("Can't Accept: ", err)
			break
		}
		go this.serveHandle(conn)
	}
	return nil
}

func (this *TcpProxy) Stop() error {
	if this.ln == nil {
		return nil
	}
	return this.ln.Close()
}

func (this *TcpProxy) LocalPort() int {
	return this.local_port
}

func (this *TcpProxy) serveHandle(local_conn net.Conn) {
	remote_conn, err := net.DialTimeout(
		string(this.local_net),
		utils.JoinHostPort(this.remote_host, this.remote_port),
		DefaultDialTimeout*time.Second,
	)
	if err != nil {
		this.log.Info("Dial remote host fall: ", err)
		return
	}
	go utils.Copy(remote_conn, local_conn)
	utils.Copy(local_conn, remote_conn)
}

func NewTcpProxy(local_net ss.NetProtocol, local_port int, remote_host string, remote_port int) *TcpProxy {
	return &TcpProxy{
		local_net:   local_net,
		local_port:  local_port,
		remote_host: remote_host,
		remote_port: remote_port,
		log:         log.NewLogger("TCP/UDP PROXY"),
	}
}
