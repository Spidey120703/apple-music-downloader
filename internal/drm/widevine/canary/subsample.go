package canary

import "github.com/Spidey120703/go-mp4"

type SubsampleCryptor struct {
	cph ICipher
}

func (c *SubsampleCryptor) Crypt(m CipherDirection, data, key, iv []byte, subsampleEntries []mp4.SubsampleEntry, _, _ uint8) (err error) {
	if err = c.cph.Initialize(m, key, iv); err != nil {
		return
	}

	if len(subsampleEntries) != 0 {
		var pos uint32
		for _, subsampleEntry := range subsampleEntries {
			pos += uint32(subsampleEntry.BytesOfClearData)
			if subsampleEntry.BytesOfProtectedData == 0 {
				continue
			}
			c.cph.CryptInplace(data[pos : pos+subsampleEntry.BytesOfProtectedData])
			pos += subsampleEntry.BytesOfProtectedData
		}
	} else {
		c.cph.CryptInplace(data)
	}
	return
}

func NewSubsampleCryptor(cph ICipher) ICryptOperator {
	return &CryptOperator{&SubsampleCryptor{cph}}
}
