package main

import (
	"downloader/internal/api"
	"downloader/internal/api/applemusic"
	"downloader/internal/api/itunes"
	"downloader/internal/config"
	"downloader/internal/downloader"
	"downloader/internal/media/m3u8/hlsutils"
	"downloader/internal/media/mp4/metadata"
	"downloader/internal/media/ttml"
	"downloader/pkg/LOG"
	"downloader/pkg/utils"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"
)

type APIContext struct {
	iTunes struct {
		Song       *itunes.Song
		MusicVideo *itunes.MusicVideo
	}
	AppleMusic struct {
		Songs       *applemusic.Songs
		Albums      *applemusic.Albums
		MusicVideos *applemusic.MusicVideos
	}
	MZPlay struct {
		WebPlayback *applemusic.WebPlaybackSong
	}
	AlbumCoverData []byte
}

const (
	ExtM4A = ".m4a"
	ExtM4V = ".m4v"
	ExtMP4 = ".mp4"
)

type FullPath struct {
	TargetPath string
	ArtistDir  string
	AlbumDir   string
	DiscDir    string
	TrackName  string
	Ext        string
}

func (fp *FullPath) AlbumPath(elem ...string) string {
	var parts []string
	for _, e := range elem {
		parts = append(parts, utils.SanitizePath(e))
	}
	return path.Join(
		fp.TargetPath,
		utils.SanitizePath(fp.ArtistDir),
		utils.SanitizePath(fp.AlbumDir),
		path.Join(parts...))
}

func (fp *FullPath) String() string {
	return path.Join(
		fp.TargetPath,
		utils.SanitizePath(fp.ArtistDir),
		utils.SanitizePath(fp.AlbumDir),
		utils.SanitizePath(fp.DiscDir),
		utils.SanitizePath(fp.TrackName)+fp.Ext)
}

type Downloader struct {
	TargetPath string
}

