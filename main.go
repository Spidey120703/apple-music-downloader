package main

import (
	"bytes"
	"downloader/api/applemusic"
	"downloader/api/itunes"
	"downloader/drm/fairplay"
	"downloader/drm/widevine"
	"downloader/log"
	"downloader/utils"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/abema/go-mp4"
)

func downloadSong(m4aPath string, itunesSongInfo *itunes.Song, song *applemusic.Songs, album *applemusic.Albums, coverData []byte, lyrics string) error {
	log.Info.Println(strings.Repeat("=", 128))
	log.Info.Printf("Downloading (%d/%d). %s", *song.Attributes.TrackNumber, *album.Attributes.TrackCount, *song.Attributes.Name)

	url, keys, err := handleTrackEnhanceHls(*song.Attributes.ExtendedAssetUrls.EnhancedHls)
	if err != nil {
		return err
	}

	mp4File, err := utils.HttpOpen(url, utils.TempDir)
	if err != nil {
		return err
	}
	defer utils.CloseQuietly(mp4File)

	log.Info.Println("Start extracting MP4 file...")
	alacInfo, err := extractAlac(mp4File)
	if alacInfo == nil || err != nil {
		return err
	}
	log.Info.Println("Extract finished")

	log.Info.Println("Start decrypting ALAC Samples...")
	samples, err := fairplay.DecryptSample(alacInfo.Samples(), *song.ID, keys)
	if err != nil {
		return err
	}
	log.Info.Println("Decrypt finished")

	m4aFile, err := os.Create(m4aPath)
	if err != nil {
		return err
	}
	defer utils.CloseQuietly(m4aFile)

	err = writeM4A(
		mp4.NewWriter(m4aFile),
		alacInfo,
		itunesSongInfo,
		song,
		album,
		utils.PackData(samples),
		coverData,
		lyrics)
	if err != nil {
		return err
	}

	log.Info.Println("Download finished, saved to:", m4aPath)
	return nil
}

func downloadMusicVideo(
	mp4Path string,
	hlsPlaylistURL string,
	itunesMusicVideoInfo *itunes.MusicVideo,
	webPlayback *applemusic.WebPlaybackSong,
	song *applemusic.Songs,
	album *applemusic.Albums,
	coverData []byte,
	token string,
) error {
	log.Info.Println(strings.Repeat("=", 128))
	log.Info.Printf("Downloading (%d/%d). %s", *song.Attributes.TrackNumber, *album.Attributes.TrackCount, *song.Attributes.Name)

	_, videoHlsInfo, audioHlsInfo, err := handleMusicVideoHls(hlsPlaylistURL)
	if err != nil {
		return err
	}

	// Processing video track
	var videoFiles []*os.File
	defer utils.CloseQuietlyAll(videoFiles)

	for _, uri := range videoHlsInfo.Urls {
		file, err := utils.HttpOpen(uri, utils.TempDir)
		if err != nil {
			return err
		}
		videoFiles = append(videoFiles, file)
	}

	var mvInfo MusicVideoInfo

	log.Info.Println("Start decrypting video track...")
	videoData, err := widevine.Decrypt(videoFiles, videoHlsInfo.Keys[widevine.KeyFormatWidevine][0], webPlayback, token)
	if err != nil {
		return err
	}
	log.Info.Println("Decrypt finished")

	err = extractAvc1([]io.ReadSeeker{videoFiles[0], bytes.NewReader(videoData)}, &mvInfo)
	if err != nil {
		return err
	}

	var videoSamples [][]byte
	for _, videoSample := range mvInfo.VideoSamples() {
		videoSamples = append(videoSamples, videoSample.Data)
	}

	mp4File, err := os.Create(mp4Path)
	if err != nil {
		return err
	}
	defer utils.CloseQuietly(mp4File)

	// Processing audio track
	var audioFiles []*os.File
	defer utils.CloseQuietlyAll(audioFiles)

	for _, uri := range audioHlsInfo.Urls {
		file, err := utils.HttpOpen(uri, utils.TempDir)
		if err != nil {
			return err
		}
		audioFiles = append(audioFiles, file)
	}

	err = extractMp4a(audioFiles, &mvInfo)
	if err != nil {
		return err
	}

	log.Info.Println("Start decrypting AAC Samples...")
	audioSamples, err := fairplay.DecryptSample(mvInfo.AudioSamples(), *song.ID, audioHlsInfo.Keys[fairplay.KeyFormatFairPlay])
	if err != nil {
		return err
	}
	log.Info.Println("Decrypt finished")

	err = writeMP4(
		mp4.NewWriter(mp4File),
		&mvInfo,
		itunesMusicVideoInfo,
		song,
		album,
		utils.PackData(videoSamples),
		utils.PackData(audioSamples),
		coverData)
	if err != nil {
		return err
	}

	log.Info.Println("Download finished, saved to:", mp4Path)

	return nil
}

