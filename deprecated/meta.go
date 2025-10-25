package deprecated

import (
	"bytes"
	"encoding/binary"
	"io"

	mp4ff "github.com/Eyevinn/mp4ff/mp4"
	"github.com/abema/go-mp4"
)

func TestTlou(reader io.ReadSeeker) {
	// Box Path: moov/trak/udta
	box, err := mp4.ExtractBox(reader, nil, mp4.BoxPath{
		mp4.BoxTypeMoov(),
		mp4.BoxTypeTrak(),
		mp4.BoxTypeUdta(),
	})
	if err != nil && len(box) != 1 {
		return
	}

	var offset = int64(box[0].Offset + 8)
	_, err = reader.Seek(offset, io.SeekStart)
	if err != nil {
		return
	}

	var size uint32
	err = binary.Read(reader, binary.BigEndian, &size)
	if err != nil {
		return
	}

	_, err = reader.Seek(offset, io.SeekStart)
	if err != nil {
		return
	}

	var buf = make([]byte, size)
	err = binary.Read(reader, binary.BigEndian, &buf)
	if err != nil {
		return
	}

	buffer := bytes.NewReader(buf)
	ludtBox, err := mp4ff.DecodeBox(0, buffer)
	if err != nil {
		return
	}
	ludt := ludtBox.(*mp4ff.LudtBox)
	println(ludt.Children[0].Type())
	println(ludt.Loudness[0].Type())
	println(ludt.Loudness[0].LoudnessBases[0].EQSetID)
	println(ludt.Loudness[0].LoudnessBases[0].DownmixID)
	println(ludt.Loudness[0].LoudnessBases[0].DRCSetID)
	println(ludt.Loudness[0].LoudnessBases[0].BsSamplePeakLevel)
	println(ludt.Loudness[0].LoudnessBases[0].BsTruePeakLevel)
	println(ludt.Loudness[0].LoudnessBases[0].MeasurementSystemForTP)
	println(ludt.Loudness[0].LoudnessBases[0].ReliabilityForTP)
	println(ludt.Children[1].Type())
	println(ludt.AlbumLoudness[0].Type())
	println(ludt.AlbumLoudness[0].LoudnessBases[0].EQSetID)
	println(ludt.AlbumLoudness[0].LoudnessBases[0].DownmixID)
	println(ludt.AlbumLoudness[0].LoudnessBases[0].DRCSetID)
	println(ludt.AlbumLoudness[0].LoudnessBases[0].BsSamplePeakLevel)
	println(ludt.AlbumLoudness[0].LoudnessBases[0].BsTruePeakLevel)
	println(ludt.AlbumLoudness[0].LoudnessBases[0].MeasurementSystemForTP)
	println(ludt.AlbumLoudness[0].LoudnessBases[0].ReliabilityForTP)
}
