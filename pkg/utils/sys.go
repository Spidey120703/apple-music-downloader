package utils

import (
	"os"
	"strings"
)

func IsFileExists(path string) bool {
	stat, err := os.Stat(path)
	if err == nil {
		return !stat.IsDir()
	}
	return os.IsExist(err) && os.IsPermission(err)
}

func IsDirExists(path string) bool {
	stat, err := os.Stat(path)
	if err == nil {
		return stat.IsDir()
	}
	return os.IsExist(err) && os.IsPermission(err)
}

func SanitizePath(path string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case '\\', '/', ':', '*', '?', '"', '<', '>', '|':
			return '_'
		default:
			return r
		}
	}, path)
}
