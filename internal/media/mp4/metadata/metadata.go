package metadata

import (
	"bytes"
	"downloader/internal/media/mp4/boxtree"
	"downloader/internal/media/mp4/cmaf"
	"encoding/binary"
	"errors"
	"reflect"

	"github.com/Spidey120703/go-mp4"
)

type Track struct {
	TrackNumber uint32
	TrackCount  uint16
	Reserved    uint16
}

type Disk struct {
	DiskNumber uint32
	DiskCount  uint16
}

type Metadata struct {
	Title          *string `ilst:"\xA9nam"`
	ArtistName     *string `ilst:"\xA9ART"`
	PlaylistArtist *string `ilst:"aART"`
	ComposerName   *string `ilst:"\xA9wrt"`
	AlbumName      *string `ilst:"\xA9alb"`
	Work           *string `ilst:"\xA9grp"`
	Genre          []byte  `ilst:"gnre"`
	Track          *Track  `ilst:"trkn"`
	DiskNumber     *Disk   `ilst:"disk"`
	Compilation    *uint8  `ilst:"cpil"`
	PlayGap        *uint8  `ilst:"pgap"`
	ReleaseDate    *string `ilst:"\xA9day"`
	AppleID        *string `ilst:"apID"`
	Owner          *string `ilst:"ownr"`
	Copyright      *string `ilst:"cprt"`
	ItemID         *uint32 `ilst:"cnID"`
	ArtistID       *uint32 `ilst:"atID"`
	Rating         *uint8  `ilst:"rtng"`
	ComposerID     *uint32 `ilst:"cmID"`
	PlaylistID     *uint32 `ilst:"plID"`
	GenreID        *uint32 `ilst:"geID"`
	StorefrontID   *uint32 `ilst:"sfID"`
	HDVideo        *uint8  `ilst:"hdvd"`
	MediaType      *uint8  `ilst:"stik"`
	PurchaseDate   *string `ilst:"purd"`
	SortName       *string `ilst:"sonm"`
	SortAlbum      *string `ilst:"soal"`
	SortArtist     *string `ilst:"soar"`
	SortComposer   *string `ilst:"soco"`
	XID            *string `ilst:"xid "`
	Flavor         *string `ilst:"flvr"`
	Cover          []byte  `ilst:"covr"`
	Lyrics         *string `ilst:"\xA9lyr"`
}

func detectBinaryDataType(data []byte) uint32 {
	if len(data) < 8 {
		return mp4.DatatypeReserved
	}
	// Joint Photographic Experts Group: ISO/IEC 10918-1
	// Portable Network Graphic: ISO/IEC 15948
	var (
		JPEGMarkerSOI = []byte{0xFF, 0xD8}
		JPEGMarkerEOI = []byte{0xFF, 0xD9}
		PNGSignature  = []byte{137, 80, 78, 71, 13, 10, 26, 10}
		BMPv2FileType = []byte("BM")
	)
	if bytes.Equal(data[:2], JPEGMarkerSOI) && bytes.Equal(data[len(data)-2:], JPEGMarkerEOI) &&
		data[2] == 0xFF && data[3]&0xF0 == 0xE0 {
		return mp4.DataTypeJPEG
	}
	if bytes.Equal(data[:8], PNGSignature) {
		return mp4.DataTypePNG
	}
	if bytes.Equal(data[:2], BMPv2FileType) {
		return mp4.DataTypeBMP
	}
	return mp4.DatatypeReserved
}

