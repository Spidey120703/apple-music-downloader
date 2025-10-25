package avc

import "github.com/abema/go-mp4"

func BoxTypeChrm() mp4.BoxType {
	return mp4.StrToBoxType("chrm")
}

type Chrm struct {
	mp4.Box
	X uint8 `mp4:"0,size=8"`
	Y uint8 `mp4:"1,size=8"`
}

func (*Chrm) GetType() mp4.BoxType {
	return BoxTypeChrm()
}

func init() {
	mp4.AddBoxDef((*Chrm)(nil))
}