func downloadAlbum(targetPath string, id string, token string) error {
	album, err := applemusic.GetAlbumData(id, token)
	if err != nil {
		return err
	}

	artistDir := utils.SanitizePath(*album.Attributes.ArtistName)
	albumDir := fmt.Sprintf(
		"%s - %s [%s]",
		*album.Attributes.ReleaseDate,
		utils.SanitizePath(*album.Attributes.Name),
		*album.Attributes.Upc)

	coverPath := path.Join(targetPath, artistDir, albumDir, "Cover")

	coverData, err := ReadCover(*album.Attributes.Artwork, coverPath)
	if err != nil {
		return err
	}

	{ // Extras

		var downloadArtwork = func(artwork *applemusic.Artwork, name string) (err error) {
			if artwork == nil {
				return nil
			}
			artworkPath := path.Join(targetPath, artistDir, albumDir, "Extras", "Artworks", name)

			err = os.MkdirAll(path.Dir(artworkPath), os.ModePerm)
			if err != nil {
				return err
			}

			_, err = DownloadArtwork(*artwork, artworkPath)
			return err
		}

		var downloadMotionVideo = func(motionVideo *applemusic.MotionVideo, name string) (err error) {
			if motionVideo == nil {
				return nil
			}
			motionVideoPath := path.Join(targetPath, artistDir, albumDir, "Extras", "MotionVideos", name)

			err = os.MkdirAll(path.Dir(motionVideoPath), os.ModePerm)
			if err != nil {
				return err
			}

			_, _, err = DownloadMotionVideo(*motionVideo, motionVideoPath)
			return err
		}

		if album.Attributes.EditorialArtwork != nil {
			err = downloadArtwork(album.Attributes.EditorialArtwork.BannerUber, "BannerUber")
			if err != nil {
				return err
			}
			err = downloadArtwork(album.Attributes.EditorialArtwork.OriginalFlowcaseBrick, "OriginalFlowcaseBrick")
			if err != nil {
				return err
			}
			err = downloadArtwork(album.Attributes.EditorialArtwork.StaticDetailSquare, "StaticDetailSquare")
			if err != nil {
				return err
			}
			err = downloadArtwork(album.Attributes.EditorialArtwork.StaticDetailTall, "StaticDetailTall")
			if err != nil {
				return err
			}
			err = downloadArtwork(album.Attributes.EditorialArtwork.StoreFlowcase, "StoreFlowcase")
			if err != nil {
				return err
			}
			err = downloadArtwork(album.Attributes.EditorialArtwork.SubscriptionHero, "SubscriptionHero")
			if err != nil {
				return err
			}
			err = downloadArtwork(album.Attributes.EditorialArtwork.SuperHeroTall, "SuperHeroTall")
			if err != nil {
				return err
			}
		}

		if album.Attributes.EditorialVideo != nil {
			err = downloadMotionVideo(album.Attributes.EditorialVideo.MotionSquareVideo1X1, "MotionSquareVideo1X1")
			if err != nil {
				return err
			}
			err = downloadMotionVideo(album.Attributes.EditorialVideo.MotionDetailSquare, "MotionDetailSquare")
			if err != nil {
				return err
			}
			err = downloadMotionVideo(album.Attributes.EditorialVideo.MotionDetailTall, "MotionDetailTall")
			if err != nil {
				return err
			}
		}

	}

	for _, track := range album.Relationships.Tracks.Data[20:] {
		//_ = json.NewEncoder(os.Stdout).Encode(track)
		discDir := fmt.Sprintf("Disc %d", *track.Attributes.DiscNumber)

		switch *track.Type {
		case "songs":

			lyrics := ""
			// Handling Lyrics
			if *track.Attributes.HasLyrics {
				ttml, err := applemusic.GetLyrics(*track.ID, token)
				if err != nil {
					return err
				}
				lyrics, err = extractTTML(ttml)
				if err != nil {
					return err
				}

				_, ttml, err = applemusic.GetSyllableLyrics(*track.ID, token)
				if err != nil {
					return err
				}

				lyricsName := fmt.Sprintf(
					"%d-%d. %s.ttml",
					*track.Attributes.DiscNumber,
					*track.Attributes.TrackNumber,
					utils.SanitizePath(*track.Attributes.Name))
				lyricsPath := path.Join(targetPath, artistDir, albumDir, "Lyrics", lyricsName)

				err = os.MkdirAll(path.Dir(lyricsPath), os.ModePerm)
				if err != nil {
					return err
				}

				err = os.WriteFile(lyricsPath, []byte(ttml), os.ModePerm)
				if err != nil {
					return err
				}
			}

			trackName := fmt.Sprintf(
				"%d. %s.m4a",
				*track.Attributes.TrackNumber,
				utils.SanitizePath(*track.Attributes.Name))
			m4aPath := path.Join(targetPath, artistDir, albumDir, discDir, trackName)

			err = os.MkdirAll(path.Dir(m4aPath), os.ModePerm)
			if err != nil {
				return err
			}

			info, err := GetITunesInfo[itunes.Song](*track.ID, "song", token)
			if err != nil {
				return err
			}

			err = downloadSong(m4aPath, info, &track, album, coverData, lyrics)
			if err != nil {
				return err
			}

		case "music-videos":

			trackName := fmt.Sprintf(
				"%d. %s.mp4",
				*track.Attributes.TrackNumber,
				utils.SanitizePath(*track.Attributes.Name))
			mp4Path := path.Join(targetPath, artistDir, albumDir, discDir, trackName)

			err = os.MkdirAll(path.Dir(mp4Path), os.ModePerm)
			if err != nil {
				return err
			}

			info, err := GetITunesInfo[itunes.MusicVideo](*track.ID, "song", token)
			if err != nil {
				return err
			}

			webPlayback, err := applemusic.GetWebPlayback(*track.ID, token)
			if err != nil {
				return err
			}
			hlsPlaylistURL, err := fixURLParamLanguage(webPlayback.HlsPlaylistURL)
			if err != nil {
				return err
			}

			err = downloadMusicVideo(mp4Path, hlsPlaylistURL, info, webPlayback, &track, album, coverData, token)
			if err != nil {
				return err
			}

		default:
			log.Warn.Printf("Type '%s' is not available to download", *track.Type)
		}

	}

	return nil
}

func main() {
	const TargetPath = "Downloads"

	token, err := applemusic.GetToken()
	if err != nil {
		panic(err)
	}

	albumIds := []string{
		//"1592027684",
		//"1781153619",
		//"1445841529",
		//"1641821680",
		//"829909653",
		//"1440831203",
		"1770789137",
	}

	for _, id := range albumIds {
		err := downloadAlbum(TargetPath, id, token)
		if err != nil {
			panic(err)
		}
	}
}
