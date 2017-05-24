package ss

import (
	"crypto/cipher"
	"crypto/md5"
	"crypto/rc4"
	"encoding/binary"

	"github.com/codahale/chacha20"
	"golang.org/x/crypto/salsa20/salsa"
	"rkproxy/utils"
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
	CryptNOT      = "xor"
	CryptSalsa20  = "salsa20"
	CryptChacha20 = "chacha20"
	CryptRc4md5   = "rc4md5"
)

type DecOrEnc int

var CipherMethod = map[string]*cipherInfo{
	CryptNOT:      {0, 0, newNOTStream},
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

type NotStreamCipher struct {
}

func (c *NotStreamCipher) XORKeyStream(dst, src []byte) {
	copy(dst, src)
	for i := 0; i < len(dst); i += 8 {
		if i+8 > len(dst) {
			for x := i; x < len(dst); x++ {
				dst[x] ^= 0xFF
			}
			break
		}
		buf := dst[i : i+8]
		uint64buf := binary.LittleEndian.Uint64(buf)
		uint64buf ^= 0xFFFFFFFFFFFFFFFF
		binary.LittleEndian.PutUint64(buf, uint64buf)
	}
}

func newNOTStream(_, _ []byte, _ DecOrEnc) (cipher.Stream, error) {
	return &NotStreamCipher{}, nil
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
