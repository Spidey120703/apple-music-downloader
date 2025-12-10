package utils

import (
	"regexp"
	"strconv"
)

var NamedPlaceholderPattern = regexp.MustCompile(`\{([_A-Za-z][_0-9A-Za-z]*)}`)

func Format(s string, ctx map[string]string) string {
	return NamedPlaceholderPattern.ReplaceAllStringFunc(s, func(n string) string {
		if val, found := ctx[n[1:len(n)-1]]; found {
			return val
		}
		return n
	})
}

func FindStringSubmatchMap(re *regexp.Regexp, s string) map[string]string {
	submatches := re.FindStringSubmatch(s)
	if submatches == nil {
		return nil
	}
	result := map[string]string{}
	for idx, name := range re.SubexpNames() {
		if len(submatches[idx]) == 0 {
			continue
		}
		if len(name) > 0 {
			result[name] = submatches[idx]
		} else {
			result[strconv.Itoa(idx)] = submatches[idx]
		}
	}
	return result
}
