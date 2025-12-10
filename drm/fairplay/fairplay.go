package fairplay

import (
	"downloader/barutils"
	"downloader/mp4/mp4utils"
	"downloader/utils"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"net"
)

const (
	SystemStringPrefix = "com.apple.fps"
	KeyFormatString    = "com.apple.streamingkeydelivery"
)

const (
	PrefetchKeyUri = "skd://itunes.apple.com/P000000000/s1/e1"
	DefaultId      = "0"
)

func DecryptSample(samples []mp4utils.Sample, adamID string, keyURIs [][]byte) (err error) {
	conn, err := net.Dial("tcp", "127.0.0.1:10020")
	if err != nil {
		return
	}
	defer utils.CloseQuietly(conn)

	var lastIndex uint32 = math.MaxUint32

	bar := barutils.NewProgressBar(int64(len(samples)), "Decrypting Samples:")

	for _, sample := range samples {
		if lastIndex != sample.SampleDescriptionIndex {
			if lastIndex != uint32(math.MaxUint32) {
				if _, err = conn.Write([]byte{0, 0, 0, 0}); err != nil {
					return
				}
			}
			keyURI := keyURIs[sample.SampleDescriptionIndex-1]
			id := adamID
			if string(keyURI) == PrefetchKeyUri {
				id = DefaultId
			}
			if len(id) == 0 {
				return errors.New("adam id is empty")
			}

			if _, err = conn.Write([]byte{byte(len(id))}); err != nil {
				return
			}
			if _, err = io.WriteString(conn, id); err != nil {
				return
			}

			if _, err = conn.Write([]byte{byte(len(keyURI))}); err != nil {
				return
			}
			if _, err = conn.Write(keyURI); err != nil {
				return
			}
		}
		lastIndex = sample.SampleDescriptionIndex

		if err = binary.Write(conn, binary.LittleEndian, uint32(len(sample.Data))); err != nil {
			return
		}
		if _, err = conn.Write(sample.Data); err != nil {
			return
		}
		if _, err = io.ReadFull(conn, sample.Data); err != nil {
			return
		}

		if err = bar.Add(1); err != nil {
			return
		}
	}
	_, _ = conn.Write([]byte{0, 0, 0, 0, 0})

	return
}
