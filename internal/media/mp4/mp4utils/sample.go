package mp4utils

import (
	"bytes"
	"downloader/internal/media/mp4/bitstruct"
	"downloader/internal/media/mp4/cmaf"
	"encoding/binary"

	"github.com/Spidey120703/go-mp4"
)

type SampleFlags struct {
	bitstruct.BaseFieldObject
	Reserved                  byte   `bit:"0,size=4,const=0"`
	IsLeading                 uint8  `bit:"1,size=2"`
	SampleDependsOn           uint8  `bit:"2,size=2"`
	SampleIsDependedOn        uint8  `bit:"3,size=2"`
	SampleHasRedundancy       uint8  `bit:"4,size=2"`
	SamplePaddingValue        byte   `bit:"5,size=3"`
	SampleIsNonSyncSample     byte   `bit:"6,size=1"`
	SampleDegradationPriority uint16 `bit:"7,size=16"`
}

func UnmarshalSampleFlags(flags uint32) (sampleFlags SampleFlags, err error) {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, flags)
	if _, err = bitstruct.Unmarshal(bytes.NewReader(buf), 4, &sampleFlags); err != nil {
		return
	}
	return
}

type Sample struct {
	Data                          []byte
	DataOffset                    uint64
	SampleDescriptionIndex        uint32
	SampleDuration                uint32
	SampleSize                    uint32
	SampleFlags                   uint32
	SampleCompositionTimeOffsetV0 uint32
	SampleCompositionTimeOffsetV1 int32
	Version                       uint8
}

func GetFullSamples(traf cmaf.TrackFragmentBox, mdat *mp4.Mdat, trex *mp4.Trex) (samples []Sample) {
	//data := mdat.Data
	for _, trun := range traf.Trun {
		dataOffset := func() uint64 {
			if trun.CheckFlag(0x1) {
				return uint64(trun.DataOffset)
			} else if traf.Tfhd.CheckFlag(0x1) {
				return traf.Tfhd.BaseDataOffset
			} else {
				return traf.Node.Parent.Info.Size + 8
			}
		}()
		for _, entry := range trun.Entries {
			sample := Sample{
				DataOffset: dataOffset,
				SampleDescriptionIndex: func() uint32 {
					if traf.Tfhd.CheckFlag(0x2) {
						return traf.Tfhd.SampleDescriptionIndex
					} else {
						return trex.DefaultSampleDescriptionIndex
					}
				}(),
				SampleDuration: func() uint32 {
					if trun.CheckFlag(0x100) {
						return entry.SampleDuration
					} else if traf.Tfhd.CheckFlag(0x8) {
						return traf.Tfhd.DefaultSampleDuration
					} else {
						return trex.DefaultSampleDuration
					}
				}(),
				SampleSize: func() uint32 {
					if trun.CheckFlag(0x200) {
						return entry.SampleSize
					} else if traf.Tfhd.CheckFlag(0x10) {
						return traf.Tfhd.DefaultSampleSize
					} else {
						return trex.DefaultSampleSize
					}
				}(),
				SampleFlags: func() uint32 {
					if trun.CheckFlag(0x4) {
						return trun.FirstSampleFlags
					} else if trun.CheckFlag(0x400) {
						return entry.SampleFlags
					} else if traf.Tfhd.CheckFlag(0x20) {
						return traf.Tfhd.DefaultSampleFlags
					} else {
						return trex.DefaultSampleFlags
					}
				}(),
			}
			if trun.CheckFlag(0x800) {
				sample.Version = trun.GetVersion()
				if trun.GetVersion() == 1 {
					sample.SampleCompositionTimeOffsetV1 = entry.SampleCompositionTimeOffsetV1
				} else {
					sample.SampleCompositionTimeOffsetV0 = entry.SampleCompositionTimeOffsetV0
				}
			}

			//sample.Data = data[:sample.SampleSize]
			//data = data[sample.SampleSize:]

			offset := sample.DataOffset - traf.Node.Parent.Info.Size - 8
			sample.Data = mdat.Data[offset : offset+uint64(sample.SampleSize)]
			samples = append(samples, sample)
			dataOffset += uint64(sample.SampleSize)
		}
	}
	return
}
