package utils

import (
	"downloader/LOG"
	"io"
	"os"
)

func CloseQuietly(closer io.Closer) {
	err := closer.Close()
	if err != nil {
		LOG.Info.Panic(err)
	}
}

func CloseQuietlyAll[T io.Closer](closers []T) {
	for _, closer := range closers {
		CloseQuietly(closer)
	}
}

func OpenFiles(pathList []string) (fileList []*os.File, err error) {
	var file *os.File
	for _, filePath := range pathList {
		if file, err = os.Open(filePath); err != nil {
			return
		}
		fileList = append(fileList, file)
	}
	return
}

type NullWriter struct {
	Size   uint64
	Offset int64
}

func NewNullWriter() *NullWriter {
	return &NullWriter{}
}

func (b *NullWriter) Close() error { return nil }

func (b *NullWriter) Write(data []byte) (n int, err error) {
	n = len(data)
	b.Offset += int64(n)
	b.Size = max(b.Size, uint64(b.Offset))
	return
}

func (b *NullWriter) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		b.Offset = offset
	case io.SeekCurrent:
		b.Offset += offset
	case io.SeekEnd:
		b.Offset = int64(b.Size) - offset
	}
	b.Offset = min(int64(b.Size), max(0, b.Offset))
	return b.Offset, nil
}
