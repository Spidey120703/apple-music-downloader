package codec

import "math/bits"

/*************************** AVC ****************************/

var (
	AVCIndicatorAVC1 = CodecIndicator{0x61, 0x76, 0x63, 0x31} // avc1
	AVCIndicatorAVC2 = CodecIndicator{0x61, 0x76, 0x63, 0x32} // avc2
	AVCIndicatorAVC3 = CodecIndicator{0x61, 0x76, 0x63, 0x33} // avc3
	AVCIndicatorAVC4 = CodecIndicator{0x61, 0x76, 0x63, 0x34} // avc4
	AVCIndicatorSVC1 = CodecIndicator{0x73, 0x76, 0x63, 0x31} // svc1
	AVCIndicatorSVC2 = CodecIndicator{0x73, 0x76, 0x63, 0x32} // svc2
	AVCIndicatorMVC1 = CodecIndicator{0x6d, 0x76, 0x63, 0x31} // mvc1
	AVCIndicatorMVC2 = CodecIndicator{0x6d, 0x76, 0x63, 0x32} // mvc2
	AVCIndicatorMVC3 = CodecIndicator{0x6d, 0x76, 0x63, 0x33} // mvc3
	AVCIndicatorMVC4 = CodecIndicator{0x6d, 0x76, 0x63, 0x34} // mvc4
)

// AVCCodec - Advanced Video Coding specification (ISO/IEC 14496-10)
type AVCCodec struct {
	Codec

	// ProfileIndicator
	// profile_idc
	ProfileIndicator uint8

	// ConstraintSetFlags
	// the byte containing the constraint_set flags (currently constraint_set0_flag through
	// constraint_set5_flag, and reserved_zero_2bits)
	ConstraintSetFlags uint8

	// LevelIndicator
	// level_idc
	LevelIndicator uint8
}

func (c *AVCCodec) Initialize(str string) error {
	if err := c.Codec.Initialize(str); err != nil {
		return err
	}
	str = str[4:]

	if len(str) == 0 {
		return nil
	} else if len(str) != 7 {
		return ErrCodecInvalid
	}

	if ch := str[0]; ch != '.' {
		return ErrCodecInvalid
	}
	str = str[1:]

	assign(16, 8, &c.ProfileIndicator, str[:2])
	assign(16, 8, &c.ConstraintSetFlags, str[2:4])
	assign(16, 8, &c.LevelIndicator, str[4:])

	return nil
}

func (c *AVCCodec) isComparable() bool {
	return c.GetCodecIndicator() == AVCIndicatorAVC1
}

func (c *AVCCodec) Compare(i ICodec) int {
	if !c.isComparable() || !i.isComparable() {
		panic(ErrCodecUncomparable)
	}
	o := i.(*AVCCodec)

	if c.ProfileIndicator != o.ProfileIndicator {
		return int(c.ProfileIndicator) - int(o.ProfileIndicator)
	}

	if c.LevelIndicator != o.LevelIndicator {
		return int(c.LevelIndicator) - int(o.LevelIndicator)
	}

	cConstraints := bits.OnesCount8(c.ConstraintSetFlags)
	oConstraints := bits.OnesCount8(o.ConstraintSetFlags)
	return oConstraints - cConstraints
}
