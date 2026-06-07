package canary

import (
	"crypto/aes"
	"crypto/cipher"
)

type CBC struct {
	direction CipherDirection
	iv        []byte
	block     cipher.Block
	mode      cipher.BlockMode
}

func (cbc *CBC) SupportsPartialBlocks() bool {
	return false
}

func (cbc *CBC) Initialize(direction CipherDirection, key, iv []byte) (err error) {
	if cbc.block, err = aes.NewCipher(key); err != nil {
		return
	}
	cbc.direction = direction
	cbc.iv = iv
	cbc.Reset()
	return
}

func (cbc *CBC) Reset() {
	switch cbc.direction {
	case ModeEncrypt:
		cbc.mode = cipher.NewCBCEncrypter(cbc.block, cbc.iv)
	case ModeDecrypt:
		cbc.mode = cipher.NewCBCDecrypter(cbc.block, cbc.iv)
	}
}

func (cbc *CBC) CryptInplace(data []byte) {
	cbc.mode.CryptBlocks(data, data)
}
