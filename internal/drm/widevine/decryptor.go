package widevine

import (
	"downloader/internal/media/mp4/mp4utils"
	"fmt"
	"math"

	"github.com/Spidey120703/go-mp4"
)

type Decryptor struct {
	mp4utils.DecryptContext
}

func New() *Decryptor {
	d := &Decryptor{}
	d.Sub = d
	return d
}

type DecryptSampleFunc func(data, key, iv []byte, subsampleEntries []mp4.SubsampleEntry, cryptByteBlock, skipByteBlock uint8) (err error)

func decryptSample(decryptSampleFunc DecryptSampleFunc, samples []mp4utils.Sample, keys [][]byte, iv []byte, senc *mp4.Senc, hasPerSampleIV bool, cryptByteBlock, skipByteBlock uint8) (err error) {
	for i := range samples {
		subsampleEntries := senc.SampleEntries[i].SubsampleEntries
		if hasPerSampleIV {
			copy(iv, senc.SampleEntries[i].InitializationVector)
		}
		if err = decryptSampleFunc(samples[i].Data, keys[samples[i].SampleDescriptionIndex-1], iv, subsampleEntries, cryptByteBlock, skipByteBlock); err != nil {
			return
		}
	}
	return
}

func (*Decryptor) DecryptSample(schemeType string, samples []mp4utils.Sample, tenc *mp4.Tenc, senc *mp4.Senc, keys [][]byte) (err error) {
	var cryptByteBlock = tenc.DefaultCryptByteBlock
	var skipByteBlock = tenc.DefaultSkipByteBlock

	var ivSize = max(tenc.DefaultConstantIVSize, tenc.DefaultPerSampleIVSize, uint8(len(tenc.DefaultKID)))
	var defaultConstantIV = tenc.DefaultConstantIV

	for _, key := range keys {
		ivSize = max(ivSize, uint8(len(key)))
	}

	ivSize = 1 << uint8(math.Ceil(math.Log2(float64(ivSize))))

	iv := make([]byte, ivSize)
	copy(iv, defaultConstantIV)

	switch schemeType {
	case "cbc1", "cbcs":
		if err = decryptSample(DecryptSampleCbcs, samples, keys, iv, senc, tenc.DefaultPerSampleIVSize != 0, cryptByteBlock, skipByteBlock); err != nil {
			return
		}
	case "cenc", "cens":
		if err = decryptSample(CryptSampleCens, samples, keys, iv, senc, tenc.DefaultPerSampleIVSize != 0, cryptByteBlock, skipByteBlock); err != nil {
			return
		}
	default:
		return fmt.Errorf("scheme type '%s' is unsupported", schemeType)
	}

	return
}
