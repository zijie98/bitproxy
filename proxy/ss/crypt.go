package ss

import (
	"crypto/cipher"
	"crypto/md5"
	"crypto/rc4"
	"crypto/sha1"
	"encoding/binary"
	"unsafe"

	"github.com/codahale/chacha20"
	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/salsa20/salsa"
	"bitproxy/utils"
)

type cipherInfo struct {
	keyLen    int
	ivLen     int
	newStream func(key, iv []byte, t DecOrEnc) (cipher.Stream, error)
}

const (
	Decrypt DecOrEnc = iota
	Encrypt
)
const (
	CryptXor      = "xor"
	CryptSalsa20  = "salsa20"
	CryptChacha20 = "chacha20"
	CryptRc4md5   = "rc4md5"
)

type DecOrEnc int

var CipherMethod = map[string]*cipherInfo{
	CryptXor:      {0, 0, newXorStream},
	CryptRc4md5:   {16, 16, newRC4MD5Stream},
	CryptChacha20: {32, 8, newChaCha20Stream},
	CryptSalsa20:  {32, 8, newSalsa20Stream},
}

func newRC4MD5Stream(key, iv []byte, _ DecOrEnc) (cipher.Stream, error) {
	h := md5.New()
	h.Write(key)
	h.Write(iv)
	rc4key := h.Sum(nil)
	return rc4.NewCipher(rc4key)
}

type XorStreamCipher struct {
	xortbl []byte
}

const saltxor = `sH3CIVoF#rWLtJo6`
const mtuLimit = 2048

func newXorStream(key, _ []byte, _ DecOrEnc) (cipher.Stream, error) {
	c := new(XorStreamCipher)
	c.xortbl = pbkdf2.Key(key, []byte(saltxor), 32, mtuLimit, sha1.New)
	return c, nil
}

const wordSize = int(unsafe.Sizeof(uintptr(0)))

func (c *XorStreamCipher) XORKeyStream(dst, src []byte) {
	c.xorBytes(dst, src, c.xortbl)
}
func (c *XorStreamCipher) xorBytes(dst, src, p []byte) {
	n := len(src)
	if len(p) < n {
		n = len(p)
	}

	w := n / wordSize
	if w > 0 {
		wordBytes := w * wordSize
		fastXORWords(dst[:wordBytes], src[:wordBytes], p[:wordBytes])
	}
	for i := (n - n%wordSize); i < n; i++ {
		dst[i] = src[i] ^ p[i]
	}
}

func fastXORWords(dst, a, b []byte) {
	dw := *(*[]uintptr)(unsafe.Pointer(&dst))
	aw := *(*[]uintptr)(unsafe.Pointer(&a))
	bw := *(*[]uintptr)(unsafe.Pointer(&b))
	n := len(b) / wordSize
	ex := n % 8
	for i := 0; i < ex; i++ {
		dw[i] = aw[i] ^ bw[i]
	}

	for i := ex; i < n; i += 8 {
		_dw := dw[i : i+8]
		_aw := aw[i : i+8]
		_bw := bw[i : i+8]
		_dw[0] = _aw[0] ^ _bw[0]
		_dw[1] = _aw[1] ^ _bw[1]
		_dw[2] = _aw[2] ^ _bw[2]
		_dw[3] = _aw[3] ^ _bw[3]
		_dw[4] = _aw[4] ^ _bw[4]
		_dw[5] = _aw[5] ^ _bw[5]
		_dw[6] = _aw[6] ^ _bw[6]
		_dw[7] = _aw[7] ^ _bw[7]
	}
}

func newChaCha20Stream(key, iv []byte, _ DecOrEnc) (cipher.Stream, error) {
	return chacha20.New(key, iv)
}

type salsaStreamCipher struct {
	nonce   [8]byte
	key     [32]byte
	counter int
}

func (c *salsaStreamCipher) XORKeyStream(dst, src []byte) {
	var buf []byte
	padLen := c.counter % 64
	dataSize := len(src) + padLen
	if cap(dst) >= dataSize {
		buf = dst[:dataSize]
	} else if utils.CopyPoolSize >= dataSize {
		buf = utils.CopyPool.Get(utils.CopyPoolSize)
		defer utils.CopyPool.Put(buf)
		buf = buf[:dataSize]
	} else {
		buf = make([]byte, dataSize)
	}

	var subNonce [16]byte
	copy(subNonce[:], c.nonce[:])
	binary.LittleEndian.PutUint64(subNonce[len(c.nonce):], uint64(c.counter/64))

	// It's difficult to avoid data copy here. src or dst maybe slice from
	// Conn.Read/Write, which can't have padding.
	copy(buf[padLen:], src[:])
	salsa.XORKeyStream(buf, buf, &subNonce, &c.key)
	copy(dst, buf[padLen:])

	c.counter += len(src)
}

func newSalsa20Stream(key, iv []byte, _ DecOrEnc) (cipher.Stream, error) {
	var c salsaStreamCipher
	copy(c.nonce[:], iv[:8])
	copy(c.key[:], key[:32])
	return &c, nil
}

func md5sum(d []byte) []byte {
	h := md5.New()
	h.Write(d)
	return h.Sum(nil)
}

// 将password md5加密, 根据md5(16位)长度不足时间重复计算
func EvpBytesToKey(password string, keyLen int) (key []byte) {
	const md5Len = 16

	cnt := (keyLen-1)/md5Len + 1
	m := make([]byte, cnt*md5Len)
	copy(m, md5sum([]byte(password)))

	// Repeatedly call md5 until bytes generated is enough.
	// Each call to md5 uses data: prev md5 sum + password.
	d := make([]byte, md5Len+len(password))
	start := 0
	for i := 1; i < cnt; i++ {
		start += md5Len
		copy(d, m[start-md5Len:start])
		copy(d[md5Len:], password)
		copy(m[start:], md5sum(d))
	}
	return m[:keyLen]
}
