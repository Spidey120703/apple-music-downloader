package hlsutils

import (
	"downloader/internal/api/applemusic"
	"downloader/internal/media/mp4/metadata"
	"downloader/internal/media/mp4/mp4utils"
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
