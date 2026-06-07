package canary

import (
	"crypto/aes"
	"crypto/cipher"
)

type CTR struct {
	iv     []byte
	block  cipher.Block
	stream cipher.Stream
}

func (ctr *CTR) SupportsPartialBlocks() bool {
	return true
}

func (ctr *CTR) Initialize(_ CipherDirection, key, iv []byte) (err error) {
	if ctr.block, err = aes.NewCipher(key); err != nil {
		return
	}
	ctr.iv = iv
	ctr.stream = cipher.NewCTR(ctr.block, ctr.iv)
	return
}

func (ctr *CTR) Reset() {}

func (ctr *CTR) CryptInplace(data []byte) {
	ctr.stream.XORKeyStream(data, data)
}
