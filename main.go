package main

import (
	"downloader/LOG"
	"downloader/api/applemusic"
	"downloader/api/itunes"
	"downloader/config"
	"downloader/m3u8/hlsutils"
	"downloader/mp4/metadata"
	ttml2 "downloader/ttml"
	"downloader/utils"
	"fmt"
	"os"
	"path"
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

	if ctx.AlbumCoverData, err = ReadCover(*ctx.AppleMusic.Albums.Attributes.Artwork, fullPath.AlbumPath("Cover")); err != nil {
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
			_, err = DownloadArtwork(*artwork, fullPath.AlbumPath("Extras", "Artworks", FilenameFormatOriginalFileName))
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
			_, err = DownloadMotionVideo(*motionVideo, fullPath.AlbumPath("Extras", "MotionVideos", name))
		}
	}

	LOG.Info.Printf("Start to download %d tracks\n", len(ctx.AppleMusic.Albums.Relationships.Tracks.Data))

	for _, track := range ctx.AppleMusic.Albums.Relationships.Tracks.Data {
		LOG.Info.Println(strings.Repeat("=", 128))
		LOG.Info.Printf("Downloading track: %d-%d %s", *track.Attributes.DiscNumber, *track.Attributes.TrackNumber, *track.Attributes.Name)
		LOG.Info.Println("Track Info:")
		LOG.Info.Printf("\t\t%16s:  %s", "Track Title", *track.Attributes.Name)
		LOG.Info.Printf("\t\t%16s:  %s", "Artist Name", *track.Attributes.ArtistName)
		LOG.Info.Printf("\t\t%16s:  %d", "Disc Number", *track.Attributes.DiscNumber)
		LOG.Info.Printf("\t\t%16s:  %d", "Track Number", *track.Attributes.TrackNumber)
		LOG.Info.Printf("\t\t%16s:  %s", "ISRC", *track.Attributes.Isrc)
		if track.Attributes.WorkName != nil {
			LOG.Info.Printf("\t\t%16s:  %s", "Work Name", *track.Attributes.WorkName)
		}
		LOG.Info.Printf("\t\t%16s:  %s", "Genre Names", strings.Join(track.Attributes.GenreNames, ", "))
		LOG.Info.Println()

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
					LOG.Info.Println(strings.Repeat("=", 128))
					LOG.Info.Printf("Downloading music videos: %s [%s]", *track.Attributes.Name, *track.Attributes.Isrc)
					LOG.Info.Println("Track Info:")
					LOG.Info.Printf("\t\t%16s:  %s", "Track Title", *track.Attributes.Name)
					LOG.Info.Printf("\t\t%16s:  %s", "Artist Name", *track.Attributes.ArtistName)
					LOG.Info.Printf("\t\t%16s:  %s", "ISRC", *track.Attributes.Isrc)
					LOG.Info.Println()
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

	fullPath.TrackName = fmt.Sprintf(
		"%d. %s",
		*ctx.AppleMusic.Songs.Attributes.TrackNumber,
		*ctx.AppleMusic.Songs.Attributes.Name)
	fullPath.Ext = ExtM4A

	if ctx.iTunes.Song == nil {
		if ctx.iTunes.Song, err = itunes.GetITunesInfo[itunes.Song](trackID, "song"); err != nil {
			return
		}
	}

	if ctx.MZPlay.WebPlayback == nil {
		if ctx.MZPlay.WebPlayback, err = applemusic.GetWebPlayback(trackID); err != nil {
			return
		}
	}

	var ttml, lyrics string
	if *ctx.AppleMusic.Songs.Attributes.HasLyrics {
		if err = d.DownloadLyrics(trackID, ctx, fullPath); err != nil {
			return
		}
		if ttml, err = applemusic.GetLyrics(trackID); err != nil {
			return
		}
		if lyrics, err = ttml2.ExtractTextFromTTML(ttml); err != nil {
			return
		}
	}

	var context = hlsutils.NewHTTPLiveStream(hlsutils.HLSParameters{
		TempDir:           config.TempPath,
		TargetPath:        fullPath.String(),
		Type:              hlsutils.MediaTypeSong,
		AdamID:            trackID,
		MasterPlaylistURI: *ctx.AppleMusic.Songs.Attributes.ExtendedAssetUrls.EnhancedHls,
		MetaData: metadata.LoadSongMetadata(metadata.Context{
			WebPlayback:     ctx.MZPlay.WebPlayback,
			AppleMusicSongs: ctx.AppleMusic.Songs,
			AppleMusicAlbum: ctx.AppleMusic.Albums,
			ItunesSong:      ctx.iTunes.Song,
			CoverData:       ctx.AlbumCoverData,
			LyricsData:      lyrics,
		}),
		IsEncrypted: true,
	})
	if err = context.Execute(); err == nil {
		LOG.Info.Printf("Download completed, saved to: %s", fullPath.String())
	}
	return
}

