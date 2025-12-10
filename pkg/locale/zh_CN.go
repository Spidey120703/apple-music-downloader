package locale

import (
	"slices"
)

var CNGenreMapping = map[string][]string{
	"Blues":              {"蓝调"},
	"Country":            {"乡村", "乡村音乐"},
	"Dance":              {"舞曲"},
	"Jazz":               {"爵士乐"},
	"Pop":                {"国际流行", "流行乐"},
	"Rap":                {"说唱"},
	"Reggae":             {"雷鬼"},
	"Rock":               {"摇滚"},
	"Techno":             {"铁克诺"},
	"Alternative":        {"另类音乐", "另类"},
	"Soundtrack":         {"原声音乐", "电影配乐"},
	"Vocal":              {"声乐"},
	"Trance":             {"迷幻舞曲"},
	"Classical":          {"古典"},
	"House":              {"House 作品", "浩室"},
	"Punk":               {"朋克"},
	"Electronic":         {"电子音乐", "电子乐"},
	"Musical":            {"音乐剧"},
	"Hard Rock":          {"硬摇滚"},
	"Folk":               {"民谣"},
	"World Music":        {"世界音乐"},
	"Indie Rock":         {"独立摇滚"},
	"Blues/R&B":          {"蓝调音乐/R&B"},
	"Children’s Music":   {"儿童音乐"},
	"Psychedelic":        {"迷幻乐"},
	"Hip Hop/Rap":        {"嘻哈/说唱"},
	"Hip-Hop/Rap":        {"嘻哈/说唱"},
	"Industrial":         {"工业乐"},
	"Holiday":            {"节庆", "节日乐"},
	"Jungle/Drum'n'bass": {"丛林/鼓打贝斯"},
	"Original Score":     {"原创音乐"},
	"R&B/Soul":           {"R&B/灵魂乐"},
	"Religious":          {"宗教乐"},
	"Rock y Alternativo": {"另类和南美摇滚"},
	"Unclassifiable":     {"不可归类音乐"},
	"World":              {"世界音乐", "世界乐"},
	"Easy Listening":     {"轻音乐"},
	"New Age":            {"新生代音乐"},
}

func TranslateFromZhCN(text string) string {
	for origin, translated := range CNGenreMapping {
		if slices.Contains(translated, text) {
			return origin
		}
	}
	return text
}
