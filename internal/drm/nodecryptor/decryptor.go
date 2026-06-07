package nodecryptor

import (
	"downloader/internal/media/mp4/mp4utils"

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

func (d *Decryptor) DecryptSample(string, []mp4utils.Sample, *mp4.Tenc, *mp4.Senc, [][]byte) error {
	return nil
}
