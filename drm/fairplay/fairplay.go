package fairplay

import (
	"downloader/utils"
	"encoding/binary"
	"io"
	"math"
	"net"

	"github.com/schollz/progressbar/v3"
)

const KeyFormatFairPlay = "com.apple.streamingkeydelivery"

const (
	PrefetchKeyUri = "skd://itunes.apple.com/P000000000/s1/e1"
	DefaultId      = "0"
)

type Sample struct {
	Data                   []byte
	SampleDescriptionIndex uint32
	SampleDuration         uint32
}

type Chunk struct {
	Samples []Sample
}

func DecryptSample(samples []Sample, songID string, keys []string) (decryptedSamples [][]byte, err error) {
	conn, err := net.Dial("tcp", "127.0.0.1:10020")
	if err != nil {
		return
	}
	defer utils.CloseQuietly(conn)

	var lastIndex uint32 = math.MaxUint32

	bar := progressbar.Default(int64(len(samples)), "Decrypting")

	for _, sample := range samples {
		if lastIndex != sample.SampleDescriptionIndex {
			if lastIndex != uint32(math.MaxUint32) {
				_, err = conn.Write([]byte{0, 0, 0, 0})
				if err != nil {
					return
				}
			}
			keyUri := keys[sample.SampleDescriptionIndex]
			id := songID
			if keyUri == PrefetchKeyUri {
				id = DefaultId
			}

			_, err = conn.Write([]byte{byte(len(id))})
			if err != nil {
				return
			}
			_, err = io.WriteString(conn, id)
			if err != nil {
				return
			}

			_, err = conn.Write([]byte{byte(len(keyUri))})
			if err != nil {
				return
			}
			_, err = io.WriteString(conn, keyUri)
			if err != nil {
				return
			}
		}
		lastIndex = sample.SampleDescriptionIndex

		err = binary.Write(conn, binary.LittleEndian, uint32(len(sample.Data)))
		if err != nil {
			return
		}
		_, err = conn.Write(sample.Data)
		if err != nil {
			return
		}

		decrypted := make([]byte, len(sample.Data))
		_, err = io.ReadFull(conn, decrypted)
		if err != nil {
			return
		}

		decryptedSamples = append(decryptedSamples, decrypted)

		err = bar.Add(1)
		if err != nil {
			return
		}
	}
	_, _ = conn.Write([]byte{0, 0, 0, 0, 0})

	return
}
