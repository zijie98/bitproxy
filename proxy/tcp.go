package proxy

import (
	"net"
	"rkproxy/log"
	"rkproxy/utils"
	"time"
)

const (
	DefaultDialTimeout = 3
)

type TcpProxy struct {
	local_port  int
	remote_host string
	remote_port int

	ln net.Listener
}

func (this *TcpProxy) Start() error {
	var err error
	this.ln, err = net.Listen("tcp", utils.JoinHostPort("", this.local_port))
	if err != nil {
		log.Info("TCP PROXY: Can't Listen port: ", this.local_port, " ", err)
		return err
	}
	log.Info("TCP PROXY: Start tcp proxy.")
	for {
		conn, err := this.ln.Accept()
		if err != nil {
			log.Info("TCP PROXY: Can't Accept: ", err)
			break
		}
		this.serveHandle(conn)
	}
	return nil
}

func (this *TcpProxy) Stop() error {
	return this.ln.Close()
}

func (this *TcpProxy) LocalPort() int {
	return this.local_port
}

func (this *TcpProxy) serveHandle(local_conn net.Conn) {
	remote_conn, err := net.DialTimeout(
		"tcp",
		utils.JoinHostPort(this.remote_host, this.remote_port),
		DefaultDialTimeout*time.Second,
	)
	if err != nil {
		log.Info("TCP PROXY: Dial remote host fall: ", err)
		return
	}
	go utils.Copy(remote_conn, local_conn)
	go utils.Copy(local_conn, remote_conn)
}

func NewTcpProxy(local_port int, remote_host string, remote_port int) *TcpProxy {
	return &TcpProxy{
		local_port:  local_port,
		remote_host: remote_host,
		remote_port: remote_port,
	}
}
