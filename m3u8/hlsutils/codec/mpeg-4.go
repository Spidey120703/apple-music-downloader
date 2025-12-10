package codec

import "strings"

var (
	MPEG4IndicatorMP4A = CodecIndicator{0x6d, 0x70, 0x34, 0x61} // mp4a
	MPEG4IndicatorMP4V = CodecIndicator{0x6d, 0x70, 0x34, 0x76} // mp4v
)

// ObjectTypeIndicationMap - OTI Object Type Indications
// https://mp4ra.org/registered-types/object-types
var ObjectTypeIndicationMap = map[uint8]string{
	0x00: "Forbidden",
	0x01: "Systems ISO/IEC 14496-1 (1)",
	0x02: "Systems ISO/IEC 14496-1 (2)",
	0x03: "Interaction Stream",
	0x04: "Extended BIFS (8)",
	0x05: "AFX Stream (9)",
	0x06: "Font Data Stream",
	0x07: "Synthetised Texture",
	0x08: "Text Stream",
	0x09: "LASeR Stream",
	0x0A: "Simple Aggregation Format (SAF) Stream",
	0x20: "Visual ISO/IEC 14496-2 (3)",
	0x21: "Visual ITU-T Recommendation H.264 | ISO/IEC 14496-10 (7)",
	0x22: "Parameter Sets for ITU-T Recommendation H.264 | ISO/IEC 14496-10 (7)",
	0x23: "Visual ISO/IEC 23008-2 | ITU-T Recommendation H.265",
	0x40: "Audio ISO/IEC 14496-3 (4)", // AAC
	0x60: "Visual ISO/IEC 13818-2 Simple Profile",
	0x61: "Visual ISO/IEC 13818-2 Main Profile",
	0x62: "Visual ISO/IEC 13818-2 SNR Profile",
	0x63: "Visual ISO/IEC 13818-2 Spatial Profile",
	0x64: "Visual ISO/IEC 13818-2 High Profile",
	0x65: "Visual ISO/IEC 13818-2 422 Profile",
	0x66: "Audio ISO/IEC 13818-7 Main Profile",
	0x67: "Audio ISO/IEC 13818-7 LowComplexity Profile",
	0x68: "Audio ISO/IEC 13818-7 Scaleable Sampling Rate Profile",
	0x69: "Audio ISO/IEC 13818-3",
	0x6A: "Visual ISO/IEC 11172-2",
	0x6B: "Audio ISO/IEC 11172-3",
	0x6C: "Visual ISO/IEC 10918-1",
	0x6D: "Portable Network Graphics (6)",
	0x6E: "Visual ISO/IEC 15444-1 (JPEG 2000)",
	0xA0: "EVRC Voice",
	0xA1: "SMV Voice",
	0xA2: "3GPP2 Compact Multimedia Format (CMF)",
	0xA3: "SMPTE VC-1 Video",
	0xA4: "Dirac Video Coder",
	0xA5: "withdrawn, unused, do not use (was AC-3)",
	0xA6: "withdrawn, unused, do not use (was Enhanced AC-3)",
	0xA7: "DRA Audio",
	0xA8: "ITU G.719 Audio",
	0xA9: "Core Substream",
	0xAA: "Core Substream + Extension Substream",
	0xAB: "Extension Substream containing only XLL",
	0xAC: "Extension Substream containing only LBR",
	0xAD: "Opus audio",
	0xAE: "withdrawn, unused, do not use (was AC-4)",
	0xAF: "Auro-Cx 3D audio",
	0xB0: "RealVideo Codec 11",
	0xB1: "VP9 Video",
	0xB2: "DTS-UHD profile 2",
	0xB3: "DTS-UHD profile 3 or higher",
	0xE1: "13K Voice",
	0xFF: "no object type specified (5)",
}

func GetObjectTypeIndicationDescription(code uint8) string {
	if desc, found := ObjectTypeIndicationMap[code]; found {
		return desc
	}
	if (code >= 0xC0 && code <= 0xE0) || (code >= 0xE2 && code <= 0xFE) {
		return "user private"
	}
	return ""
}

// AudioObjectTypeMap - AOT Audio Object Types (ISO/IEC 14496-3)
var AudioObjectTypeMap = map[uint8]string{
	0:  "Null",
	1:  "AAC main",
	2:  "AAC LC", // Low Complexity (AAC-LC)
	3:  "AAC SSR",
	4:  "AAC LTP",
	5:  "SBR", // Spectral Band Replication (HE-AAC, High Efficiency AAC)
	6:  "AAC Scalable",
	7:  "TwinVQ",
	8:  "CELP",
	9:  "HVXC",
	12: "TTSI",
	13: "Main synthetic",
	14: "Wavetable synthesis",
	15: "General MIDI",
	16: "Algorithmic Synthesis and Audio FX",
	17: "ER AAC LC",
	19: "ER AAC LTP",
	20: "ER AAC scalable",
	21: "ER TwinVQ",
	22: "ER BSAC",
	23: "ER AAC LD",
	24: "ER CELP",
	25: "ER HVXC",
	26: "ER HILN",
	27: "ER Parametric",
	28: "SSC",
	29: "PS", // Parametric Stereo (HE-AACv2, High Efficiency AAC v2)
	30: "MPEG Surround",
	31: "(escape)",
	32: "Layer-1",
	33: "Layer-2",
	34: "Layer-3",
	35: "DST",
	36: "ALS",
	37: "SLS",
	38: "SLS non-core",
	39: "ER AAC ELD",
	40: "SMR Simple",
	41: "SMR Main",
	42: "USAC",
	43: "SAOC",
	44: "LD MPEG Surround",
	45: "SAOC-DE",
	46: "Audio Sync",
}

