package main

// note: Extracted Container Classes

type Sample struct {
	Data                   []byte
	SampleDescriptionIndex uint32
	SampleDuration         uint32
}

type Chunk struct {
	samples []Sample
}

type ICodecInfo interface {
	Duration() uint64
	Samples() []Sample
}

func PackData(data [][]byte) (ret []byte) {
	for _, d := range data {
		ret = append(ret, d...)
	}
	return
}
