package bitstruct

import (
	"downloader/internal/media/mp4/bitstruct/bitio"
	"errors"
	"io"
	"math"
)

const LengthUnlimited = math.MaxUint32

type IFieldObject interface {
	// GetFieldSize returns size of dynamic field
	GetFieldSize(name string) uint

	// GetFieldLength returns length of dynamic field
	GetFieldLength(name string) uint

	// IsOptFieldEnabled check whether if the optional field is enabled
	IsOptFieldEnabled(name string) bool

	// StringifyField returns field value as string
	StringifyField(name string, indent string, depth int) (string, bool)

	IsPString(name string, bytes []byte, remainingSize uint64) bool

	BeforeUnmarshal(r io.ReadSeeker, size uint64) (n uint64, override bool, err error)

	OnReadField(name string, r bitio.ReadSeeker, leftBits uint64) (rbits uint64, override bool, err error)

	OnWriteField(name string, w bitio.Writer) (wbits uint64, override bool, err error)
}

type BaseFieldObject struct {
}

// GetFieldSize returns size of dynamic field
func (box *BaseFieldObject) GetFieldSize(string) uint {
	panic(errors.New("GetFieldSize not implemented"))
}

// GetFieldLength returns length of dynamic field
func (box *BaseFieldObject) GetFieldLength(string) uint {
	panic(errors.New("GetFieldLength not implemented"))
}

// IsOptFieldEnabled check whether if the optional field is enabled
func (box *BaseFieldObject) IsOptFieldEnabled(string) bool {
	return false
}

// StringifyField returns field value as string
func (box *BaseFieldObject) StringifyField(string, string, int) (string, bool) {
	return "", false
}

func (*BaseFieldObject) IsPString(name string, bytes []byte, remainingSize uint64) bool {
	return true
}

func (*BaseFieldObject) BeforeUnmarshal(io.ReadSeeker, uint64) (uint64, bool, error) {
	return 0, false, nil
}

func (*BaseFieldObject) OnReadField(string, bitio.ReadSeeker, uint64) (uint64, bool, error) {
	return 0, false, nil
}

func (*BaseFieldObject) OnWriteField(string, bitio.Writer) (uint64, bool, error) {
	return 0, false, nil
}
