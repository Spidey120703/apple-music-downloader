package widevine

import (
	"downloader/internal/drm/widevine/canary"
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
	d.ISampleDecryptor = d
	return d
}

func (*Decryptor) DecryptSample(schemeType string, samples []mp4utils.Sample, tenc *mp4.Tenc, senc *mp4.Senc, keys [][]byte) (err error) {
	switch schemeType {
	case "cbcs", "cbc1", "cenc", "cens":
	default:
		return fmt.Errorf("scheme type '%s' is unsupported", schemeType)
	}

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

	var cph canary.ICipher
	//goland:noinspection GoDfaConstantCondition
	switch schemeType {
	case "cbcs", "cbc1":
		cph = &canary.CBC{}
	case "cenc", "cens":
		cph = &canary.CTR{}
	}

	var cryptor canary.ICryptOperator
	//goland:noinspection GoDfaConstantCondition
	switch schemeType {
	case "cbcs", "cens":
		cryptor = canary.NewSubsamplePatternCryptor(cph)
	case "cenc", "cbc1":
		cryptor = canary.NewSubsampleCryptor(cph)
	}

	for i := range samples {
		subsampleEntries := senc.SampleEntries[i].SubsampleEntries
		if tenc.DefaultPerSampleIVSize > 0 {
			copy(iv, senc.SampleEntries[i].InitializationVector)
		}
		if err = cryptor.Decrypt(samples[i].Data, keys[samples[i].SampleDescriptionIndex-1], iv, subsampleEntries, cryptByteBlock, skipByteBlock); err != nil {
			return
		}
	}

	return
}
