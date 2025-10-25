package main

import (
	"errors"
	"net/http"
	"path/filepath"
	"regexp"
	"slices"

	"github.com/Eyevinn/hls-m3u8/m3u8"
)

func ReadM3U8(url string) (playlist m3u8.Playlist, listType m3u8.ListType, err error) {
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer CloseQuietly(resp.Body)
	return m3u8.DecodeFrom(resp.Body, true)
}

const (
	PrefetchKeyUri = "skd://itunes.apple.com/P000000000/s1/e1"
	DefaultId      = "0"
)

func isValidKey(uri *m3u8.Key) bool {
	if uri.Keyformat != "com.apple.streamingkeydelivery" {
		return false
	}
	matched, err := regexp.MatchString(`^skd://(?:itunes.apple.com/)?[0-9A-Za-z/%\-_]+$`, uri.URI)
	if err != nil {
		return false
	}
	//return matched && (strings.HasSuffix(uri, "/c23") || strings.HasSuffix(uri, "/c6"))
	return matched
}

var AlacCodecs = []string{
	"audio-alac-stereo-44100-16",
	"audio-alac-stereo-44100-24",
	"audio-alac-stereo-48000-24",
	"audio-alac-stereo-88200-24",
	"audio-alac-stereo-96000-24",
	"audio-alac-stereo-176400-24",
	"audio-alac-stereo-192000-24",
}

func handleTrackEnhanceHls(enhancedHlsM3U8Url string) (url string, keys []string, err error) {

	var handleMasterM3U8 = func(m3u8Url string) (uri string, err error) {
		Info.Printf("Located master playlist (HLS m3u8): %s", m3u8Url)
		playlist, listType, err := ReadM3U8(m3u8Url)
		if err != nil {
			return
		}

		switch listType {
		case m3u8.MASTER:
			masterPlaylist := playlist.(*m3u8.MasterPlaylist)
			var groupId string
			var flag = -1
			for _, alternative := range masterPlaylist.GetAllAlternatives() {
				index := slices.Index(AlacCodecs, alternative.GroupId)
				if index > flag {
					groupId = alternative.GroupId
					flag = index
				}
			}
			if len(groupId) == 0 {
				return "", errors.New("codec alac not found")
			}
			for _, variant := range masterPlaylist.Variants {
				if variant.Audio == groupId {
					uri = variant.URI
					break
				}
			}
		case m3u8.MEDIA:
			return "", errors.New("inappropriate m3u8 type")
		}

		if len(uri) == 0 {
			return "", errors.New("alac track not found")
		}

		return
	}

	var handleMediaM3U8 = func(m3u8Url string, keys *[]string) (uri string, err error) {
		Info.Printf("Located media playlist (HLS m3u8) URL: %s", m3u8Url)
		playlist, listType, err := ReadM3U8(m3u8Url)
		if err != nil {
			return
		}

		switch listType {
		case m3u8.MASTER:
			return "", errors.New("inappropriate m3u8 type")
		case m3u8.MEDIA:
			mediaPlaylist := playlist.(*m3u8.MediaPlaylist)
			uri = mediaPlaylist.Segments[0].URI
			for i, segment := range mediaPlaylist.Segments {
				if uint(i) >= mediaPlaylist.Count() || segment == nil {
					break
				}
				for _, key := range segment.Keys {
					if isValidKey(&key) && !slices.Contains(*keys, key.URI) {
						Info.Printf("Found URI Key: %s", key.URI)
						*keys = append(*keys, key.URI)
					}
				}
			}
		}
		if len(uri) == 0 {
			return "", errors.New("alac track not found")
		}

		return
	}

	baseUrl, _ := filepath.Split(enhancedHlsM3U8Url)

	mediaM3U8Uri, err := handleMasterM3U8(enhancedHlsM3U8Url)
	if err != nil {
		return "", keys, err
	}

	mp4Uri, err := handleMediaM3U8(baseUrl+mediaM3U8Uri, &keys)
	if err != nil {
		return "", keys, err
	}
	url = baseUrl + mp4Uri

	Info.Printf("Located media URL: %s", url)
	return
}

