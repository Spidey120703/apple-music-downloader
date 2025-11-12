package main

import (
	"downloader/api/applemusic"
	"downloader/utils"
	"io"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
)

func ReadCover(data applemusic.Artwork, coverPath string) ([]byte, error) {
	coverPath, err := DownloadArtwork(data, coverPath)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(coverPath)
	if err != nil {
		return nil, err
	}
	defer utils.CloseQuietly(file)

	return io.ReadAll(file)
}

func DownloadArtwork(data applemusic.Artwork, artworkPath string) (string, error) {
	URL := *data.URL
	URL = strings.Replace(URL, "{w}", strconv.Itoa(*data.Width), 1)
	URL = strings.Replace(URL, "{h}", strconv.Itoa(*data.Height), 1)

	splited := strings.Split(URL, "/")
	ext := path.Ext(splited[len(splited)-2])
	URL = URL[:strings.LastIndexByte(URL, '.')] + ext
	artworkPath = artworkPath + ext

	_, err := utils.DownloadFile(URL, artworkPath)
	if err != nil && !os.IsExist(err) {
		return "", err
	}
	return artworkPath, nil
}

func DownloadMotionVideo(data applemusic.MotionVideo, videoPath string) (string, string, error) {
	previewPath, err := DownloadArtwork(*data.PreviewFrame, videoPath)
	if err != nil {
		return "", "", err
	}

	videoUrl, err := handleVideoM3U8(*data.Video)
	if err != nil {
		return previewPath, "", err
	}

	videoPath, err = utils.DownloadFile(videoUrl, videoPath+".mp4")

	return previewPath, videoPath, nil
}

func fixURLParamLanguage(playlistURL string) (string, error) {
	parse, err := url.Parse(playlistURL)
	if err != nil {
		return playlistURL, err
	}
	query := parse.Query()
	query.Set("l", "en")
	parse.RawQuery = query.Encode()

	return parse.String(), nil
}
