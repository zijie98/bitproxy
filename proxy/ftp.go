package proxy

import (
	"net"
	"rkproxy/log"
	"rkproxy/utils"
)

type FtpProxy struct {
	local_port  uint
	remote_host string
	remote_port uint
	log         *log.Logger

	ln net.Listener
}

func (this *FtpProxy) handle(conn net.Conn) {
	this.log.Info("nothing..")
}

func (this *FtpProxy) Start() (err error) {
	this.ln, err = net.Listen("tcp", utils.JoinHostPort("", this.local_port))
	if err != nil {
		this.log.Info("Can't Listen port: ", this.local_port, " ", err)
		return err
	}
	for {
		conn, err := this.ln.Accept()
		if err != nil {
			this.log.Info("Can't Accept: ", this.local_port, " ", err)
		}
		this.handle(conn)
	}
	return nil
}

func (tihs *FtpProxy) Stop() error {
	return nil
}

func (this *FtpProxy) LocalPort() uint {
	return this.local_port
}

func (this *FtpProxy) Traffic() (uint64, error) {
	return 0, nil
}

func NewFtpProxy(local_port uint, remote_host string, remote_port uint) *FtpProxy {
	return &FtpProxy{
		local_port:  local_port,
		remote_host: remote_host,
		remote_port: remote_port,
		log:         log.NewLogger("FtpProxy"),
	}
}
