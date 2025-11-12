package alac

import "github.com/abema/go-mp4"

/*
TODO: This project needs a new Context to distinguish between duplicated BoxType in SampleEntry.

	[github.com/abema/go-mp4/box_info.go@v1.4.1:11]
	type mp4.Context struct {
		...
		// ADD: UnderStsd represents whether current box is under the stsd box
		UnderStsd bool
	}
*/

// Alac - ALACSpecificConfig https://github.com/macosforge/alac/blob/master/codec/ALACAudioTypes.h#L162
type Alac struct {
	mp4.FullBox `mp4:"0,extend"`

	FrameLength       uint32 `mp4:"1,size=32"`
	CompatibleVersion uint8  `mp4:"2,size=8"`
	BitDepth          uint8  `mp4:"3,size=8"`
	Pb                uint8  `mp4:"4,size=8"`
	Mb                uint8  `mp4:"5,size=8"`
	Kb                uint8  `mp4:"6,size=8"`
	NumChannels       uint8  `mp4:"7,size=8"`
	MaxRun            uint16 `mp4:"8,size=16"`
	MaxFrameByte      uint32 `mp4:"9,size=32"`
	AvgBitRate        uint32 `mp4:"10,size=32"`
	SampleRate        uint32 `mp4:"11,size=32"`
}

func BoxTypeAlac() mp4.BoxType {
	return mp4.StrToBoxType("alac")
}

func (*Alac) GetType() mp4.BoxType {
	return BoxTypeAlac()
}

func init() {
	mp4.AddBoxDef((*Alac)(nil))
	mp4.AddAnyTypeBoxDefEx(&mp4.AudioSampleEntry{}, BoxTypeAlac(), func(context mp4.Context) bool {
		return context.UnderStsd
	})
}