func (d *Downloader) DownloadMusicVideo(trackID string, ctx APIContext, fullPath FullPath) (err error) {
	var mvType metadata.MusicVideoType
	if len(ctx.AppleMusic.MusicVideos.Relationships.Albums.Data) > 0 {
		mvType = metadata.MusicVideoTypeAlbumRelated
		fullPath.TrackName = fmt.Sprintf(
			"%d. %s",
			*ctx.AppleMusic.MusicVideos.Attributes.TrackNumber,
			*ctx.AppleMusic.MusicVideos.Attributes.Name)
	} else {
		mvType = metadata.MusicVideoTypeSongsRelated
		fullPath.TrackName = fmt.Sprintf(
			"%s [%s]",
			*ctx.AppleMusic.MusicVideos.Attributes.Name,
			*ctx.AppleMusic.MusicVideos.Attributes.Isrc)
	}
	fullPath.Ext = ExtM4V

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

	artwork := ctx.AppleMusic.MusicVideos.Attributes.Artwork

	var coverData []byte
	if artwork != nil {
		if coverData, err = ReadCover(*artwork, path.Join(config.TempPath, FilenameFormatUUID)); err != nil {
			return
		}
	}

	var context = hlsutils.NewHTTPLiveStream(hlsutils.HLSParameters{
		TempDir:     config.TempPath,
		TargetPath:  fullPath.String(),
		Type:        hlsutils.MediaTypeMusicVideo,
		WebPlayback: ctx.MZPlay.WebPlayback,
		MetaData: metadata.LoadMusicVideoMetadata(metadata.Context{
			Type:                  mvType,
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
	var ttml string
	if _, ttml, err = applemusic.GetSyllableLyrics(trackID); err != nil {
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

	if err = os.WriteFile(lyricsPath, []byte(ttml), os.ModePerm); err != nil {
		return
	}

	return
}

func main() {
	var err error
	applemusic.RefreshToken()

	// https://music.apple.com/cn/album/waking-up-deluxe-version/1446021230?l=en-GB&ls
	// https://music.apple.com/cn/album/ghost-stories/829909653
	// https://music.apple.com/cn/album/dreaming-out-loud/1489409642
	// https://music.apple.com/cn/album/dreaming-out-loud/1445841529
	albumID := "1446021230"

	downloader := Downloader{
		TargetPath: config.TargetPath,
	}
	if err = downloader.DownloadAlbum(albumID, APIContext{}, FullPath{}); err != nil {
		panic(err)
	}
}

func main1() {
	applemusic.RefreshToken()
	ID := "1446009426"

	webPlayback, err := applemusic.GetWebPlayback(ID)
	if err != nil {
		panic(err)
	}

	hls := hlsutils.NewHTTPLiveStream(hlsutils.HLSParameters{
		Type:        hlsutils.MediaTypeMusicVideo,
		WebPlayback: webPlayback,
		TempDir:     config.TempPath,
		TargetPath:  path.Join(config.TempPath, "a.m4v"),
		IsEncrypted: true,
	})
	if err = hls.Execute(); err != nil {
		panic(err)
	}
}