func (d *Downloader) DownloadAlbum(albumID string, ctx APIContext, fullPath FullPath) (err error) {
	if ctx.AppleMusic.Albums == nil {
		if ctx.AppleMusic.Albums, err = applemusic.GetAlbumData(albumID); err != nil {
			return
		}
	}

	{
		if len(fullPath.TargetPath) == 0 {
			fullPath.TargetPath = d.TargetPath
		}
		if len(fullPath.ArtistDir) == 0 {
			fullPath.ArtistDir = *ctx.AppleMusic.Albums.Attributes.ArtistName
		}
		fullPath.AlbumDir = fmt.Sprintf(
			"%s - %s [%s]",
			*ctx.AppleMusic.Albums.Attributes.ReleaseDate,
			*ctx.AppleMusic.Albums.Attributes.Name,
			*ctx.AppleMusic.Albums.Attributes.Upc)
	}

	{
		LOG.Info.Printf("Downloading album: %s", fullPath.AlbumDir)
		LOG.Info.Println("Album Info:")
		LOG.Info.Printf("\t\t%16s:  %s", "Album Name", *ctx.AppleMusic.Albums.Attributes.Name)
		LOG.Info.Printf("\t\t%16s:  %s", "Artist Name", *ctx.AppleMusic.Albums.Attributes.ArtistName)
		LOG.Info.Printf("\t\t%16s:  %s", "Genre Names", strings.Join(ctx.AppleMusic.Albums.Attributes.GenreNames, ", "))
		LOG.Info.Printf("\t\t%16s:  %s", "Release Date", *ctx.AppleMusic.Albums.Attributes.ReleaseDate)
		LOG.Info.Printf("\t\t%16s:  %s", "Record Label", *ctx.AppleMusic.Albums.Attributes.RecordLabel)
		LOG.Info.Printf("\t\t%16s:  %s", "Copyright", *ctx.AppleMusic.Albums.Attributes.Copyright)
		LOG.Info.Printf("\t\t%16s:  %s", "UPC", *ctx.AppleMusic.Albums.Attributes.Upc)
		LOG.Info.Println()
	}

	if ctx.AlbumCoverData, err = downloader.ReadCover(*ctx.AppleMusic.Albums.Attributes.Artwork, fullPath.AlbumPath("Cover{original_file_ext}")); err != nil {
		return
	}

	if ctx.AppleMusic.Albums.Attributes.EditorialArtwork != nil {
		artworks := make(map[string]*applemusic.Artwork)
		artworks["BannerUber"] = ctx.AppleMusic.Albums.Attributes.EditorialArtwork.BannerUber
		artworks["OriginalFlowcaseBrick"] = ctx.AppleMusic.Albums.Attributes.EditorialArtwork.OriginalFlowcaseBrick
		artworks["StaticDetailSquare"] = ctx.AppleMusic.Albums.Attributes.EditorialArtwork.StaticDetailSquare
		artworks["StaticDetailTall"] = ctx.AppleMusic.Albums.Attributes.EditorialArtwork.StaticDetailTall
		artworks["StoreFlowcase"] = ctx.AppleMusic.Albums.Attributes.EditorialArtwork.StoreFlowcase
		artworks["SubscriptionHero"] = ctx.AppleMusic.Albums.Attributes.EditorialArtwork.SubscriptionHero
		artworks["SuperHeroTall"] = ctx.AppleMusic.Albums.Attributes.EditorialArtwork.SuperHeroTall

		for _, artwork := range artworks {
			if artwork == nil {
				continue
			}
			_, err = downloader.DownloadArtwork(*artwork, fullPath.AlbumPath("Extras", "Artworks", downloader.FilenameFormatOriginalFileName))
		}
	}

	if ctx.AppleMusic.Albums.Attributes.EditorialVideo != nil {
		motionVideos := make(map[string]*applemusic.MotionVideo)
		motionVideos["MotionSquareVideo1X1"] = ctx.AppleMusic.Albums.Attributes.EditorialVideo.MotionSquareVideo1X1
		motionVideos["MotionDetailSquare"] = ctx.AppleMusic.Albums.Attributes.EditorialVideo.MotionDetailSquare
		motionVideos["MotionDetailTall"] = ctx.AppleMusic.Albums.Attributes.EditorialVideo.MotionDetailTall

		for _, motionVideo := range motionVideos {
			if motionVideo == nil {
				continue
			}
			name := path.Base(*motionVideo.Video)
			name = name[:strings.LastIndex(name, ".")] + ExtMP4
			_, err = downloader.DownloadMotionVideo(*motionVideo, fullPath.AlbumPath("Extras", "MotionVideos", name))
		}
	}

	LOG.Info.Printf("Start to download %d tracks\n", len(ctx.AppleMusic.Albums.Relationships.Tracks.Data))

	for _, track := range ctx.AppleMusic.Albums.Relationships.Tracks.Data {
		LOG.Info.Println(strings.Repeat("=", 128))
		fullPath.DiscDir = fmt.Sprintf("Disc %d", *track.Attributes.DiscNumber)

		switch *track.Type {
		case "songs":
			ctx.AppleMusic.Songs = track.AsSongs()
			if err = d.DownloadSong(*track.ID, ctx, fullPath); err != nil {
				return
			}

			if track.Relationships.MusicVideos != nil && len(track.Relationships.MusicVideos.Data) > 0 {
				LOG.Info.Println(strings.Repeat("=", 128))
				LOG.Info.Printf("Downloading relative music videos: %d-%d %s", *track.Attributes.DiscNumber, *track.Attributes.TrackNumber, *track.Attributes.Name)

				fullPath.DiscDir = "Music Videos"
				for _, musicVideo := range track.Relationships.MusicVideos.Data {
					LOG.Info.Println(">" + strings.Repeat("=", 128))
					ctx.AppleMusic.MusicVideos = &musicVideo
					if err = d.DownloadMusicVideo(*musicVideo.ID, ctx, fullPath); err != nil {
						return
					}
				}
			}
		case "music-videos":
			ctx.AppleMusic.MusicVideos = track.AsMusicVideos()
			if err = d.DownloadMusicVideo(*track.ID, ctx, fullPath); err != nil {
				return
			}
		default:
			LOG.Warn.Printf("Type '%s' is not available to download", *track.Type)
		}
	}

	return
}

