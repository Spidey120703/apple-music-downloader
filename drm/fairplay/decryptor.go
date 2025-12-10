package fairplay

import (
	"downloader/mp4/cmaf"
	"downloader/mp4/mp4utils"

	"github.com/Spidey120703/go-mp4"
)

type Decryptor struct {
	mp4utils.DecryptContext
	AdamID string
}

func New(adamID string) *Decryptor {
	d := &Decryptor{}
	d.Sub = d
	d.AdamID = adamID
	return d
}

func (d *Decryptor) DecryptSegment(seg *cmaf.Segment, keyURIs [][]byte) (err error) {
	return d.DecryptContext.DecryptSegment(seg, keyURIs)
}

func (d *Decryptor) DecryptSample(schemeType string, samples []mp4utils.Sample, _ *mp4.Tenc, _ *mp4.Senc, keyURIs [][]byte) (err error) {
	switch schemeType {
	case "cbcs":
		err = DecryptSample(samples, d.AdamID, keyURIs)
		if err != nil {
			return
		}
	}
	return
}
