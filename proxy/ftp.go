package proxy

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/molisoft/bitproxy/utils"
)

type FtpProxy struct {
	localPort  uint
	remoteHost string
	remotePort uint

	done bool

	log *utils.Logger
	ln  net.Listener
}

type FtpReturn struct {
	Ip   string
	Port uint
	Code int
}

func (r *FtpReturn) IsPassiveMode() bool {
	switch r.Code {
	case PASV, EPSV:
		return true
	}
	return false
}

func (c *FtpReturn) IsOpening() bool {
	if c.Code == OPENING {
		return true
	}
	return false
}

type FtpCommand struct {
	Command string
	Params  string
}

const (
	download = iota
	upload
)

type FtpCommandFlag struct {
	uploadOrDownload int
}

func (c *FtpCommandFlag) lastIsUpload() bool {
	if c.uploadOrDownload == upload {
		return true
	}
	return false
}
func (c *FtpCommandFlag) lastIsDownload() bool {
	if c.uploadOrDownload == download {
		return true
	}
	return false
}
func (c *FtpCommandFlag) Mark(command *FtpCommand) {
	if command.isUpload() {
		c.uploadOrDownload = upload
	} else if command.isDownload() {
		c.uploadOrDownload = download
	}
}

func (c *FtpCommand) isUpload() bool {
	if c.Command == C_STOR {
		return true
	}
	return false
}

func (c *FtpCommand) isDownload() bool {
	if c.Command == C_RETR {
		return true
	}
	return false
}

func (c *FtpCommand) isQuit() bool {
	if c.Command == C_QUIT {
		return true
	}
	return false
}

const (
	PASV    = 227
	EPSV    = 229
	QUIT    = 221
	OPENING = 150

	C_STOR = "STOR"
	C_RETR = "RETR"
	C_QUIT = "QUIT"
)

var errQuit = errors.New("Quit")
var errNotMatch = errors.New("Not match to the returned")
var errCommand = errors.New("Error command")

func (this *FtpProxy) handle(localConn net.Conn) {
	defer func() {
		err := localConn.Close()
		if err != nil {
			this.log.Info("local_conn close error ", err.Error())
		}
	}()

	remoteConn, err := net.DialTimeout("tcp", utils.JoinHostPort(this.remoteHost, this.remotePort), 30*time.Second)
	if err != nil {
		this.log.Info("Dial to remote host error: ", err)
		return
	}
	defer func() {
		err := remoteConn.Close()
		if err != nil {
			this.log.Info("remote_conn close error ", err.Error())
		}
	}()

	var readNotify = make(utils.ReadNotify)
	var done = make(chan bool, 5)
	var closed = false
	var commandMark FtpCommandFlag

	finish := func() {
		closed = true
		done <- true
	}
	notify := func(n int64, err error) {
		readNotify <- n
	}

	go func() {
		for !closed {
			select {
			case <-time.After(60 * 2 * time.Second):
				finish()
				return
			case <-readNotify:
			}
		}
	}()

	// remote -> self -> local
	go func() {
		var pasvResp *FtpReturn
		var pasvServerConn net.Conn
		var pasvClientConn net.Conn
		remote := bufio.NewReader(remoteConn)
		defer func() {
			if pasvServerConn != nil {
				pasvServerConn.Close()
			}
			if pasvClientConn != nil {
				pasvClientConn.Close()
			}
		}()
		for !closed {
			str, err := remote.ReadBytes('\n')
			if err != nil {
				break
			}
			readNotify <- 0
			resp, str, err := this.parseResult(str)
			if err != nil && err != errNotMatch {
				this.log.Info("Parse return error: ", err)
				if err == errQuit {
					finish()
					return
				}
				continue
			}
			if resp != nil {
				if resp.IsPassiveMode() {
					pasvResp = resp
					go func() {
						pasvClientConn, err = this.listenPasvConnFromClient(resp)
						if err != nil {
							this.log.Info("Lisent client conn error :", err)
							return
						}
						pasvServerConn, err = this.dialToServer(resp)
						if err != nil {
							this.log.Info("Dial to pasv ftp server error: ", err)
							return
						}
					}()
				}
				if pasvResp != nil && resp.IsOpening() {
					if commandMark.lastIsDownload() {
						go func() {
							if pasvClientConn == nil || pasvServerConn == nil {
								return
							}
							_, e := utils.CopyWithAfter(pasvClientConn, pasvServerConn, notify, notify)
							if e != nil {
								pasvServerConn.Close()
								pasvServerConn = nil
								pasvClientConn.Close()
								pasvClientConn = nil
							}
						}()
					} else {
						go func() {
							if pasvClientConn == nil || pasvServerConn == nil {
								return
							}
							_, e := utils.CopyWithAfter(pasvServerConn, pasvClientConn, notify, notify)
							if e != nil {
								pasvServerConn.Close()
								pasvServerConn = nil
								pasvClientConn.Close()
								pasvClientConn = nil
							}
						}()
					}
				}
			}
			localConn.Write(str)
		}
	}()

	// local -> self -> remote
	go func() {
		local := bufio.NewReader(localConn)
		for !closed {
			str, err := local.ReadBytes('\n')
			if err != nil {
				break
			}
			command, err := this.parseCommand(str)
			if err != nil {
				this.log.Info("Parse command error ", err)
				continue
			}
			if command.isQuit() {
				return
			}
			commandMark.Mark(command)
			remoteConn.Write(str)
		}
	}()
	//go copy_data(remote_conn, local_conn)

	<-done
	this.log.Info("exit .. ", localConn.RemoteAddr().String())
}

