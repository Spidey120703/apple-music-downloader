package lagacy

import (
	"crypto/aes"
	"crypto/cipher"

	"github.com/Spidey120703/go-mp4"
)

func CryptSampleCenc(data, key, iv []byte, subsampleEntries []mp4.SubsampleEntry) (err error) {
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
			stream.XORKeyStream(data[pos:pos+subsampleEntry.BytesOfProtectedData], data[pos:pos+subsampleEntry.BytesOfProtectedData])
			pos += subsampleEntry.BytesOfProtectedData
		}
	} else {
		stream.XORKeyStream(data, data)
	}

	return
}
