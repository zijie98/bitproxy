package utils

import (
	"io"
	"net"
	"time"
)

var CopyPool BytePool

const CopyPoolSize = 4096 // 4KB
const maxNBuf = 1024

func init() {
	CopyPool.Init(1*time.Hour, maxNBuf)
}

type CopyCallbackFunc func(int64)

func Copy(dst io.Writer, src io.Reader, limit *Limiter, call CopyCallbackFunc) (written int64, err error) {
	buf := CopyPool.Get(CopyPoolSize)
	defer CopyPool.Put(buf)
	for {
		SetTimeout(src, time.Now().Add(60*time.Second))
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw > 0 {
				written += int64(nw)
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er != nil {
			//if er == io.EOF {
			//}
			err = er
			break
		}
		SetTimeout(src, time.Time{})
		if limit != nil {
			limit.Limit()
		}
	}
	if call != nil {
		call(written)
	}
	return written, err
}

func SetTimeout(src io.Reader, time time.Time) {
	switch src.(type) {
	case net.Conn:
		src.(net.Conn).SetReadDeadline(time)
	}
}

// 1s / (L/B) = sleep
type Limiter struct {
	Rate uint // KB/每秒
}

func (l *Limiter) Limit() {
	if l.Rate <= 0 {
		return
	}
	s := (1000 * 1000) / (l.Rate / (CopyPoolSize / 1024)) // 间隔时间
	t := time.Duration(s) * time.Microsecond              // t = ms
	//fmt.Printf("sleep ... %dns  - %dms \n", t, s)
	time.Sleep(t)
}
