package hlsutils

import (
	"downloader/api/applemusic"
	"downloader/mp4/metadata"
	"downloader/mp4/mp4utils"
	"io"

	"github.com/Spidey120703/hls-m3u8/m3u8"
)

type MediaType int

const (
	MediaTypeVideoOnly MediaType = iota
	MediaTypeSong
	MediaTypeMusicVideo
)

type HLSParameters struct {
	TempDir           string
	TargetPath        string
	Type              MediaType
	MasterPlaylistURI string
	AdamID            string
	WebPlayback       *applemusic.WebPlaybackSong
	MetaData          *metadata.Metadata
	IsEncrypted       bool
}

type MediaPlaylistEntry struct {
	MediaPlaylistURI string
	MediaPlaylist    *m3u8.MediaPlaylist
	KeyURIs          map[string][]string
	URIs             []string
	Readers          io.ReadSeeker
	Decryptor        mp4utils.IDecryptor
	Muxer            *mp4utils.MuxContext
}

type IHandler interface {
	Execute() (err error)
}

type Context struct {
	Type                 MediaType
	AdamID               string
	TempDir              string
	TargetPath           string
	MetaData             *metadata.Metadata
	SessionData          map[string]string
	MasterPlaylistURI    string
	MasterPlaylist       *m3u8.MasterPlaylist
	Variant              *m3u8.Variant
	WebPlayback          *applemusic.WebPlaybackSong
	Muxer                *mp4utils.MuxContext
	MediaPlaylistEntries []*MediaPlaylistEntry
	IsEncrypted          bool
}

func NewHTTPLiveStream(p HLSParameters) (ctx *Context) {
	if len(p.TempDir) == 0 || len(p.TargetPath) == 0 {
		panic("directories config is empty")
	}
	ctx = &Context{}
	ctx.Type = p.Type
	ctx.AdamID = p.AdamID
	ctx.MasterPlaylistURI = p.MasterPlaylistURI
	ctx.TempDir = p.TempDir
	ctx.TargetPath = p.TargetPath
	ctx.WebPlayback = p.WebPlayback
	ctx.MetaData = p.MetaData
	ctx.IsEncrypted = p.IsEncrypted
	if len(p.MasterPlaylistURI) == 0 && p.WebPlayback != nil {
		ctx.MasterPlaylistURI = p.WebPlayback.HlsPlaylistURL
	}
	return
}

func (ctx *Context) Execute() (err error) {
	handlers := []IHandler{
		&PlaylistHandler{ctx},
		&DecryptHandler{ctx},
		&MuxHandler{ctx},
	}
	for _, handler := range handlers {
		err = handler.Execute()
		if err != nil {
			break
		}
	}
	return
}

func Main1() {
	applemusic.RefreshToken()
	if false {
		//webPlayback, err := applemusic.GetWebPlayback("1770791066")
		webPlayback, err := applemusic.GetWebPlayback("1446022245")
		if err != nil {
			return
		}
		var context = NewHTTPLiveStream(HLSParameters{
			TempDir:     "Temp/",
			Type:        MediaTypeMusicVideo,
			WebPlayback: webPlayback,
		})
		if err := context.Execute(); err != nil {
			panic(err)
		}
	}
	{
		var context = NewHTTPLiveStream(HLSParameters{
			TempDir:           "Temp/",
			Type:              MediaTypeSong,
			AdamID:            "1440830132",
			MasterPlaylistURI: "https://aod.itunes.apple.com/itunes-assets/HLSMusic112/v4/83/bf/38/83bf38c1-c3f4-d4a9-88ff-1dc2976599f2/P496362886_default.m3u8",
		})
		if err := context.Execute(); err != nil {
			println()
			println()
			println()
			println()
			println()
			println()
			println()
			println()
			println()
			println()
			panic(err)
		}
	}

	//err = context.Initialize("https://aod.itunes.apple.com/itunes-assets/HLSMusic112/v4/83/bf/38/83bf38c1-c3f4-d4a9-88ff-1dc2976599f2/P496362886_default.m3u8")
	//if err != nil {
	//	panic(err)
	//}
}

//func Main() {
//	bar := progressbar.NewOptions(100,
//		progressbar.OptionUseANSICodes(true),
//		progressbar.OptionEnableColorCodes(true),
//		progressbar.OptionShowIts(),
//		progressbar.OptionSetTheme(progressbar.Theme{
//			Saucer:        "=",
//			SaucerHead:    ansi.CSIFg256(237),
//			SaucerPadding: "-",
//			BarStart:      ansi.CSIFgRGB(249, 38, 114),
//			BarEnd:        ansi.CSIReset,
//		}),
//		progressbar.OptionShowElapsedTimeOnFinish(),
//		progressbar.OptionShowCount(),
//	)
//
//	for i := 0; i < 100; i++ {
//		bar.Add(1)
//		time.Sleep(100 * time.Millisecond)
//	}
//}
