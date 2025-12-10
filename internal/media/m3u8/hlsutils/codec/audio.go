package codec

var (
	ALACIndicator = CodecIndicator{0x61, 0x6c, 0x61, 0x63} // alac
	EC3Indicator  = CodecIndicator{0x65, 0x63, 0x2d, 0x33} // ec-3
)

type ALACCodec struct {
	Codec
}

type EC3Codec struct {
	Codec
}
