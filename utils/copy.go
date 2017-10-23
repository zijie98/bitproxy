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

type AfterCallbackFunc func(int64, error)
type BeforeCallBackFunc func()
type ReadNotify chan int64

func CopyWithBefore(dst io.Writer, src io.Reader, beforeReadFunc BeforeCallBackFunc, beforeWriteFunc BeforeCallBackFunc) (written int64, err error) {
	return Copy(dst, src, nil, beforeReadFunc, beforeWriteFunc, nil, nil, nil, 0)
}

func CopyWithAfter(dst io.Writer, src io.Reader, afterReadFunc AfterCallbackFunc, afterWriteFunc AfterCallbackFunc) (written int64, err error) {
	return Copy(dst, src, nil, nil, nil, afterReadFunc, afterWriteFunc, nil, 0)
}

func CopyWithNone(dst io.Writer, src io.Reader) (written int64, err error) {
	return Copy(dst, src, nil, nil, nil, nil, nil, nil, 0)
}

func CopyWithTimeout(dst io.Writer, src io.Reader, afterReadFunc AfterCallbackFunc, timeout time.Duration) (w int64, err error) {
	return Copy(dst, src, nil, nil, nil, afterReadFunc, nil, nil, timeout)
}

func Copy(dst io.Writer, src io.Reader, limit *Limit, beforeReadFunc BeforeCallBackFunc, beforeWriteFunc BeforeCallBackFunc, afterReadFunc AfterCallbackFunc, afterWriteFunc AfterCallbackFunc, exitedFunc AfterCallbackFunc, timeout time.Duration) (written int64, err error) {
	buf := CopyPool.Get(CopyPoolSize)
	defer CopyPool.Put(buf)
	defer func() {
		if exitedFunc != nil {
			exitedFunc(written, err)
		}
	}()
	for {
		if timeout != 0 {
			SetTimeout(src, timeout)
		}
		if beforeReadFunc != nil {
			beforeReadFunc()
		}
		nr, er := src.Read(buf)
		if afterReadFunc != nil {
			afterReadFunc(int64(nr), er)
		}
		if nr > 0 {
			if beforeWriteFunc != nil {
				beforeWriteFunc()
			}
			//fmt.Println(string(buf[0:nr]))
			nw, ew := dst.Write(buf[0:nr])
			if afterWriteFunc != nil {
				afterWriteFunc(int64(nw), ew)
			}
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
			err = er
			break
		}
		if limit != nil {
			limit.Limit()
		}
	}
	return written, err
}

func SetTimeout(src io.Reader, d time.Duration) {
	t := time.Now()
	switch src.(type) {
	case net.Conn:
		src.(net.Conn).SetReadDeadline(t.Add(d))
	}
}

// 1s / (L/B) = sleep
type Limit struct {
	Rate uint // KB/每秒
}

func (l *Limit) Limit() {
	if l.Rate <= 0 {
		return
	}
	s := (1000 * 1000) / (l.Rate / (CopyPoolSize / 1024)) // 间隔时间
	t := time.Duration(s) * time.Microsecond              // t = ms
	//fmt.Printf("sleep ... %dns  - %dms \n", t, s)
	time.Sleep(t)
}