func handleVideoM3U8(videoM3U8Url string) (url string, err error) {

	var handleMasterM3U8 = func(m3u8Url string) (uri string, err error) {
		Info.Printf("Located master playlist (HLS m3u8) URL: %s", m3u8Url)
		playlist, listType, err := ReadM3U8(m3u8Url)
		if err != nil {
			return
		}

		switch listType {
		case m3u8.MASTER:
			masterPlaylist := playlist.(*m3u8.MasterPlaylist)
			for _, variant := range masterPlaylist.Variants {
				uri = variant.URI
			}
		case m3u8.MEDIA:
			return "", errors.New("inappropriate m3u8 type")
		}

		if len(uri) == 0 {
			return "", errors.New("alac track not found")
		}

		return
	}

	var handleMediaM3U8 = func(m3u8Url string) (uri string, err error) {
		Info.Printf("Located media playlist (HLS m3u8) URL: %s", m3u8Url)
		playlist, listType, err := ReadM3U8(m3u8Url)
		if err != nil {
			return
		}

		switch listType {
		case m3u8.MASTER:
			return "", errors.New("inappropriate m3u8 type")
		case m3u8.MEDIA:
			mediaPlaylist := playlist.(*m3u8.MediaPlaylist)
			uri = mediaPlaylist.Map.URI
		}
		if len(uri) == 0 {
			return "", errors.New("alac track not found")
		}

		return
	}

	mediaM3U8Url, err := handleMasterM3U8(videoM3U8Url)
	if err != nil {
		return "", err
	}

	baseUrl, _ := filepath.Split(mediaM3U8Url)

	mp4Uri, err := handleMediaM3U8(mediaM3U8Url)
	if err != nil {
		return "", err
	}

	url = baseUrl + mp4Uri

	Info.Printf("Located media URL: %s", url)
	return
}

func handleMusicVideoHls(masterM3U8Url string) (metaData map[string]string, videoUrls []string, audioUrls []string, videoKeys []string, audioKeys []string, err error) {

	var handleMasterM3U8 = func(m3u8Url string) (videoUri string, audioUri string, meta map[string]string, err error) {
		Info.Printf("Located master playlist (HLS m3u8) URL: %s", m3u8Url)
		playlist, listType, err := ReadM3U8(m3u8Url)
		if err != nil {
			return
		}

		switch listType {
		case m3u8.MASTER:
			masterPlaylist := playlist.(*m3u8.MasterPlaylist)

			meta = make(map[string]string)
			for _, data := range masterPlaylist.SessionDatas {
				meta[data.DataId] = data.Value
			}

			var audioGroupId string
			for _, variant := range masterPlaylist.Variants {
				if variant.Iframe {
					continue
				}
				videoUri = variant.URI
				audioGroupId = variant.Audio
			}

			for _, alter := range masterPlaylist.GetAllAlternatives() {
				if alter.Type != "AUDIO" {
					continue
				}
				if audioGroupId == alter.GroupId {
					audioUri = alter.URI
				}
			}
		case m3u8.MEDIA:
			return "", "", nil, errors.New("inappropriate m3u8 type")
		}

		if len(videoUri) == 0 || len(audioUri) == 0 {
			return videoUri, audioUri, meta, errors.New("media m3u8 not found")
		}

		return
	}

	var handleMediaM3U8 = func(m3u8Url string, keys *[]string) (urls []string, err error) {
		Info.Printf("Located media playlist (HLS m3u8) URL: %s", m3u8Url)
		playlist, listType, err := ReadM3U8(m3u8Url)
		if err != nil {
			return
		}

		switch listType {
		case m3u8.MASTER:
			return nil, errors.New("inappropriate m3u8 type")
		case m3u8.MEDIA:
			mediaPlaylist := playlist.(*m3u8.MediaPlaylist)

			urls = append(urls, mediaPlaylist.Map.URI)
			for i, segment := range mediaPlaylist.Segments {
				if uint(i) >= mediaPlaylist.Count() || segment == nil {
					break
				}
				urls = append(urls, segment.URI)
				for _, key := range segment.Keys {
					if isValidKey(&key) && !slices.Contains(*keys, key.URI) {
						Info.Printf("Found URI Key: %s", key.URI)
						*keys = append(*keys, key.URI)
					}
				}
			}
		}
		if len(urls) == 0 {
			return nil, errors.New("file not found")
		}

		return
	}

	videoM3U8Uri, audioM3U8Uri, metaData, err := handleMasterM3U8(masterM3U8Url)
	if err != nil {
		return
	}

	var baseUrl string

	baseUrl, _ = filepath.Split(videoM3U8Uri)
	videoUrls, err = handleMediaM3U8(videoM3U8Uri, &videoKeys)
	if err != nil {
		return
	}
	for i, url := range videoUrls {
		videoUrls[i] = baseUrl + url
	}

	baseUrl, _ = filepath.Split(audioM3U8Uri)
	audioUrls, err = handleMediaM3U8(audioM3U8Uri, &audioKeys)
	if err != nil {
		return
	}
	for i, url := range audioUrls {
		audioUrls[i] = baseUrl + url
	}

	return
}
