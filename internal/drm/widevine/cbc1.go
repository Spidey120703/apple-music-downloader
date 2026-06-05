package widevine

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"

	"github.com/Spidey120703/go-mp4"
)

func CryptSampleCbc1(cm CryptMode, data, key, iv []byte, subsampleEntries []mp4.SubsampleEntry) (err error) {
	if len(subsampleEntries) != 0 {
		var pos uint32
		for _, subsampleEntry := range subsampleEntries {
			pos += uint32(subsampleEntry.BytesOfClearData)
			if subsampleEntry.BytesOfProtectedData == 0 {
				continue
			}
			if err = cbc1Crypt(cm, data[pos:pos+subsampleEntry.BytesOfProtectedData], key, iv); err != nil {
				return
			}
			pos += subsampleEntry.BytesOfProtectedData
		}
	} else {
		if err = cbc1Crypt(cm, data, key, iv); err != nil {
			return
		}
	}
	return
}

func cbc1Crypt(cm CryptMode, sample, key, iv []byte) (err error) {
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

	numToCryptByte := size & ^0xf
	mode.CryptBlocks(sample[:numToCryptByte], sample[:numToCryptByte])

	return
}

func EncryptSampleCbc1(data, key, iv []byte, subsampleEntries []mp4.SubsampleEntry) error {
	return CryptSampleCbc1(ModeEncrypt, data, key, iv, subsampleEntries)
}

func DecryptSampleCbc1(data, key, iv []byte, subsampleEntries []mp4.SubsampleEntry) error {
	return CryptSampleCbc1(ModeDecrypt, data, key, iv, subsampleEntries)
}
