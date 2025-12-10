package hlsutils

import (
	"downloader/internal/config"
	"downloader/internal/media/m3u8/hlsutils/codec"
	"downloader/pkg/LOG"
	"downloader/pkg/utils"
	"errors"
	"fmt"
	"math"
	"net/http"
	"slices"
	"strings"

	"github.com/Spidey120703/hls-m3u8/m3u8"
)

func OpenM3U8(url string) (playlist m3u8.Playlist, listType m3u8.ListType, err error) {
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer utils.CloseQuietly(resp.Body)

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("http status: %d %s", resp.StatusCode, resp.Status)
		return
	}

	return m3u8.DecodeFrom(resp.Body, true)
}

func getNumPixels(resolution string) int {
	var x, y int
	_, err := fmt.Sscanf(resolution, "%dx%d", &x, &y)
	if err != nil {
		return 0
	}
	return x * y
}

// variantLess
//
//	(1) Resolution: a < b (video only)
//	(2) FrameRate: a < b (video only)
//	(3) Codecs: a < b (mp4v, avc1, hev1, av01 | aac, ec-3, alac)
//	(4) Bandwidth: a < b
//	(5) VideoRange: a < b (SDR, HDR)
func variantLess(a, b *m3u8.Variant, isVideo bool) bool {

	if isVideo {
		resolutionDiff := getNumPixels(a.Resolution) - getNumPixels(b.Resolution)
		if resolutionDiff != 0 {
			return resolutionDiff < 0
		}

		rateDiff := a.FrameRate - b.FrameRate
		if math.Abs(rateDiff) > 1e-4 {
			return rateDiff < 0
		}
	}

	codecDiff, _ := codec.LessStr(strings.Split(a.Codecs, ",")[0], strings.Split(b.Codecs, ",")[0])
	if codecDiff {
		return codecDiff
	}

	if a.Bandwidth != b.Bandwidth {
		return a.Bandwidth < b.Bandwidth
	}

	if a.VideoRange != b.VideoRange {
		return a.VideoRange == "SDR" && b.VideoRange == "HDR"
	}

	return false
}

func completeURI(base, uri string) string {
	if !strings.HasPrefix(uri, "http") {
		return base[:strings.LastIndex(base, "/")+1] + strings.TrimLeft(uri, "/")
	}
	return uri
}

type PlaylistHandler struct {
	*Context
}

func (ctx *PlaylistHandler) loadMasterPlaylist() (err error) {
	if len(ctx.MasterPlaylistURI) == 0 {
		LOG.Error.Println("master playlist URI not specified")
		return errors.New("master playlist URI not specified")
	}

	playlist, listType, err := OpenM3U8(ctx.MasterPlaylistURI)
	if err != nil {
		return
	}

	switch listType {
	case m3u8.MASTER:
		ctx.MasterPlaylist = playlist.(*m3u8.MasterPlaylist)
	case m3u8.MEDIA:
		return errors.New("inappropriate m3u8 type")
	}
	return
}

func (ctx *PlaylistHandler) extractSessionData() error {
	ctx.SessionData = make(map[string]string)
	for _, sessionData := range ctx.MasterPlaylist.SessionDatas {
		ctx.SessionData[sessionData.DataId] = sessionData.Value
	}
	return nil
}

func (ctx *PlaylistHandler) selectVariant() (err error) {
	var variant *m3u8.Variant

	for _, v := range ctx.MasterPlaylist.Variants {
		if v.Iframe {
			continue
		}
		if variant == nil || variantLess(variant, v, ctx.Type == MediaTypeMusicVideo) {
			variant = v
			continue
		}
	}

	if variant == nil {
		return errors.New("no variant found")
	}

	ctx.Variant = variant
	ctx.MediaPlaylistEntries = append(ctx.MediaPlaylistEntries, &MediaPlaylistEntry{
		MediaPlaylistURI: completeURI(ctx.MasterPlaylistURI, variant.URI),
	})

	if ctx.Type == MediaTypeSong || ctx.Type == MediaTypeVideoOnly {
		return
	}

	for _, alternative := range variant.Alternatives {
		if len(alternative.URI) == 0 {
			continue
		}
		ctx.MediaPlaylistEntries = append(ctx.MediaPlaylistEntries, &MediaPlaylistEntry{
			MediaPlaylistURI: completeURI(ctx.MasterPlaylistURI, alternative.URI),
		})
	}
	return
}

func (ctx *PlaylistHandler) loadMediaPlaylist() (err error) {
	for _, entry := range ctx.MediaPlaylistEntries {
		var playlist m3u8.Playlist
		var listType m3u8.ListType
		playlist, listType, err = OpenM3U8(entry.MediaPlaylistURI)
		if err != nil {
			return
		}

		switch listType {
		case m3u8.MASTER:
			return errors.New("inappropriate m3u8 type")
		case m3u8.MEDIA:
			entry.MediaPlaylist = playlist.(*m3u8.MediaPlaylist)
		}
	}
	return
}

func (ctx *PlaylistHandler) extractKeyURIs() (err error) {
	for _, entry := range ctx.MediaPlaylistEntries {
		entry.KeyURIs = make(map[string][]string)
		for _, segment := range entry.MediaPlaylist.GetAllSegments() {
			for _, key := range segment.Keys {
				if !slices.Contains(entry.KeyURIs[key.Keyformat], key.URI) {
					entry.KeyURIs[key.Keyformat] = append(entry.KeyURIs[key.Keyformat], key.URI)
				}
			}
		}
	}
	return
}

func (ctx *PlaylistHandler) downloadSegments() (err error) {
	LOG.Info.Println("Downloading media segments...")

	for _, entry := range ctx.MediaPlaylistEntries {
		entry.URIs = []string{
			completeURI(entry.MediaPlaylistURI, entry.MediaPlaylist.Map.URI),
		}
		for _, segment := range entry.MediaPlaylist.GetAllSegments() {
			uri := completeURI(entry.MediaPlaylistURI, segment.URI)
			if slices.Contains(entry.URIs, uri) {
				continue
			}
			entry.URIs = append(entry.URIs, uri)
		}

		if err = utils.MultiDownload(entry.URIs, ctx.TempDir, config.NumThreads); err != nil {
			return
		}
	}

	LOG.Info.Println("Download completed.")
	return
}

func (ctx *PlaylistHandler) Execute() (err error) {
	if err = ctx.loadMasterPlaylist(); err != nil {
		return
	}
	if err = ctx.extractSessionData(); err != nil {
		return
	}
	if err = ctx.selectVariant(); err != nil {
		return
	}
	if err = ctx.loadMediaPlaylist(); err != nil {
		return
	}
	if err = ctx.extractKeyURIs(); err != nil {
		return
	}
	return ctx.downloadSegments()
}