func (d *Downloader) DownloadSong(trackID string, ctx APIContext, fullPath FullPath) (err error) {
	if ctx.AppleMusic.Songs == nil {
		if ctx.AppleMusic.Songs, err = applemusic.GetSongData(trackID); err != nil {
			return
		}
	}
	if ctx.AppleMusic.Albums == nil {
		if len(ctx.AppleMusic.Songs.Relationships.Albums.Data) == 0 {
			return errors.New("no albums related")
		}
		ctx.AppleMusic.Albums = &ctx.AppleMusic.Songs.Relationships.Albums.Data[0]
	}

	{
		if len(fullPath.TargetPath) == 0 {
			fullPath.TargetPath = d.TargetPath
		}
		if len(fullPath.ArtistDir) == 0 {
			fullPath.ArtistDir = *ctx.AppleMusic.Songs.Attributes.ArtistName
		}
		if len(fullPath.AlbumDir) == 0 {
			fullPath.AlbumDir = fmt.Sprintf(
				"%s - %s [%s]",
				*ctx.AppleMusic.Albums.Attributes.ReleaseDate,
				*ctx.AppleMusic.Albums.Attributes.Name,
				*ctx.AppleMusic.Albums.Attributes.Upc)
		}
		if len(fullPath.DiscDir) == 0 {
			fullPath.DiscDir = fmt.Sprintf("Disc %d", *ctx.AppleMusic.Songs.Attributes.DiscNumber)
		}

		fullPath.TrackName = fmt.Sprintf(
			"%d. %s",
			*ctx.AppleMusic.Songs.Attributes.TrackNumber,
			*ctx.AppleMusic.Songs.Attributes.Name)
		fullPath.Ext = ExtM4A
	}

	{
		LOG.Info.Printf("Downloading song: %d-%d %s", *ctx.AppleMusic.Songs.Attributes.DiscNumber, *ctx.AppleMusic.Songs.Attributes.TrackNumber, *ctx.AppleMusic.Songs.Attributes.Name)
		LOG.Info.Println("Media Info:")
		LOG.Info.Printf("\t\t%16s:  %s", "Track Title", *ctx.AppleMusic.Songs.Attributes.Name)
		LOG.Info.Printf("\t\t%16s:  %s", "Artist Name", *ctx.AppleMusic.Songs.Attributes.ArtistName)
		LOG.Info.Printf("\t\t%16s:  %d", "Disc Number", *ctx.AppleMusic.Songs.Attributes.DiscNumber)
		LOG.Info.Printf("\t\t%16s:  %d", "Track Number", *ctx.AppleMusic.Songs.Attributes.TrackNumber)
		LOG.Info.Printf("\t\t%16s:  %s", "ISRC", *ctx.AppleMusic.Songs.Attributes.Isrc)
		if ctx.AppleMusic.Songs.Attributes.WorkName != nil {
			LOG.Info.Printf("\t\t%16s:  %s", "Work Name", *ctx.AppleMusic.Songs.Attributes.WorkName)
		}
		LOG.Info.Printf("\t\t%16s:  %s", "Genre Names", strings.Join(ctx.AppleMusic.Songs.Attributes.GenreNames, ", "))
		LOG.Info.Println()
	}

	if ctx.iTunes.Song == nil {
		if ctx.iTunes.Song, err = itunes.GetITunesInfo[itunes.Song](trackID, "song"); err != nil {
			return
		}
	}
	if ctx.MZPlay.WebPlayback == nil {
		if ctx.MZPlay.WebPlayback, err = applemusic.GetWebPlayback(trackID); err != nil {
			LOG.Error.Printf("failed to get MZPlay web playback assets: %v", err)
			ctx.MZPlay.WebPlayback = &applemusic.WebPlaybackSong{}
		}
	}

	var ttmlRaw, lyrics string
	if *ctx.AppleMusic.Songs.Attributes.HasLyrics {
		if err = d.DownloadLyrics(trackID, ctx, fullPath); err != nil {
			LOG.Error.Printf("failed to download lyrics: %v", err)
		}
		if ttmlRaw, err = applemusic.GetLyrics(trackID); err != nil {
			LOG.Error.Printf("failed to download lyrics: %v", err)
		}
		if ttmlRaw != "" {
			if lyrics, err = ttml.ExtractTextFromTTML(ttmlRaw); err != nil {
				return
			}
		}
	}

	if ctx.AlbumCoverData == nil {
		if artwork := ctx.AppleMusic.Songs.Attributes.Artwork; artwork != nil {
			if ctx.AlbumCoverData, err = downloader.ReadCover(*artwork, path.Join(config.Get().Storage.TempPath, downloader.FilenameFormatUUID)); err != nil {
				return
			}
		}
	}

	var params = hlsutils.HLSParameters{
		TempDir:    config.Get().Storage.TempPath,
		TargetPath: fullPath.String(),
		Type:       hlsutils.MediaTypeSong,
		AdamID:     trackID,
		MetaData: metadata.LoadSongMetadata(metadata.Context{
			WebPlayback:     ctx.MZPlay.WebPlayback,
			AppleMusicSongs: ctx.AppleMusic.Songs,
			AppleMusicAlbum: ctx.AppleMusic.Albums,
			ItunesSong:      ctx.iTunes.Song,
			CoverData:       ctx.AlbumCoverData,
			LyricsData:      lyrics,
		}),
		IsEncrypted: true,
	}

	if ctx.AppleMusic.Songs.Attributes.ExtendedAssetUrls.EnhancedHls != nil {
		params.MasterPlaylistURI = *ctx.AppleMusic.Songs.Attributes.ExtendedAssetUrls.EnhancedHls
	} else {
		LOG.Warn.Printf("No enhanced HLS found, falling back to download 256 kbps AAC")
		for _, asset := range ctx.MZPlay.WebPlayback.Assets {
			if asset.Flavor == "28:ctrp256" {
				params.MediaPlaylistURI = asset.URL
				params.WebPlayback = ctx.MZPlay.WebPlayback
			}
		}
	}

	if params.MasterPlaylistURI == "" && params.MediaPlaylistURI == "" {
		LOG.Error.Printf("No downloadable media assets found.")
		return
	}

	var context = hlsutils.NewHTTPLiveStream(params)
	if err = context.Execute(); err == nil {
		LOG.Info.Printf("Download completed, saved to: %s", fullPath.String())
	}
	return
}

