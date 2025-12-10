package codec

import (
	"errors"
	"strconv"
)

// "codecs" parameter (ISO/IEC 14496-15, RFC 6381)

var ErrCodecInvalid = errors.New("codec invalid")
var ErrCodecUncomparable = errors.New("codec uncomparable")
var ErrCodecUnsupported = errors.New("codec unsupported")

type CodecIndicator [4]byte

func StrToCodecIndicator(str string) (CodecIndicator, error) {
	if len(str) != 4 {
		return [4]byte{}, ErrCodecInvalid
	}
	return CodecIndicator{str[0], str[1], str[2], str[3]}, nil
}

type ICodec interface {
	Initialize(string) error
	GetCodecIndicator() CodecIndicator
	Compare(ICodec) int
	isComparable() bool
}

type Codec struct {
	CodecIndicator CodecIndicator
}

func (c *Codec) Initialize(str string) error {
	if len(str) < 4 {
		return ErrCodecInvalid
	}
	indicator, err := StrToCodecIndicator(str[:4])
	if err != nil {
		return err
	}
	c.CodecIndicator = indicator
	return nil
}

func (c *Codec) GetCodecIndicator() CodecIndicator {
	return c.CodecIndicator
}

func (c *Codec) Compare(ICodec) int {
	return 0
}

func (c *Codec) isComparable() bool {
	return true
}

type intType interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

func assign[T intType](base int, bitSize int, dest *T, str string) {
	dec, err := strconv.ParseUint(str, base, bitSize)
	if err != nil {
		panic(err)
	}
	*dest = T(dec)
}

type Family int

const (
	FamilyMPEG4 Family = iota
	FamilyAVC
	FamilyHEVC
	FamilyAV1
	FamilyAtmos
	FamilyALAC
)

func (c CodecIndicator) GetCodecFamily() Family {
	switch c {
	case
		AVCIndicatorAVC1,
		AVCIndicatorAVC2,
		AVCIndicatorAVC3,
		AVCIndicatorAVC4,
		AVCIndicatorSVC1,
		AVCIndicatorSVC2,
		AVCIndicatorMVC1,
		AVCIndicatorMVC2,
		AVCIndicatorMVC3,
		AVCIndicatorMVC4:
		return FamilyAVC
	case
		HEVCIndicatorHEV1,
		HEVCIndicatorHVC1:
		return FamilyHEVC
	case AV1Indicator:
		return FamilyAV1
	case
		MPEG4IndicatorMP4A,
		MPEG4IndicatorMP4V:
		return FamilyMPEG4
	case EC3Indicator:
		return FamilyAtmos
	case ALACIndicator:
		return FamilyALAC
	default:
		panic(ErrCodecUnsupported)
	}
}

func Initialize(str string) (c ICodec, err error) {
	if len(str) < 4 {
		return nil, ErrCodecInvalid
	}
	indicator, err := StrToCodecIndicator(str[:4])
	if err != nil {
		return nil, err
	}
	switch indicator.GetCodecFamily() {
	case FamilyMPEG4:
		c = &MP4Codec{}
	case FamilyAVC:
		c = &AVCCodec{}
	case FamilyHEVC:
		c = &HEVCCodec{}
	case FamilyAV1:
		c = &AV1Codec{}
	case FamilyAtmos:
		c = &EC3Codec{}
	case FamilyALAC:
		c = &ALACCodec{}
	default:
		panic(ErrCodecUnsupported)
	}

	if err = c.Initialize(str); err != nil {
		return
	}
	return
}

func Less(a ICodec, b ICodec) bool {
	if a.GetCodecIndicator() != b.GetCodecIndicator() {
		return a.GetCodecIndicator().GetCodecFamily()-b.GetCodecIndicator().GetCodecFamily() < 0
	}
	return a.Compare(b) < 0
}

func LessStr(a string, b string) (bool, error) {
	aa, err := Initialize(a)
	if err != nil {
		return false, err
	}

	bb, err := Initialize(b)
	if err != nil {
		return false, err
	}

	return Less(aa, bb), nil
}