func GetAudioObjectTypeDescription(code uint8) string {
	if desc, found := AudioObjectTypeMap[code]; found {
		return desc
	}
	return "(reserved)"
}

const (
	AudioObjectTypeNull uint8 = iota
	AudioObjectTypeAAC_main
	AudioObjectTypeAAC_LC
	AudioObjectTypeAAC_SSR
	AudioObjectTypeAAC_LTP
	AudioObjectTypeSBR
	AudioObjectTypeAAC_Scalable
	AudioObjectTypeTwinVQ
	AudioObjectTypeCELP
	AudioObjectTypeHVXC
	_
	_
	AudioObjectTypeTTSI
	AudioObjectTypeMainSynthetic
	AudioObjectTypeWavetableSynthesis
	AudioObjectTypeGeneralMIDI
	AudioObjectTypeAlgorithmicSynthesisAndAudioFX
	AudioObjectTypeER_AAC_LC
	_
	AudioObjectTypeER_AAC_LTP
	AudioObjectTypeER_AACScalable
	AudioObjectTypeER_TwinVQ
	AudioObjectTypeER_BSAC
	AudioObjectTypeER_AAC_LD
	AudioObjectTypeER_CELP
	AudioObjectTypeER_HVXC
	AudioObjectTypeER_HILN
	AudioObjectTypeERParametric
	AudioObjectTypeSSC
	AudioObjectTypePS
	AudioObjectTypeMPEG_Surround
	AudioObjectTypeEscape
	AudioObjectTypeLayer1
	AudioObjectTypeLayer2
	AudioObjectTypeLayer3
	AudioObjectTypeDST
	AudioObjectTypeALS
	AudioObjectTypeSLS
	AudioObjectTypeSLS_nonCore
	AudioObjectTypeER_AAC_ELD
	AudioObjectTypeSMR_Simple
	AudioObjectTypeSMR_Main
	AudioObjectTypeUSAC
	AudioObjectTypeSAOC
	AudioObjectTypeLD_MPEG_Surround
	AudioObjectTypeSAOC_DE
	AudioObjectTypeAudioSync
)

type MP4Codec struct {
	Codec
	ObjectTypeIndication uint8
	ObjectTypeID         uint8
}

func (c *MP4Codec) Initialize(str string) error {
	if err := c.Codec.Initialize(str); err != nil {
		return err
	}
	str = str[4:]

	if ch := str[0]; ch != '.' {
		return ErrCodecInvalid
	}
	str = str[1:]

	parts := strings.Split(str, ".")
	if len(parts) < 1 || len(parts) > 3 {
		return ErrCodecInvalid
	}

	assign(16, 8, &c.ObjectTypeIndication, parts[0])
	if len(parts) == 2 {
		assign(10, 8, &c.ObjectTypeID, parts[1])
	}
	return nil
}

func (c *MP4Codec) GetObjectTypeIndicationDescription() string {
	return GetObjectTypeIndicationDescription(c.ObjectTypeIndication)
}

func (c *MP4Codec) isComparable() bool {
	return c.ObjectTypeIndication == 0x40 &&
		c.GetCodecIndicator() == MPEG4IndicatorMP4A &&
		(c.ObjectTypeID == AudioObjectTypeAAC_LC ||
			c.ObjectTypeID == AudioObjectTypeSBR ||
			c.ObjectTypeID == AudioObjectTypePS)
}

func (c *MP4Codec) Compare(i ICodec) int {
	o := i.(*MP4Codec)
	if !c.isComparable() || !o.isComparable() {
		panic(ErrCodecUncomparable)
	}

	if c.ObjectTypeIndication != o.ObjectTypeIndication {
		return int(c.ObjectTypeIndication) - int(o.ObjectTypeIndication)
	}

	if c.ObjectTypeID != o.ObjectTypeID {
		return int(c.ObjectTypeID) - int(o.ObjectTypeID)
	}

	return 0
}

type MP4ACodec struct {
	MP4Codec
}

func (c *MP4ACodec) GetAudioObjectTypeDescription() string {
	return GetAudioObjectTypeDescription(c.ObjectTypeID)
}
