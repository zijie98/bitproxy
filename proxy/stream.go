/**
tcp udp proxy
**/
package proxy

import (
	"net"

	"rkproxy/libs"
	"rkproxy/proxy/ss"
	"rkproxy/utils"
)

type StreamProxy struct {
	local_port     uint
	local_net      ss.NetProtocol
	remote_host    string
	remote_port    uint
	rate           uint
	enable_traffic bool

	ln  net.Listener
	log *utils.Logger
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
			continue
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
	return libs.GetTraffic(this.local_port)
}

func (this *StreamProxy) handle(local_conn net.Conn) {
	defer func() {
		err := local_conn.Close()
		if err != nil {
			this.log.Info("local_conn close error ", err.Error())
		}
	}()

	remote_conn, err := net.Dial(
		string(this.local_net),
		utils.JoinHostPort(this.remote_host, this.remote_port),
	)
	if err != nil {
		this.log.Info("Dial remote host error: ", err)
		return
	}
	defer func() {
		err := remote_conn.Close()
		if err != nil {
			this.log.Info("Remote close error ", err.Error())
		}
	}()

	var limit = &utils.Limiter{Rate: this.rate}

	var traffic_stats = func(n int64) {
		if this.enable_traffic == false {
			return
		}
		libs.AddTrafficStats(this.local_port, n)
	}

	var done = make(chan bool, 2)
	var copy_data = func(dsc net.Conn, src net.Conn, l *utils.Limiter, timeout bool) {
		_, err := utils.Copy(dsc, src, l, traffic_stats)
		if err != nil {
			done <- true
		}
	}
	go copy_data(remote_conn, local_conn, limit, false)
	go copy_data(local_conn, remote_conn, nil, true)
	<-done
}

func NewStreamProxy(local_net ss.NetProtocol, local_port uint, remote_host string, remote_port uint, rate uint) *StreamProxy {
	return &StreamProxy{
		local_net:   local_net,
		local_port:  local_port,
		remote_host: remote_host,
		remote_port: remote_port,
		rate:        rate,
		log:         utils.NewLogger("TCP/UDP PROXY"),
	}
}
