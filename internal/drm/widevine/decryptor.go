package widevine

import (
	"downloader/internal/media/mp4/mp4utils"
	"errors"

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

func (*Decryptor) DecryptSample(schemeType string, samples []mp4utils.Sample, tenc *mp4.Tenc, senc *mp4.Senc, keys [][]byte) (err error) {
	var constantIVSize = tenc.DefaultConstantIVSize
	var constantIV = tenc.DefaultConstantIV

	if constantIVSize == 0 {
		constantIVSize = uint8(len(keys[0]))
	}

	iv := make([]byte, constantIVSize)
	copy(iv, constantIV)

	switch schemeType {
	case "cbcs":
		for i := range samples {
			subsampleEntries := senc.SampleEntries[i].SubsampleEntries
			err = DecryptSampleCbcs(samples[i].Data, keys[samples[i].SampleDescriptionIndex-1], iv, tenc, subsampleEntries)
			if err != nil {
				return
			}
		}
	case "cenc":
		for i := range samples {
			subsampleEntries := senc.SampleEntries[i].SubsampleEntries
			if tenc.DefaultPerSampleIVSize != 0 {
				copy(iv, senc.SampleEntries[i].InitializationVector)
			}
			err = CryptSampleCenc(samples[i].Data, keys[samples[i].SampleDescriptionIndex-1], iv, subsampleEntries)
			if err != nil {
				return
			}
		}
	default:
		return errors.New("scheme is unsupported")
	}

	return
}