func (d *Downloader) DownloadMusicVideo(trackID string, ctx APIContext, fullPath FullPath) (err error) {
	if ctx.AppleMusic.MusicVideos == nil {
		if ctx.AppleMusic.MusicVideos, err = applemusic.GetMusicVideoData(trackID); err != nil {
			return
		}
	}
	if ctx.AppleMusic.Albums == nil {
		ctx.AppleMusic.Albums = &ctx.AppleMusic.MusicVideos.Relationships.Albums.Data[0]
	}

	var mvSrc metadata.MusicVideoType
	{
		if len(fullPath.TargetPath) == 0 {
			fullPath.TargetPath = d.TargetPath
		}
		if len(fullPath.ArtistDir) == 0 {
			fullPath.ArtistDir = *ctx.AppleMusic.MusicVideos.Attributes.ArtistName
		}

		if ctx.AppleMusic.MusicVideos.Attributes.TrackNumber != nil {
			if len(fullPath.AlbumDir) == 0 {
				fullPath.AlbumDir = fmt.Sprintf(
					"%s - %s [%s]",
					*ctx.AppleMusic.Albums.Attributes.ReleaseDate,
					*ctx.AppleMusic.Albums.Attributes.Name,
					*ctx.AppleMusic.Albums.Attributes.Upc)
			}
			if len(fullPath.DiscDir) == 0 {
				fullPath.DiscDir = fmt.Sprintf("Disc %d", *ctx.AppleMusic.MusicVideos.Attributes.DiscNumber)
			}

			mvSrc = metadata.MusicVideoTypeFromAlbum
			fullPath.TrackName = fmt.Sprintf(
				"%d. %s",
				*ctx.AppleMusic.MusicVideos.Attributes.TrackNumber,
				*ctx.AppleMusic.MusicVideos.Attributes.Name)
		} else {
			if len(fullPath.DiscDir) == 0 {
				fullPath.DiscDir = "Music Videos"
			}

			mvSrc = metadata.MusicVideoFromSongs
			fullPath.TrackName = fmt.Sprintf(
				"%s [%s]",
				*ctx.AppleMusic.MusicVideos.Attributes.Name,
				*ctx.AppleMusic.MusicVideos.Attributes.Isrc)
		}
		fullPath.Ext = ExtM4V
	}

	{
		if mvSrc == metadata.MusicVideoTypeFromAlbum {
			LOG.Info.Printf("Downloading music video: %d-%d %s", *ctx.AppleMusic.MusicVideos.Attributes.DiscNumber, *ctx.AppleMusic.MusicVideos.Attributes.TrackNumber, *ctx.AppleMusic.MusicVideos.Attributes.Name)
			LOG.Info.Println("Media Info:")
			LOG.Info.Printf("\t\t%16s:  %s", "Track Title", *ctx.AppleMusic.MusicVideos.Attributes.Name)
			LOG.Info.Printf("\t\t%16s:  %s", "Artist Name", *ctx.AppleMusic.MusicVideos.Attributes.ArtistName)
			LOG.Info.Printf("\t\t%16s:  %d", "Disc Number", *ctx.AppleMusic.MusicVideos.Attributes.DiscNumber)
			LOG.Info.Printf("\t\t%16s:  %d", "Track Number", *ctx.AppleMusic.MusicVideos.Attributes.TrackNumber)
			LOG.Info.Printf("\t\t%16s:  %s", "ISRC", *ctx.AppleMusic.MusicVideos.Attributes.Isrc)
			if ctx.AppleMusic.MusicVideos.Attributes.WorkName != nil {
				LOG.Info.Printf("\t\t%16s:  %s", "Work Name", *ctx.AppleMusic.MusicVideos.Attributes.WorkName)
			}
			LOG.Info.Printf("\t\t%16s:  %s", "Genre Names", strings.Join(ctx.AppleMusic.MusicVideos.Attributes.GenreNames, ", "))
		} else {
			LOG.Info.Printf("Downloading music video: %s [%s]", *ctx.AppleMusic.MusicVideos.Attributes.Name, *ctx.AppleMusic.MusicVideos.Attributes.Isrc)
			LOG.Info.Println("Media Info:")
			LOG.Info.Printf("\t\t%16s:  %s", "Track Title", *ctx.AppleMusic.MusicVideos.Attributes.Name)
			LOG.Info.Printf("\t\t%16s:  %s", "Artist Name", *ctx.AppleMusic.MusicVideos.Attributes.ArtistName)
			LOG.Info.Printf("\t\t%16s:  %s", "ISRC", *ctx.AppleMusic.MusicVideos.Attributes.Isrc)
		}
		LOG.Info.Println()
	}

	if ctx.iTunes.MusicVideo == nil {
		if ctx.iTunes.MusicVideo, err = itunes.GetITunesInfo[itunes.MusicVideo](trackID, "song"); err != nil {
			return
		}
	}
	if ctx.MZPlay.WebPlayback == nil {
		if ctx.MZPlay.WebPlayback, err = applemusic.GetWebPlayback(trackID); err != nil {
			LOG.Error.Printf("failed to fetch HLS manifest: %v", err)
			return nil
		}
	}

	var coverData []byte
	if artwork := ctx.AppleMusic.MusicVideos.Attributes.Artwork; artwork != nil {
		if coverData, err = downloader.ReadCover(*artwork, path.Join(config.Get().Storage.TempPath, downloader.FilenameFormatUUID)); err != nil {
			return
		}
	}

	var context = hlsutils.NewHTTPLiveStream(hlsutils.HLSParameters{
		TempDir:     config.Get().Storage.TempPath,
		TargetPath:  fullPath.String(),
		Type:        hlsutils.MediaTypeMusicVideo,
		WebPlayback: ctx.MZPlay.WebPlayback,
		MetaData: metadata.LoadMusicVideoMetadata(metadata.Context{
			Type:                  mvSrc,
			WebPlayback:           ctx.MZPlay.WebPlayback,
			AppleMusicMusicVideos: ctx.AppleMusic.MusicVideos,
			AppleMusicAlbum:       ctx.AppleMusic.Albums,
			ItunesMusicVideo:      ctx.iTunes.MusicVideo,
			CoverData:             coverData,
		}),
		IsEncrypted: true,
	})
	if err = context.Execute(); err == nil {
		LOG.Info.Printf("Download completed, saved to: %s", fullPath.String())
	}
	return
}

