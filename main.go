package main

import (
	"downloader/applemusic"
	"downloader/itunes"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/abema/go-mp4"
)

func downloadSong(m4aPath string, itunesSongInfo *itunes.Song, song *applemusic.Songs, album *applemusic.Albums, coverData []byte, lyrics string) error {
	Info.Println(strings.Repeat("=", 128))
	Info.Printf("Downloading (%d/%d). %s", *song.Attributes.TrackNumber, *album.Attributes.TrackCount, *song.Attributes.Name)

	url, keys, err := handleTrackEnhanceHls(*song.Attributes.ExtendedAssetUrls.EnhancedHls)
	if err != nil {
		return err
	}

	mp4File, err := HttpOpen(url, TempDir)
	if err != nil {
		return err
	}
	defer CloseQuietly(mp4File)

	Info.Println("Start extracting MP4 file...")
	alacInfo, err := extractAlac(mp4File)
	if alacInfo == nil || err != nil {
		return err
	}
	Info.Println("Extract finished")

	Info.Println("Start decrypting ALAC samples...")
	samples, err := decryptSample(alacInfo.Samples(), song, keys)
	if err != nil {
		return err
	}
	Info.Println("Decrypt finished")

	m4aFile, err := os.Create(m4aPath)
	if err != nil {
		return err
	}
	defer CloseQuietly(m4aFile)

	err = writeM4A(
		mp4.NewWriter(m4aFile),
		alacInfo,
		itunesSongInfo,
		song,
		album,
		PackData(samples),
		coverData,
		lyrics)
	if err != nil {
		return err
	}

	Info.Println("Download finished, saved to:", m4aPath)
	return nil
}

// todo: something wrong with the samples decrypting, video track still cannot decrypt
func downloadMusicVideo(mp4Path string, hlsPlaylistURL string, itunesMusicVideoInfo *itunes.MusicVideo, song *applemusic.Songs, album *applemusic.Albums, coverData []byte) error {
	Info.Println(strings.Repeat("=", 128))
	Info.Printf("Downloading (%d/%d). %s", *song.Attributes.TrackNumber, *album.Attributes.TrackCount, *song.Attributes.Name)

	metaData, videoUrls, audioUrls, videoKeys, audioKeys, _ := handleMusicVideoHls(hlsPlaylistURL)
	println(metaData, audioUrls, audioKeys)

	var videoFiles []*os.File
	defer CloseQuietlyAll(videoFiles)

	for _, uri := range videoUrls {
		file, err := HttpOpen(uri, TempDir)
		if err != nil {
			return err
		}
		videoFiles = append(videoFiles, file)
	}

	var videoInfo VideoInfo
	err := extractAvc(videoFiles, &videoInfo)
	if err != nil {
		return err
	}

	Info.Println("Start decrypting AVC samples...")
	videoSamples, err := decryptSample(videoInfo.VideoSamples(), song, videoKeys)
	if err != nil {
		return err
	}
	Info.Println("Decrypt finished")

	mp4File, err := os.Create(mp4Path)
	if err != nil {
		return err
	}
	defer CloseQuietly(mp4File)

	var audioFiles []*os.File
	defer CloseQuietlyAll(audioFiles)

	for _, uri := range audioUrls {
		file, err := HttpOpen(uri, TempDir)
		if err != nil {
			return err
		}
		audioFiles = append(audioFiles, file)
	}

	err = extractMp4a(audioFiles, &videoInfo)
	if err != nil {
		return err
	}

	Info.Println("Start decrypting AAC samples...")
	audioSamples, err := decryptSample(videoInfo.AudioSamples(), song, audioKeys)
	if err != nil {
		return err
	}
	Info.Println("Decrypt finished")

	err = writeMP4(
		mp4.NewWriter(mp4File),
		&videoInfo,
		itunesMusicVideoInfo,
		song,
		album,
		PackData(videoSamples),
		PackData(audioSamples),
		coverData)
	if err != nil {
		return err
	}

	Info.Println("Download finished, saved to:", mp4Path)

	return nil
}

func downloadAlbum(targetPath string, id string, token string) error {
	album, err := GetAlbumData(id, token)
	if err != nil {
		return err
	}

	artistDir := SanitizePath(*album.Attributes.ArtistName)
	albumDir := fmt.Sprintf(
		"%s - %s [%s]",
		*album.Attributes.ReleaseDate,
		SanitizePath(*album.Attributes.Name),
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

	for _, track := range album.Relationships.Tracks.Data[19:] {
		//_ = json.NewEncoder(os.Stdout).Encode(track)
		discDir := fmt.Sprintf("Disc %d", *track.Attributes.DiscNumber)

		switch *track.Type {
		case "songs":

			lyrics := ""
			// Handling Lyrics
			if *track.Attributes.HasLyrics {
				ttml, err := GetLyrics(*track.ID, token)
				if err != nil {
					return err
				}
				lyrics, err = extractTTML(ttml)
				if err != nil {
					return err
				}

				_, ttml, err = GetSyllableLyrics(*track.ID, token)
				if err != nil {
					return err
				}

				lyricsName := fmt.Sprintf(
					"%d-%d. %s.ttml",
					*track.Attributes.DiscNumber,
					*track.Attributes.TrackNumber,
					SanitizePath(*track.Attributes.Name))
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
				SanitizePath(*track.Attributes.Name))
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
				SanitizePath(*track.Attributes.Name))
			mp4Path := path.Join(targetPath, artistDir, albumDir, discDir, trackName)

			err = os.MkdirAll(path.Dir(mp4Path), os.ModePerm)
			if err != nil {
				return err
			}

			info, err := GetITunesInfo[itunes.MusicVideo](*track.ID, "song", token)
			if err != nil {
				return err
			}

			hlsPlaylistURL, err := GetMusicVideo("1770791066", token)
			if err != nil {
				return err
			}
			hlsPlaylistURL, err = fixURLParamLanguage(hlsPlaylistURL)
			if err != nil {
				return err
			}

			err = downloadMusicVideo(mp4Path, hlsPlaylistURL, info, &track, album, coverData)
			if err != nil {
				return err
			}

		default:
			Warn.Printf("Type '%s' is not available to download", *track.Type)
		}

	}

	return nil
}

func main() {
	const TargetPath = "Downloads"

	token, err := GetToken()
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
