package widevine

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"

	"github.com/Spidey120703/go-mp4"
)

/*************************** CBCS ****************************/

type CryptMode int

const (
	ModeEncrypt CryptMode = iota
	ModeDecrypt
)

func CryptSampleCbcs(cm CryptMode, data, key, iv []byte, subsampleEntries []mp4.SubsampleEntry, cryptByteBlock, skipByteBlock uint8) (err error) {
	numInCryptByte := int(cryptByteBlock) << 4
	numInSkipByte := int(skipByteBlock) << 4

	if len(subsampleEntries) != 0 {
		var pos uint32
		for _, subsampleEntry := range subsampleEntries {
			pos += uint32(subsampleEntry.BytesOfClearData)
			if subsampleEntry.BytesOfProtectedData == 0 {
				continue
			}
			if err = cbcsCrypt(cm, data[pos:pos+subsampleEntry.BytesOfProtectedData], key, iv, numInCryptByte, numInSkipByte); err != nil {
				return
			}
			pos += subsampleEntry.BytesOfProtectedData
		}
	} else {
		if err = cbcsCrypt(cm, data, key, iv, numInCryptByte, numInSkipByte); err != nil {
			return
		}
	}
	return
}

func cbcsCrypt(cm CryptMode, sample, key, iv []byte, numInCryptByte, numInSkipByte int) (err error) {
	var block cipher.Block
	if block, err = aes.NewCipher(key); err != nil {
		return
	}

	var mode cipher.BlockMode
	switch cm {
	case ModeEncrypt:
		mode = cipher.NewCBCEncrypter(block, iv)
	case ModeDecrypt:
		mode = cipher.NewCBCDecrypter(block, iv)
	default:
		return fmt.Errorf("unknown crypto action mode: %d", cm)
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
	return CryptSampleCbcs(ModeEncrypt, data, key, iv, subsampleEntries, cryptByteBlock, skipByteBlock)
}

func DecryptSampleCbcs(data, key, iv []byte, subsampleEntries []mp4.SubsampleEntry, cryptByteBlock, skipByteBlock uint8) error {
	if cryptByteBlock == 0 && skipByteBlock == 0 {
		return DecryptSampleCbc1(data, key, iv, subsampleEntries)
	}
	return CryptSampleCbcs(ModeDecrypt, data, key, iv, subsampleEntries, cryptByteBlock, skipByteBlock)
}
