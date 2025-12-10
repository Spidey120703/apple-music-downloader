package codec

import "strings"

/*************************** HEVC ****************************/

var (
	HEVCIndicatorHEV1 = CodecIndicator{0x68, 0x65, 0x76, 0x31} // hev1
	HEVCIndicatorHVC1 = CodecIndicator{0x68, 0x76, 0x63, 0x31} // hvc1
)

var HEVCGeneralProfileSpaceMap = map[uint8]uint8{
	'A': 1,
	'B': 2,
	'C': 3,
}

var HEVCGeneralTierFlagMap = map[uint8]uint8{
	'L': 0,
	'H': 1,
}

// HEVCCodec - High Efficiency Video Coding specification (ISO/IEC 23008-2)
type HEVCCodec struct {
	Codec

	// GeneralProfile
	// the general_profile_space, encoded as no character (general_profile_space == 0), or
	// 'A', 'B', 'C' for general_profile_space 1, 2, 3, followed by the general_profile_idc encoded as a
	// decimal number;
	GeneralProfile struct {

		// GeneralProfileSpace
		// general_profile_space
		GeneralProfileSpace uint8

		// GeneralProfileIndicator
		// general_profile_idc
		GeneralProfileIndicator uint8
	}

	// GeneralProfileCompatibilityFlags
	// the 32 bits of the general_profile_compatibility_flags, but in reverse bit order, i.e. with
	// general_profile_compatibility_flag[31] as the most significant bit, followed by
	// general_profile_compatibility_flag[30], and down to general_profile_compatibility_flag[0]
	// as the least significant bit, where general_profile_compatibility_flag[i] for
	// i in the range of 0 to 31, inclusive, are specified in ISO/IEC 23008-2, encoded in hexadecimal (leading
	// zeroes may be omitted);
	GeneralProfileCompatibilityFlags uint32

	// GeneralTierLevel
	// the general_tier_flag, encoded as "L" (general_tier_flag == 0) or "H" (general_tier_flag
	// == 1), followed by the general_level_idc, encoded as a decimal number;
	GeneralTierLevel struct {

		// GeneralTierFlag
		// general_tier_flag
		GeneralTierFlag uint8

		// GeneralLevelIndicator
		// general_level_idc
		GeneralLevelIndicator uint8
	}

	// ConstraintFlags
	// each of the 6 bytes of the constraint flags, starting from the byte containing the
	// general_progressive_source_flag, each encoded as a  hexadecimal number, and the encoding of each
	// byte separated by a period; trailing bytes that are zero may be omitted.
	ConstraintFlags [6]uint8
}

func (c *HEVCCodec) Initialize(str string) error {
	if err := c.Codec.Initialize(str); err != nil {
		return err
	}
	str = str[4:]

	if len(str) == 0 {
		return nil
	}

	if ch := str[0]; ch != '.' {
		return ErrCodecInvalid
	}
	str = str[1:]

	parts := strings.Split(str, ".")
	if len(parts) < 3 || len(parts) > 6 {
		return ErrCodecInvalid
	}

	{
		profileStr := parts[0]
		if len(profileStr) == 0 {
			return ErrCodecInvalid
		} else if len(profileStr) == 1 {
			c.GeneralProfile.GeneralProfileSpace = 0
		} else if len(profileStr) == 2 {
			space, found := HEVCGeneralProfileSpaceMap[profileStr[0]]
			if !found {
				return ErrCodecInvalid
			}
			c.GeneralProfile.GeneralProfileSpace = space
			profileStr = profileStr[1:]
		}
		assign(10, 8, &c.GeneralProfile.GeneralProfileIndicator, profileStr)
	}

	assign(16, 32, &c.GeneralProfileCompatibilityFlags, parts[1])

	{
		tierLevelStr := parts[2]
		if len(tierLevelStr) < 2 {
			return ErrCodecInvalid
		}

		tier, found := HEVCGeneralTierFlagMap[tierLevelStr[0]]
		if !found {
			return ErrCodecInvalid
		}
		c.GeneralTierLevel.GeneralTierFlag = tier

		assign(10, 8, &c.GeneralTierLevel.GeneralLevelIndicator, tierLevelStr[1:])
	}

	for idx, part := range parts[3:] {
		assign(16, 8, &c.ConstraintFlags[idx], part)
	}

	return nil
}

func (c *HEVCCodec) isComparable() bool {
	return c.GetCodecIndicator() == HEVCIndicatorHVC1 || c.GetCodecIndicator() == HEVCIndicatorHEV1
}

func (c *HEVCCodec) Compare(i ICodec) int {
	if !c.isComparable() || !i.isComparable() {
		panic(ErrCodecUncomparable)
	}
	o := i.(*HEVCCodec)

	if c.GeneralProfile.GeneralProfileIndicator != o.GeneralProfile.GeneralProfileIndicator {
		return int(c.GeneralProfile.GeneralProfileIndicator) - int(o.GeneralProfile.GeneralProfileIndicator)
	}

	if c.GeneralTierLevel.GeneralTierFlag != o.GeneralTierLevel.GeneralTierFlag {
		return int(c.GeneralTierLevel.GeneralTierFlag) - int(o.GeneralTierLevel.GeneralTierFlag)
	}

	if c.GeneralTierLevel.GeneralLevelIndicator != o.GeneralTierLevel.GeneralLevelIndicator {
		return int(c.GeneralTierLevel.GeneralLevelIndicator) - int(o.GeneralTierLevel.GeneralLevelIndicator)
	}

	if c.GeneralProfileCompatibilityFlags != o.GeneralProfileCompatibilityFlags {
		return int(c.GeneralProfileCompatibilityFlags) - int(o.GeneralProfileCompatibilityFlags)
	}

	return 0
}
