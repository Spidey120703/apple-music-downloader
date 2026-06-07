package canary

import "github.com/Spidey120703/go-mp4"

type SubsamplePatternCryptor struct {
	cph ICipher
}

func (c *SubsamplePatternCryptor) cryptPattern(data []byte, numInCryptByte, numInSkipByte int) (offset int, err error) {
	c.cph.Reset()

	var size = len(data)

	if numInSkipByte == 0 {
		numToCryptByte := size & ^0xf
		c.cph.CryptInplace(data[:numToCryptByte])
		return
	}

	var pos int
	for size-pos >= numInCryptByte {
		c.cph.CryptInplace(data[pos : pos+numInCryptByte])
		pos += numInCryptByte

		if size-pos < numInSkipByte {
			if c.cph.SupportsPartialBlocks() {
				offset = numInSkipByte - size + pos
			}
			return
		}
		pos += numInSkipByte
	}

	if c.cph.SupportsPartialBlocks() {
		offset = min(pos-size, 0)
	}

	return
}

func (c *SubsamplePatternCryptor) Crypt(m CipherDirection, data, key, iv []byte, subsampleEntries []mp4.SubsampleEntry, cryptByteBlock, skipByteBlock uint8) (err error) {
	numInCryptByte := int(cryptByteBlock) << 4
	numInSkipByte := int(skipByteBlock) << 4

	if err = c.cph.Initialize(m, key, iv); err != nil {
		return
	}

	if len(subsampleEntries) != 0 {
		var pos uint32

		// offset tracks pattern encryption continuation across subsample boundaries
		var offset int

		for _, subsampleEntry := range subsampleEntries {
			pos += uint32(subsampleEntry.BytesOfClearData)

			if subsampleEntry.BytesOfProtectedData == 0 {
				continue
			}

			if int(pos)+offset >= len(data) {
				break
			}

			// offset can propagate across subsample boundaries when a pattern encryption
			// segment spans multiple subsamples. In such cases, the remaining offset may
			// exceed the current protected data size and must be carried to the next one.
			if offset > int(subsampleEntry.BytesOfProtectedData) {
				offset -= int(subsampleEntry.BytesOfProtectedData)
				continue
			}

			if offset, err = c.cryptPattern(data[int(pos)+offset:pos+subsampleEntry.BytesOfProtectedData], numInCryptByte, numInSkipByte); err != nil {
				return
			}
			pos += subsampleEntry.BytesOfProtectedData
		}
	} else {
		if _, err = c.cryptPattern(data, numInCryptByte, numInSkipByte); err != nil {
			return
		}
	}
	return
}

func NewSubsamplePatternCryptor(cph ICipher) ICryptOperator {
	return &CryptOperator{&SubsamplePatternCryptor{cph}}
}