func (m *Metadata) Walk(callback func(mp4.BoxType, *mp4.Data) error) (err error) {

	t := reflect.TypeOf(*m)
	v := reflect.ValueOf(*m)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)
		if value.Pointer() == 0 {
			continue
		}

		boxType := mp4.StrToBoxType(field.Tag.Get("ilst"))

		data := &mp4.Data{}
		switch field.Type.Kind() {
		case reflect.Pointer:
			switch field.Type.Elem().Kind() {
			// *string
			case reflect.String:
				data.DataType = mp4.DataTypeUTF8
				data.Data = []byte(value.Elem().Interface().(string))
			// *uint8 | *byte
			case reflect.Uint8:
				data.DataType = mp4.DataTypeInt
				data.Data = []byte{value.Elem().Interface().(uint8)}
			// *uint32
			case reflect.Uint32:
				data.DataType = mp4.DataTypeInt
				data.Data = make([]byte, 4)
				binary.BigEndian.PutUint32(data.Data, value.Elem().Interface().(uint32))
			// * struct {}
			case reflect.Struct:
				data.DataType = mp4.DatatypeReserved
				data.Data = make([]byte, 0)
				{
					for i := 0; i < value.Elem().NumField(); i++ {
						switch value.Elem().Field(i).Type().Kind() {
						case reflect.Uint8:
							data.Data = append(data.Data, value.Elem().Field(i).Interface().(uint8))
						case reflect.Uint16:
							data.Data = binary.BigEndian.AppendUint16(data.Data, value.Elem().Field(i).Interface().(uint16))
						case reflect.Uint32:
							data.Data = binary.BigEndian.AppendUint32(data.Data, value.Elem().Field(i).Interface().(uint32))
						default:
							continue
						}
					}
				}
			default:
				return errors.New("unhandled data pointer")
			}
		// []byte | []uint8
		case reflect.Slice:
			switch field.Type.Elem().Kind() {
			case reflect.Uint8:
				data.Data = value.Interface().([]byte)
				data.DataType = detectBinaryDataType(data.Data)
			default:
				return errors.New("unhandled data array")
			}
		default:
			return errors.New("unhandled data")
		}

		err = callback(boxType, data)
		if err != nil {
			return err
		}
	}
	return
}

func (m *Metadata) Attach(root *boxtree.BoxNode) (err error) {
	var header *cmaf.Header
	header, err = cmaf.InitializeHeader(root)
	if err != nil {
		return err
	}

	meta := boxtree.BoxNode{
		Info: &mp4.BoxInfo{Type: mp4.BoxTypeMeta(), Context: mp4.Context{UnderUdta: true}},
		Box:  &mp4.Meta{},
		Path: boxtree.ToAppendedPath(header.Moov.Udta.Node.Path, mp4.BoxTypeMeta()),
	}

	{
		hdlr := boxtree.BoxNode{
			Info: &mp4.BoxInfo{Type: mp4.BoxTypeHdlr(), Context: mp4.Context{UnderUdta: true}},
			Box: &mp4.MetadataHandlerBox{
				HandlerType: [4]byte{'m', 'd', 'i', 'r'},
				Name:        [14]byte{'a', 'p', 'p', 'l'},
			},
			Path: boxtree.ToAppendedPath(meta.Path, mp4.BoxTypeHdlr()),
		}
		meta.Children = append(meta.Children, &hdlr)

		ilst := boxtree.BoxNode{
			Info: &mp4.BoxInfo{Type: mp4.BoxTypeIlst(), Context: mp4.Context{UnderUdta: true}},
			Box:  &mp4.Ilst{},
			Path: boxtree.ToAppendedPath(meta.Path, mp4.BoxTypeIlst()),
		}

		if err = m.Walk(func(boxType mp4.BoxType, data *mp4.Data) (err error) {
			item := &boxtree.BoxNode{
				Info: &mp4.BoxInfo{Type: boxType, Context: mp4.Context{UnderUdta: true, UnderIlst: true}},
				Box: &mp4.IlstMetaContainer{
					AnyTypeBox: mp4.AnyTypeBox{
						Type: boxType,
					},
				},
				Path: boxtree.ToAppendedPath(ilst.Path, boxType),
			}

			item.Children = append(item.Children, &boxtree.BoxNode{
				Info: &mp4.BoxInfo{Type: mp4.BoxTypeData(), Context: mp4.Context{UnderUdta: true, UnderIlst: true, UnderIlstMeta: true}},
				Box:  data,
				Path: boxtree.ToAppendedPath(item.Path, mp4.BoxTypeData()),
			})
			if err = item.Caching(); err != nil {
				return
			}
			ilst.Children = append(ilst.Children, item)
			return
		}); err != nil {
			return
		}

		if err = ilst.Caching(); err != nil {
			return
		}

		meta.Children = append(meta.Children, &ilst)

		meta.Children = append(meta.Children, &boxtree.BoxNode{
			Info: &mp4.BoxInfo{Type: mp4.BoxTypeFree(), Context: mp4.Context{UnderUdta: true}},
			Box:  &mp4.Free{},
			Path: boxtree.ToAppendedPath(meta.Path, mp4.BoxTypeFree()),
		})
	}

	if err = meta.Caching(); err != nil {
		return
	}

	header.Moov.Udta.Node.Children = append(header.Moov.Udta.Node.Children, &meta)
	if err = header.Moov.Udta.Node.Caching(); err != nil {
		return
	}
	return
}
