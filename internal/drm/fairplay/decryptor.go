package fairplay

import (
	"downloader/internal/media/mp4/mp4utils"
	"errors"

	"github.com/Spidey120703/go-mp4"
)

type Decryptor struct {
	mp4utils.DecryptContext
	AdamID string
}

func New(adamID string) *Decryptor {
	d := &Decryptor{
		AdamID: adamID,
	}
	d.ISampleDecryptor = d
	return d
}

func (d *Decryptor) DecryptSample(schemeType string, samples []mp4utils.Sample, _ *mp4.Tenc, _ *mp4.Senc, keyURIs [][]byte) (err error) {
	switch schemeType {
	case "cbcs":
		err = DecryptSample(samples, d.AdamID, keyURIs)
		if err != nil {
			return
		}
	default:
		return errors.New("scheme is unsupported")
	}
	return
}
