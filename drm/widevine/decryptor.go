package widevine

import (
	"downloader/mp4/mp4utils"

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
	iv := make([]byte, tenc.DefaultConstantIVSize)
	copy(iv, tenc.DefaultConstantIV)

	switch schemeType {
	case "cbcs":
		for i := range samples {
			subsampleEntries := senc.SampleEntries[i].SubsampleEntries
			err = DecryptSampleCbcs(samples[i].Data, keys[samples[i].SampleDescriptionIndex-1], iv, tenc, subsampleEntries)
			if err != nil {
				return
			}
		}
	}
	return
}