func (this *FtpProxy) parseCommand(b []byte) (command *FtpCommand, err error) {
	c := strings.Split(string(b), " ")
	command = &FtpCommand{}
	if len(c) == 0 {
		return nil, errCommand
	}
	if len(c) >= 1 {
		command.Command = c[0] //strings.TrimSpace(c[0])
	}
	if len(c) >= 2 {
		command.Params = c[1]
	}
	return command, nil
}

func (this *FtpProxy) parseResult(b []byte) (resp *FtpReturn, src []byte, err error) {
	src = b
	resp, err = this.parseFtpRetrun(string(src))
	if err != nil {
		return nil, nil, err
	}
	switch resp.Code {
	case PASV, EPSV:
		if err != nil {
			this.log.Info("Get pasv error: ", err)
			return nil, nil, err
		}
		if resp.Code == EPSV {
			src = []byte(fmt.Sprintf("229 Entering Extended Passive Mode (|||%d|).\r\n", resp.Port))
		} else if resp.Code == PASV {
			ip, _ := utils.PublicIp()
			ip = strings.Replace(ip, ".", ",", -1)
			port := fmt.Sprintf("%d,%d", resp.Port>>8, resp.Port&0xFF)
			src = []byte(fmt.Sprintf("227 Entering Passive Mode (%s,%s).\r\n", ip, port))
		}
	case QUIT:
		return nil, nil, errQuit
	}
	return
}

func (this *FtpProxy) dialToServer(pasv *FtpReturn) (conn net.Conn, err error) {
	serverAddress := utils.JoinHostPort(pasv.Ip, pasv.Port)
	conn, err = net.Dial("tcp", serverAddress)
	return
}

func (this *FtpProxy) listenPasvConnFromClient(resp *FtpReturn) (clientConn net.Conn, err error) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", resp.Port))
	if err != nil {
		this.log.Info("Pasv listen error ", err)
		return
	}
	clientConn, err = ln.Accept()
	if err != nil {
		this.log.Info("Accept client error : ", err)
		//if nerr, ok := err.(net.Error); ok && nerr.Temporary() {
	}
	return
}

func (this *FtpProxy) parseFtpRetrun(str string) (ret *FtpReturn, err error) {
	r := regexp.MustCompile(`\d+`)
	matches := r.FindAllString(str, -1)
	if matches == nil {
		return nil, errNotMatch
	}
	matchedCount := len(matches)

	uintBuff := make([]int, matchedCount)
	for i, e := range matches {
		uintBuff[i], err = strconv.Atoi(e)
		if err != nil {
			return nil, err
		}
	}
	ret = new(FtpReturn)
	if matchedCount >= 1 {
		// code
		ret.Code = int(uintBuff[0])
	}
	if matchedCount >= 2 {
		if matchedCount == 2 {
			// port
			ret.Port = uint(uintBuff[1])
			return
		}
		if matchedCount >= 7 {
			// host
			ip := net.IPv4(byte(uintBuff[1]), byte(uintBuff[2]), byte(uintBuff[3]), byte(uintBuff[4]))
			ret.Ip = ip.String()
			// port
			ret.Port = uint(uintBuff[5]*256 + uintBuff[6])
		}
	}
	return
}

func (this *FtpProxy) Start() (err error) {
	this.log.Info("Listen port", this.localPort)

	this.ln, err = net.Listen("tcp", utils.JoinHostPort("", this.localPort))
	if err != nil {
		this.log.Info("Can't Listen port: ", this.localPort, " ", err)
		return err
	}
	for !this.done {
		conn, err := this.ln.Accept()
		if err != nil {
			this.log.Info("Can't Accept: ", this.localPort, " ", err)
			if netErr, ok := err.(net.Error); ok && netErr.Temporary() {
				continue
			} else {
				return err
			}
		}
		this.log.Info("Accept ", conn.RemoteAddr().String())
		go this.handle(conn)
	}
	return nil
}

func (this *FtpProxy) Stop() error {
	this.done = true
	if this.ln != nil {
		this.ln.Close()
	}
	return nil
}

func (this *FtpProxy) LocalPort() uint {
	return this.localPort
}

func (this *FtpProxy) Traffic() (uint64, error) {
	return 0, nil
}

func NewFtpProxy(localPort uint, remoteHost string, remotePort uint) Proxyer {
	return &FtpProxy{
		localPort:  localPort,
		remoteHost: remoteHost,
		remotePort: remotePort,
		log:        utils.NewLogger("FtpProxy"),
	}
}
