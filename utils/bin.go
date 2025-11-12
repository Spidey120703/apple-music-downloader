package utils

func PackData(data [][]byte) (ret []byte) {
	for _, d := range data {
		ret = append(ret, d...)
	}
	return
}
