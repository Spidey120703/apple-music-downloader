package lagacy

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"

	"github.com/Spidey120703/go-mp4"
)

/*************************** CBCS ****************************/

type CryptDirection int

const (
	DirEncrypt CryptDirection = iota
	DirDecrypt
)

func CryptSampleCbcs(dir CryptDirection, data, key, iv []byte, subsampleEntries []mp4.SubsampleEntry, cryptByteBlock, skipByteBlock uint8) (err error) {
	numInCryptByte := int(cryptByteBlock) << 4
	numInSkipByte := int(skipByteBlock) << 4

	if len(subsampleEntries) != 0 {
		var pos uint32
		for _, subsampleEntry := range subsampleEntries {
			pos += uint32(subsampleEntry.BytesOfClearData)
			if subsampleEntry.BytesOfProtectedData == 0 {
				continue
			}
			if err = cbcsCrypt(dir, data[pos:pos+subsampleEntry.BytesOfProtectedData], key, iv, numInCryptByte, numInSkipByte); err != nil {
				return
			}
			pos += subsampleEntry.BytesOfProtectedData
		}
	} else {
		if err = cbcsCrypt(dir, data, key, iv, numInCryptByte, numInSkipByte); err != nil {
			return
		}
	}
	return
}

func cbcsCrypt(dir CryptDirection, sample, key, iv []byte, numInCryptByte, numInSkipByte int) (err error) {
	var block cipher.Block
	if block, err = aes.NewCipher(key); err != nil {
		return
	}

	var mode cipher.BlockMode
	switch dir {
	case DirEncrypt:
		mode = cipher.NewCBCEncrypter(block, iv)
	case DirDecrypt:
		mode = cipher.NewCBCDecrypter(block, iv)
	default:
		return fmt.Errorf("unknown crypto action mode: %d", dir)
	}

	var size = len(sample)

	if numInSkipByte == 0 {
		numToCryptByte := size & ^0xf
		mode.CryptBlocks(sample[:numToCryptByte], sample[:numToCryptByte])
		return
	}

	var pos int
	for size-pos >= numInCryptByte {
		mode.CryptBlocks(sample[pos:pos+numInCryptByte], sample[pos:pos+numInCryptByte])
		pos += numInCryptByte
		if size-pos < numInSkipByte {
			return
		}
		pos += numInSkipByte
	}

	return
}

func EncryptSampleCbcs(data, key, iv []byte, subsampleEntries []mp4.SubsampleEntry, cryptByteBlock, skipByteBlock uint8) error {
	if cryptByteBlock == 0 && skipByteBlock == 0 {
		return EncryptSampleCbc1(data, key, iv, subsampleEntries)
	}
	return CryptSampleCbcs(DirEncrypt, data, key, iv, subsampleEntries, cryptByteBlock, skipByteBlock)
}

func DecryptSampleCbcs(data, key, iv []byte, subsampleEntries []mp4.SubsampleEntry, cryptByteBlock, skipByteBlock uint8) error {
	if cryptByteBlock == 0 && skipByteBlock == 0 {
		return DecryptSampleCbc1(data, key, iv, subsampleEntries)
	}
	return CryptSampleCbcs(DirDecrypt, data, key, iv, subsampleEntries, cryptByteBlock, skipByteBlock)
}
