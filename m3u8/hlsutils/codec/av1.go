package codec

import "strings"

var (
	AV1Indicator = CodecIndicator{0x61, 0x76, 0x30, 0x31} // av01
)

var AV1ProfileMap = map[uint8]string{
	0: "Main",
	1: "High",
	2: "Professional",
}

var AV1TierMap = map[uint8]uint8{
	'M': 0,
	'H': 1,
}

var AV1SeqTierMap = map[uint8]string{
	0: "Main tier",
	1: "High tier",
}

var ChromaSubsamplingFormatMap = map[uint8]string{
	0: "YUV 4:4:4",
	4: "YUV 4:2:2",
	6: "YUV 4:2:0",
	7: "YUV 4:0:0 (Monochrome)",
}

// AV1Codec
// https://aomediacodec.github.io/av1-isobmff/#codecsparam
type AV1Codec struct {
	SampleEntry CodecIndicator
	Profile     uint8
	LevelTier   struct {
		Level struct {
			X uint8
			Y uint8
		}
		SeqTier uint8
	}
	BitDepth          uint8
	Monochrome        bool
	ChromaSubsampling struct {
		SubsamplingX         bool
		SubsamplingY         bool
		ChromaSamplePosition bool
	}
	ColorPrimaries          uint8
	TransferCharacteristics uint8
	MatrixCoefficients      uint8
	VideoFullRangeFlag      bool
}

func (c *AV1Codec) Initialize(str string) error {
	sampleEntry, err := StrToCodecIndicator(str[:4])
	if err != nil {
		return err
	}
	c.SampleEntry = sampleEntry
	str = str[4:]

	if len(str) == 0 {
		return ErrCodecInvalid
	}

	if ch := str[0]; ch != '.' {
		return ErrCodecInvalid
	}
	str = str[1:]

	parts := strings.Split(str, ".")

	if len(parts) != 3 && len(parts) != 9 {
		return ErrCodecInvalid
	}

	{
		profile := parts[0]
		if len(profile) != 1 || (profile[0] != '0' && profile[0] != '1' && profile[0] != '2') {
			return ErrCodecInvalid
		}
		c.Profile = profile[0] - '0'
	}

	{
		levelTier := parts[1]
		if len(levelTier) != 3 {
			return ErrCodecInvalid
		}

		var level uint8
		assign(10, 8, &level, levelTier[:2])

		c.LevelTier.Level.X = 2 + (level >> 2)
		c.LevelTier.Level.Y = level & 3

		tier, found := AV1TierMap[levelTier[2]]
		if !found {
			return ErrCodecInvalid
		}
		c.LevelTier.SeqTier = tier
	}

	{
		var bitDepth uint8
		assign(10, 8, &bitDepth, parts[2])

		if bitDepth != 8 && bitDepth != 10 && bitDepth != 12 {
			return ErrCodecInvalid
		}

		c.BitDepth = bitDepth
	}

	if len(parts) == 3 {
		return nil
	}

	{
		monochrome := parts[3]
		if len(monochrome) != 1 {
			return ErrCodecInvalid
		}
		c.Monochrome = monochrome[0] != '0'
	}

	{
		ccc := parts[4]
		if len(ccc) != 3 &&
			(ccc[0] != '0' && ccc[0] != '1') ||
			(ccc[1] != '0' && ccc[1] != '1') ||
			(ccc[2] != '0' && ccc[2] != '1') {
			return ErrCodecInvalid
		}
		c.ChromaSubsampling.SubsamplingX = ccc[0] == '1'
		c.ChromaSubsampling.SubsamplingY = ccc[1] == '1'
		if ccc[0] == '1' && ccc[1] == '1' {
			c.ChromaSubsampling.ChromaSamplePosition = ccc[2] == '1'
		}
	}

	assign(10, 8, &c.ColorPrimaries, parts[5])
	assign(10, 8, &c.TransferCharacteristics, parts[6])
	assign(10, 8, &c.MatrixCoefficients, parts[7])

	{
		flag := parts[8]
		if len(flag) != 1 && flag[0] != '0' && flag[0] != '1' {
			return ErrCodecInvalid
		}
		c.VideoFullRangeFlag = flag[0] == '1'
	}

	return nil
}

func (c *AV1Codec) GetCodecIndicator() CodecIndicator {
	return c.SampleEntry
}

func (c *AV1Codec) GetProfileName() string {
	return AV1ProfileMap[c.Profile]
}

func (c *AV1Codec) GetTierDescription() string {
	return AV1SeqTierMap[c.LevelTier.SeqTier]
}

func (c *AV1Codec) GetChromaSubsampling() (ccc uint8) {
	if c.ChromaSubsampling.SubsamplingX {
		ccc = 7
	}
	if c.ChromaSubsampling.SubsamplingY {
		ccc |= 4
	}
	if c.ChromaSubsampling.ChromaSamplePosition {
		ccc |= 1
	}
	return
}

func (c *AV1Codec) GetChromaSubsamplingFormat() string {
	return ChromaSubsamplingFormatMap[c.GetChromaSubsampling()]
}

func (c *AV1Codec) isComparable() bool {
	return c.GetCodecIndicator() == AV1Indicator
}

func (c *AV1Codec) Compare(i ICodec) int {
	if !c.isComparable() || !i.isComparable() {
		panic(ErrCodecUncomparable)
	}
	o := i.(*AV1Codec)

	if c.Profile != o.Profile {
		return int(c.Profile) - int(o.Profile)
	}

	if c.LevelTier.SeqTier != o.LevelTier.SeqTier {
		return int(c.LevelTier.SeqTier) - int(o.LevelTier.SeqTier)
	}

	if c.LevelTier.Level.X != o.LevelTier.Level.X {
		return int(c.LevelTier.Level.X) - int(o.LevelTier.Level.X)
	}
	if c.LevelTier.Level.Y != o.LevelTier.Level.Y {
		return int(c.LevelTier.Level.Y) - int(o.LevelTier.Level.Y)
	}

	if c.BitDepth != o.BitDepth {
		return int(c.BitDepth) - int(o.BitDepth)
	}

	if c.GetChromaSubsampling() != o.GetChromaSubsampling() {
		return int(o.GetChromaSubsampling()) - int(c.GetChromaSubsampling())
	}

	return 0
}
