package canary

import (
	"crypto/aes"
	"crypto/cipher"

	"github.com/Spidey120703/go-mp4"
)

type ICryptor interface {
	Crypt(m Mode, data, key, iv []byte, subsampleEntries []mp4.SubsampleEntry, cryptByteBlock, skipByteBlock int) (err error)
	Decrypt(data, key, iv []byte, subsampleEntries []mp4.SubsampleEntry, cryptByteBlock, skipByteBlock uint8) (err error)
	Encrypt(data, key, iv []byte, subsampleEntries []mp4.SubsampleEntry, cryptByteBlock, skipByteBlock uint8) (err error)
}

type ICipher interface {
	Initializer(mode Mode, key, iv []byte) (err error)
	CryptInplace(data []byte)
}

type Mode int

const (
	ModeEncrypt Mode = iota
	ModeDecrypt
)

type CTR struct {
	block  cipher.Block
	stream cipher.Stream
}

func (ctr *CTR) Initializer(_ Mode, key, iv []byte) (err error) {
	if ctr.block, err = aes.NewCipher(key); err != nil {
		return
	}
	ctr.stream = cipher.NewCTR(ctr.block, iv)
	return
}

func (ctr *CTR) CryptInplace(data []byte) {
	ctr.stream.XORKeyStream(data, data)
}

type CBC struct {
	block cipher.Block
	mode  cipher.BlockMode
}

func (cbc *CBC) Initializer(mode Mode, key, iv []byte) (err error) {
	if cbc.block, err = aes.NewCipher(key); err != nil {
		return
	}
	switch mode {
	case ModeEncrypt:
		cbc.mode = cipher.NewCBCEncrypter(cbc.block, iv)
	case ModeDecrypt:
		cbc.mode = cipher.NewCBCDecrypter(cbc.block, iv)
	}
	return
}

func (cbc *CBC) CryptInplace(data []byte) {
	cbc.mode.CryptBlocks(data, data)
}

type SampleCryptor struct {
	cph ICipher
}

func (s *SampleCryptor) Crypt(m Mode, data, key, iv []byte, subsampleEntries []mp4.SubsampleEntry, _, _ uint8) (err error) {
	if err = s.cph.Initializer(m, key, iv); err != nil {
		return
	}

	if len(subsampleEntries) != 0 {
		var pos uint32
		for _, subsampleEntry := range subsampleEntries {
			pos += uint32(subsampleEntry.BytesOfClearData)
			if subsampleEntry.BytesOfProtectedData == 0 {
				continue
			}
			s.cph.CryptInplace(data[pos : pos+subsampleEntry.BytesOfProtectedData])
			pos += subsampleEntry.BytesOfProtectedData
		}
	} else {
		s.cph.CryptInplace(data)
	}
	return
}

func (s *SampleCryptor) Decrypt(data, key, iv []byte, subsampleEntries []mp4.SubsampleEntry, _, _ uint8) (err error) {
	return s.Crypt(ModeDecrypt, data, key, iv, subsampleEntries, 0, 0)
}

func (s *SampleCryptor) Encrypt(data, key, iv []byte, subsampleEntries []mp4.SubsampleEntry, _, _ uint8) (err error) {
	return s.Crypt(ModeEncrypt, data, key, iv, subsampleEntries, 0, 0)
}

type SubsampleCryptor struct {
	cph ICipher
}

func (s *SubsampleCryptor) crypt(data []byte, numInCryptByte, numInSkipByte int) (err error) {
	var size = len(data)

	if numInSkipByte == 0 {
		numToCryptByte := size & ^0xf
		s.cph.CryptInplace(data[:numToCryptByte])
		return
	}

	var pos int
	for size-pos >= numInCryptByte {
		s.cph.CryptInplace(data[pos : pos+numInCryptByte])
		pos += numInCryptByte
		if size-pos < numInSkipByte {
			return
		}
		pos += numInSkipByte
	}

	return
}

func (s *SubsampleCryptor) Crypt(m Mode, data, key, iv []byte, subsampleEntries []mp4.SubsampleEntry, cryptByteBlock, skipByteBlock uint8) (err error) {
	numInCryptByte := int(cryptByteBlock) << 4
	numInSkipByte := int(skipByteBlock) << 4

	if err = s.cph.Initializer(m, key, iv); err != nil {
		return
	}

	if len(subsampleEntries) != 0 {
		var pos uint32
		for _, subsampleEntry := range subsampleEntries {
			pos += uint32(subsampleEntry.BytesOfClearData)
			if subsampleEntry.BytesOfProtectedData == 0 {
				continue
			}
			if err = s.crypt(data[pos:pos+subsampleEntry.BytesOfProtectedData], numInCryptByte, numInSkipByte); err != nil {
				return
			}
			pos += subsampleEntry.BytesOfProtectedData
		}
	} else {
		if err = s.crypt(data, numInCryptByte, numInSkipByte); err != nil {
			return
		}
	}
	return
}

func (s *SubsampleCryptor) Decrypt(data, key, iv []byte, subsampleEntries []mp4.SubsampleEntry, cryptByteBlock, skipByteBlock uint8) (err error) {
	return s.Crypt(ModeDecrypt, data, key, iv, subsampleEntries, cryptByteBlock, skipByteBlock)
}

func (s *SubsampleCryptor) Encrypt(data, key, iv []byte, subsampleEntries []mp4.SubsampleEntry, cryptByteBlock, skipByteBlock uint8) (err error) {
	return s.Crypt(ModeEncrypt, data, key, iv, subsampleEntries, cryptByteBlock, skipByteBlock)
}
