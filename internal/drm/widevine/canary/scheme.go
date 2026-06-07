package canary

import (
	"github.com/Spidey120703/go-mp4"
)

type CipherDirection int

const (
	ModeEncrypt CipherDirection = iota
	ModeDecrypt
)

type ICryptor interface {
	Crypt(m CipherDirection, data, key, iv []byte, subsampleEntries []mp4.SubsampleEntry, cryptByteBlock, skipByteBlock uint8) (err error)
}

type ICryptOperator interface {
	Decrypt(data, key, iv []byte, subsampleEntries []mp4.SubsampleEntry, cryptByteBlock, skipByteBlock uint8) (err error)
	Encrypt(data, key, iv []byte, subsampleEntries []mp4.SubsampleEntry, cryptByteBlock, skipByteBlock uint8) (err error)
}

type ICipher interface {
	Initialize(direction CipherDirection, key, iv []byte) (err error)
	Reset()
	CryptInplace(data []byte)
	SupportsPartialBlocks() bool
}

type CryptOperator struct {
	ICryptor
}

func (c *CryptOperator) Decrypt(data, key, iv []byte, subsampleEntries []mp4.SubsampleEntry, cryptByteBlock, skipByteBlock uint8) (err error) {
	return c.Crypt(ModeDecrypt, data, key, iv, subsampleEntries, cryptByteBlock, skipByteBlock)
}

func (c *CryptOperator) Encrypt(data, key, iv []byte, subsampleEntries []mp4.SubsampleEntry, cryptByteBlock, skipByteBlock uint8) (err error) {
	return c.Crypt(ModeEncrypt, data, key, iv, subsampleEntries, cryptByteBlock, skipByteBlock)
}
