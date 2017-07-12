/**
tcp udp proxy
**/
package proxy

import (
	"net"
	"time"

	"io"
	"rkproxy/log"
	"rkproxy/proxy/ss"
	"rkproxy/utils"
)

const (
	DefaultDialTimeout = 5
)

type StreamProxy struct {
	local_port  uint
	local_net   ss.NetProtocol
	remote_host string
	remote_port uint

	ln  net.Listener
	log *log.Logger
}

func (this *StreamProxy) Start() (err error) {
	this.ln, err = net.Listen(string(this.local_net), utils.JoinHostPort("", this.local_port))
	if err != nil {
		this.log.Info("Can't Listen port: ", this.local_port, " ", err)
		return err
	}
	this.log.Info("Listen port", this.local_port)
	for {
		conn, err := this.ln.Accept()
		if err != nil {
			this.log.Info("Can't Accept: ", err)
			break
		}
		this.log.Info("Client ip", conn.RemoteAddr().String())
		go this.handle(conn)
	}
	return nil
}

func (this *StreamProxy) Stop() error {
	if this.ln == nil {
		return nil
	}
	return this.ln.Close()
}

func (this *StreamProxy) LocalPort() uint {
	return this.local_port
}

func (this *StreamProxy) Traffic() (uint64, error) {
	return 0, nil
}

func (this *StreamProxy) handle(local_conn net.Conn) {
	remote_conn, err := net.DialTimeout(
		string(this.local_net),
		utils.JoinHostPort(this.remote_host, this.remote_port),
		DefaultDialTimeout*time.Second,
	)
	if err != nil {
		this.log.Info("Dial remote host fall: ", err)
		return
	}
	var done = make(chan bool, 2)
	go func(dsc io.WriteCloser, src io.ReadCloser) {
		_, err := utils.Copy(dsc, src, nil)
		if err != nil {
			done <- true
		}
	}(remote_conn, local_conn)

	go func(dsc io.WriteCloser, src io.ReadCloser) {
		_, err := utils.Copy(dsc, src, nil)
		if err != nil {
			done <- true
		}
	}(local_conn, remote_conn)

	<-done
	local_conn.Close()
	remote_conn.Close()
}

func NewStreamProxy(local_net ss.NetProtocol, local_port uint, remote_host string, remote_port uint) *StreamProxy {
	return &StreamProxy{
		local_net:   local_net,
		local_port:  local_port,
		remote_host: remote_host,
		remote_port: remote_port,
		log:         log.NewLogger("TCP/UDP PROXY"),
	}
}
