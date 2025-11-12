package utils

import (
	"downloader/log"
	"io"
)

func CloseQuietly(closer io.Closer) {
	err := closer.Close()
	if err != nil {
		log.Info.Panic(err)
	}
}

func CloseQuietlyAll[T io.Closer](closers []T) {
	for _, closer := range closers {
		CloseQuietly(closer)
	}
}