func (d *Downloader) DownloadLyrics(trackID string, ctx APIContext, fullPath FullPath) (err error) {
	var ttmlRaw string
	if _, ttmlRaw, err = applemusic.GetSyllableLyrics(trackID); err != nil {
		return err
	}

	lyricsPath := fullPath.AlbumPath("Lyrics", fmt.Sprintf(
		"%d-%d. %s.ttml",
		*ctx.AppleMusic.Songs.Attributes.DiscNumber,
		*ctx.AppleMusic.Songs.Attributes.TrackNumber,
		*ctx.AppleMusic.Songs.Attributes.Name))

	if err = os.MkdirAll(path.Dir(lyricsPath), os.ModePerm); err != nil {
		return
	}

	if err = os.WriteFile(lyricsPath, []byte(ttmlRaw), os.ModePerm); err != nil {
		return
	}

	return
}

var AppleMusicURLPattern = regexp.MustCompile(`^https://(?:beta.)?music.apple.com/(?P<storefront>[a-z]{2})/(?P<catalog_type>[a-z\-]+)/(?:[%0-9A-Za-z\-]+/)?(?P<itunes_id>[0-9]+|p\.[0-9A-Za-z]+)(?:\?(?P<query_strings>.+?))?$`)

func (d *Downloader) Download(targetUrl string) (err error) {
	submatches := utils.FindStringSubmatchMap(AppleMusicURLPattern, targetUrl)
	if submatches == nil || submatches["itunes_id"] == "" {
		return errors.New("invalid targetUrl")
	}
	if submatches["storefront"] != config.Get().AppleMusic.Storefront {
		LOG.Warn.Printf("storefront mismatch, this may cause errors during processing")
	}

	switch submatches["catalog_type"] {
	case "album":
		{
			var values url.Values
			if values, err = url.ParseQuery(submatches["query_strings"]); err != nil {
				return
			}
			if values.Has("i") {
				return d.DownloadSong(values.Get("i"), APIContext{}, FullPath{})
			}
			return d.DownloadAlbum(submatches["itunes_id"], APIContext{}, FullPath{})
		}
	case "song":
		return d.DownloadSong(submatches["itunes_id"], APIContext{}, FullPath{})
	case "music-video":
		return d.DownloadMusicVideo(submatches["itunes_id"], APIContext{}, FullPath{})
	case "artist":
		return fmt.Errorf("unsupport catalog type: %s", submatches["catalog_type"])
	default:
		return errors.New("invalid catalog type")
	}
}

