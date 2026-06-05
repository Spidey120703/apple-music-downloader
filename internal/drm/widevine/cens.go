package widevine

import (
	"crypto/aes"
	"crypto/cipher"

	"github.com/Spidey120703/go-mp4"
)

func censCrypt(stream cipher.Stream, sample []byte, numInCryptByte, numInSkipByte int) (err error) {
	var size = len(sample)

	if numInSkipByte == 0 {
		numToCryptByte := size & ^0xf
		stream.XORKeyStream(sample[:numToCryptByte], sample[:numToCryptByte])
		return
	}

	var pos int
	for size-pos >= numInCryptByte {
		stream.XORKeyStream(sample[pos:pos+numInCryptByte], sample[pos:pos+numInCryptByte])
		pos += numInCryptByte
		if size-pos < numInSkipByte {
			return
		}
		pos += numInSkipByte
	}

	return
}

func CryptSampleCens(data, key, iv []byte, subsampleEntries []mp4.SubsampleEntry, cryptByteBlock, skipByteBlock uint8) (err error) {
	if cryptByteBlock == 0 && skipByteBlock == 0 {
		return CryptSampleCenc(data, key, iv, subsampleEntries)
	}

	numInCryptByte := int(cryptByteBlock) << 4
	numInSkipByte := int(skipByteBlock) << 4

	var block cipher.Block
	if block, err = aes.NewCipher(key); err != nil {
		return
	}

	stream := cipher.NewCTR(block, iv)

	if len(subsampleEntries) != 0 {
		var pos uint32
		for _, subsampleEntry := range subsampleEntries {
			pos += uint32(subsampleEntry.BytesOfClearData)
			if subsampleEntry.BytesOfProtectedData == 0 {
				continue
			}
			if err = censCrypt(stream, data[pos:pos+subsampleEntry.BytesOfProtectedData], numInCryptByte, numInSkipByte); err != nil {
				return
			}
			pos += subsampleEntry.BytesOfProtectedData
		}
	} else {
		if err = censCrypt(stream, data, numInCryptByte, numInSkipByte); err != nil {
			return
		}
	}

	return
}
