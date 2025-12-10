package widevine

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"

	"github.com/Spidey120703/go-mp4"
)

/*************************** CBCS ****************************/

type CryptoAction int

const (
	CryptoActionEncrypt CryptoAction = iota
	CryptoActionDecrypt
)

func CryptSampleCbcs(act CryptoAction, sample []byte, key []byte, iv []byte, tenc *mp4.Tenc, subsampleEntries []mp4.SubsampleEntry) (err error) {
	numberInCryptByte := int(tenc.DefaultCryptByteBlock) << 4
	numberInSkipByte := int(tenc.DefaultSkipByteBlock) << 4
	var pos uint32
	if len(subsampleEntries) > 0 {
		for _, subsampleEntry := range subsampleEntries {
			pos += uint32(subsampleEntry.BytesOfClearData)
			if subsampleEntry.BytesOfProtectedData == 0 {
				continue
			}
			if err = cbcsCrypt(act, sample[pos:pos+subsampleEntry.BytesOfProtectedData], key, iv, numberInCryptByte, numberInSkipByte); err != nil {
				return
			}
			pos += subsampleEntry.BytesOfProtectedData
		}
	} else {
		if err = cbcsCrypt(act, sample, key, iv, numberInCryptByte, numberInSkipByte); err != nil {
			return
		}
	}
	return
}

func cbcsCrypt(act CryptoAction, data []byte, key []byte, iv []byte, numberInCryptByte, numberInSkipByte int) (err error) {
	var pos int
	var size = len(data)
	var aesCbcCrypto cipher.Block
	if aesCbcCrypto, err = aes.NewCipher(key); err != nil {
		return
	}
	var crypter cipher.BlockMode
	switch act {
	case CryptoActionEncrypt:
		crypter = cipher.NewCBCEncrypter(aesCbcCrypto, iv)
	case CryptoActionDecrypt:
		crypter = cipher.NewCBCDecrypter(aesCbcCrypto, iv)
	default:
		return fmt.Errorf("unknown crypt action mode: %d", act)
	}
	if numberInSkipByte == 0 {
		numberToCryptByte := size & ^0xf
		crypter.CryptBlocks(data[:numberToCryptByte], data[:numberToCryptByte])
		return
	}
	for size-pos >= numberInCryptByte {
		crypter.CryptBlocks(data[pos:pos+numberInCryptByte], data[pos:pos+numberInCryptByte])
		pos += numberInCryptByte
		if size-pos < numberInSkipByte {
			return
		}
		pos += numberInSkipByte
	}

	return
}

func EncryptSampleCbcs(data []byte, key []byte, iv []byte, tenc *mp4.Tenc, subsample []mp4.SubsampleEntry) error {
	return CryptSampleCbcs(CryptoActionEncrypt, data, key, iv, tenc, subsample)
}

func DecryptSampleCbcs(data []byte, key []byte, iv []byte, tenc *mp4.Tenc, subsample []mp4.SubsampleEntry) error {
	return CryptSampleCbcs(CryptoActionDecrypt, data, key, iv, tenc, subsample)
}
