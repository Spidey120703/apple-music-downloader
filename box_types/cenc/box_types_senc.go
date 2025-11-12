package cenc

import (
	"encoding/binary"
	"errors"
	"io"

	"github.com/abema/go-mp4"
)

func BoxTypeSenc() mp4.BoxType {
	return mp4.StrToBoxType("senc")
}

type SubsampleEntry struct {
	BytesOfClearData     uint16 `mp4:"0,size=16"`
	BytesOfProtectedData uint32 `mp4:"1,size=32"`
}

type SampleEntry struct {
	SubsampleCount   uint16           `mp4:"0,size=16"`
	SubsampleEntries []SubsampleEntry `mp4:"1,len=dynamic,size=dynamic"`
}

type Senc struct {
	mp4.FullBox     `mp4:"0,extend"`
	SampleCount     uint32        `mp4:"1,size=32"`
	SampleEntriesV2 []SampleEntry `mp4:"2,len=dynamic,size=dynamic"`
}

func (*Senc) GetType() mp4.BoxType {
	return BoxTypeSenc()
}

func Unmarshal(r io.ReadSeeker) (*Senc, error) {
	senc := &Senc{}

	var offset uint32

	// BoxSize
	var boxSize uint32
	if err := binary.Read(r, binary.BigEndian, &boxSize); err != nil {
		return nil, err
	}
	offset += 4

	// BoxType
	var boxType mp4.BoxType
	if _, err := io.ReadFull(r, boxType[:]); err != nil || boxType != BoxTypeSenc() {
		return nil, err
	}
	offset += 4

	// Version
	if err := binary.Read(r, binary.BigEndian, &senc.Version); err != nil {
		return nil, err
	}
	offset += 1

	// Flags
	if _, err := io.ReadFull(r, senc.Flags[:]); err != nil {
		return nil, err
	}
	offset += 3

	// SampleCount
	if err := binary.Read(r, binary.BigEndian, &senc.SampleCount); err != nil {
		return nil, err
	}
	offset += 4

	for i := 0; i < int(senc.SampleCount); i++ {
		sampleEntry := SampleEntry{}
		// SubsampleCount
		if err := binary.Read(r, binary.BigEndian, &sampleEntry.SubsampleCount); err != nil {
			return nil, err
		}
		offset += 2

		for j := 0; j < int(sampleEntry.SubsampleCount); j++ {
			subsampleEntry := SubsampleEntry{}

			// BytesOfClearData
			if err := binary.Read(r, binary.BigEndian, &subsampleEntry.BytesOfClearData); err != nil {
				return nil, err
			}
			offset += 2

			// BytesOfProtectedData
			if err := binary.Read(r, binary.BigEndian, &subsampleEntry.BytesOfProtectedData); err != nil {
				return nil, err
			}
			offset += 4

			sampleEntry.SubsampleEntries = append(sampleEntry.SubsampleEntries, subsampleEntry)
		}

		senc.SampleEntriesV2 = append(senc.SampleEntriesV2, sampleEntry)
	}

	if offset != boxSize {
		return nil, errors.New("invalid box size")
	}

	return senc, nil
}

//func (senc *Senc) BeforeUnmarshal(r io.ReadSeeker, size uint64, ctx mp4.Context) (uint64, bool, error) {
//	buf := make([]byte, 10)
//	if _, err := io.ReadFull(r, buf); err != nil {
//		return 0, false, err
//	}
//	if _, err := r.Seek(-int64(len(buf)), io.SeekCurrent); err != nil {
//		return 0, false, err
//	}
//	if buf[0]|buf[1]|buf[2]|buf[3] != 0x00 {
//		subsampleCount := binary.BigEndian.Uint16(buf[8:])
//		senc.subsampleCount = subsampleCount
//		return 0, true, nil
//	}
//	return 0, false, nil
//}

//func (senc *Senc) GetFieldSize(name string, ctx mp4.Context) uint {
//	switch name {
//	case "SampleEntriesV2":
//		return 16 + uint(senc.subsampleCount*48)
//	case "SubsampleEntries":
//		return 48
//	}
//	panic(fmt.Errorf("invalid name of dynamic-size field: boxType=trun fieldName=%s", name))
//}
//
//// GetFieldLength returns length of dynamic field
//func (senc *Senc) GetFieldLength(name string, ctx mp4.Context) uint {
//	switch name {
//	case "SampleEntriesV2":
//		return uint(senc.SampleCount)
//	case "SubsampleEntries":
//		return uint(senc.subsampleCount)
//	}
//	panic(fmt.Errorf("invalid name of dynamic-length field: boxType=trun fieldName=%s", name))
//}

func init() {
	mp4.AddBoxDef((*Senc)(nil))
}
