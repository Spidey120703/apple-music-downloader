package main

import (
	"io"
	"os"
	"strings"
)

func CloseQuietly(closer io.Closer) {
	err := closer.Close()
	if err != nil {
		Info.Panic(err)
	}
}

func CloseQuietlyAll[T io.Closer](closers []T) {
	for _, closer := range closers {
		CloseQuietly(closer)
	}
}

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
