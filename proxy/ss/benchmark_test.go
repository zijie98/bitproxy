package ss

import (
	"encoding/binary"
	"testing"
)

func BenchmarkNotStreamCipher_XORKeyStream2(b *testing.B) {
	b.N = 10000

	buff := make([]byte, 14049)
	for i := 0; i < 14049; i++ {
		buff[i] = 'A'
	}

	for i := 0; i < b.N; i++ {

		l := len(buff)
		for i := 0; i < l; i++ {
			*(&buff[i]) ^= 0xFF
		}
	}
}

func BenchmarkNotStreamCipher_XORKeyStream(b *testing.B) {
	b.N = 10000

	buff := make([]byte, 14049)
	for i := 0; i < 14049; i++ {
		buff[i] = 'A'
	}

	for c := 0; c < b.N; c++ {
		for i := 0; i < len(buff); i += 8 {
			if i+8 > len(buff) {
				for x := i; x < len(buff); x++ {
					buff[x] ^= 0xFF
				}
				break
			}
			buf := buff[i : i+8]
			uint64buf := binary.LittleEndian.Uint64(buf)
			uint64buf ^= 0xFFFFFFFFFFFFFFFF
			binary.LittleEndian.PutUint64(buf, uint64buf)
		}
	}
}