func main() {
	var err error

	if err = config.LoadConfig(); err != nil {
		panic(err)
	}

	if err = api.RefreshToken(); err != nil {
		panic(err)
	}

	// https://music.apple.com/cn/album/waking-up-deluxe-version/1446021230?l=en-GB&ls
	// https://music.apple.com/cn/album/ghost-stories/829909653
	// https://music.apple.com/cn/album/dreaming-out-loud/1489409642
	// https://music.apple.com/cn/album/dreaming-out-loud/1445841529
	// albumID := "820166930"
	//amURL := "https://music.apple.com/cn/album/dreaming-out-loud/1445841529"
	//amURL := "https://music.apple.com/cn/album/1581087024"
	//amURL := "https://music.apple.com/cn/album/mylo-xyloto/726372830"
	//amURL := "https://music.apple.com/cn/album/dispatch-vol-1-original-soundtrack/1858896502"
	//amURL := "https://music.apple.com/cn/album/dispatch-vol-2-original-soundtrack/1858898497"
	//amURL := "https://music.apple.com/cn/album/dispatch-vol-3-original-soundtrack/1858897742"
	//amURL := "https://music.apple.com/cn/album/dispatch-vol-4-original-soundtrack/1858902822"
	//amURL := "https://music.apple.com/cn/album/radio/1640300198?i=1640300200"
	//amURL := "https://music.apple.com/cn/music-video/making-the-record/1446022245?l=en-GB"
	//amURL := "https://music.apple.com/cn/album/feel-the-way-i-do/1599218112?i=1599218120"
	amURL := "https://music.apple.com/cn/album/kissing-someone-else/1533848159?i=1533848809"
	//amURL := "https://music.apple.com/cn/album/%E6%A2%A6%E9%BE%99-loom-tour-2026-%E4%B8%AD%E5%9B%BD%E5%B7%A1%E6%BC%94%E7%89%B9%E5%88%AB%E7%89%88/1879820916?l=en-GB&ls"

	amDownloader := Downloader{
		TargetPath: config.Get().Storage.TargetPath,
	}
	if err = amDownloader.Download(amURL); err != nil {
		panic(err)
	}
}

func main3() {
	if err := config.LoadConfig(); err != nil {
		panic(err)
	}

	data, _ := json.MarshalIndent(config.Get(), "", "  ")
	println(string(data))
}

func main1() {
	_ = api.RefreshToken()
	var uri = "https://aod-ssl.itunes.apple.com/itunes-assets/Music114/v4/90/8a/62/908a623f-e478-642a-47cf-d5e472a15930/mzaf_A1533848809.rphq.aac.wa.m3u8"

	var webPlayback *applemusic.WebPlaybackSong
	var err error
	if webPlayback, err = applemusic.GetWebPlayback("1533848809"); err != nil {
		LOG.Error.Printf("failed to fetch HLS manifest: %v", err)
	}
	var context = hlsutils.NewHTTPLiveStream(hlsutils.HLSParameters{
		TempDir:          config.Get().Storage.TempPath,
		TargetPath:       "Downloads/a.mp4",
		Type:             hlsutils.MediaTypeMusicVideo,
		MediaPlaylistURI: uri,
		WebPlayback:      webPlayback,
		IsEncrypted:      true,
	})
	if err = context.Execute(); err == nil {
		LOG.Info.Printf("Download completed, saved to: %s", "Downloads/a.mp4")
	}
}
