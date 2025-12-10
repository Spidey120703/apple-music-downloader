package main

import (
	"downloader/api/applemusic"
	"downloader/config"
	"downloader/m3u8/hlsutils"
	"downloader/mp4/metadata"
	"downloader/utils"
	"io"
	"os"
	"path"
	"strconv"
)

func ReadCover(data applemusic.Artwork, coverPath string) ([]byte, error) {
	var err error
	if coverPath, err = DownloadArtwork(data, coverPath); err != nil {
		return nil, err
	}

	var file *os.File
	if file, err = os.Open(coverPath); err != nil {
		return nil, err
	}
	defer utils.CloseQuietly(file)

	return io.ReadAll(file)
}

func DownloadArtwork(data applemusic.Artwork, artworkPath string) (string, error) {
	var err error
	URL := *data.URL

	URL, artworkPath = GetFormattedImageURLName(URL, artworkPath, map[string]string{
		"w": strconv.Itoa(*data.Width),
		"h": strconv.Itoa(*data.Height),
	})

	if artworkPath, err = utils.DownloadFile(URL, artworkPath); err != nil && !os.IsExist(err) {
		return "", err
	}
	return artworkPath, nil
}

func DownloadMotionVideo(data applemusic.MotionVideo, videoPath string) (string, error) {
	previewData, err := ReadCover(*data.PreviewFrame, path.Join(config.TempPath, FilenameFormatUUID))
	if err != nil {
		return "", err
	}

	hls := hlsutils.NewHTTPLiveStream(hlsutils.HLSParameters{
		MasterPlaylistURI: *data.Video,
		TempDir:           config.TempPath,
		TargetPath:        videoPath,
		MetaData: &metadata.Metadata{
			Cover: previewData,
		},
	})
	if err = hls.Execute(); err != nil {
		panic(err)
	}

	return videoPath, nil
}
