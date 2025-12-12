package applemusic

import (
	"downloader/internal/config"
	"downloader/pkg/utils"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

var ErrEntityNotFound = func(t string) error {
	return fmt.Errorf("%s not found", t)
}

func GetAlbumData(id string) (*Albums, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		"https://amp-api.music.apple.com/v1/catalog/"+config.Storefront+"/albums/"+id,
		nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", Authorization)
	req.Header.Set("User-Agent", config.UserAgent)
	req.Header.Set("Origin", config.Origin)
	req.Header.Set("Referer", config.Referer)

	query := req.URL.Query()
	query.Set("extend", "artistBio,bornOrFormed,editorialArtwork,editorialNotes,editorialVideo,extendedAssetUrls,hero,isGroup,offers,origin,plainEditorialNotes,seoDescription,seoTitle,artistUrl,contentRating")
	query.Set("include", "albums,record-labels,artists,persons,bands,composers,credits,lyrics,songs,music-videos,tracks,genres")
	query.Set("include[playlists]", "curator")
	query.Set("include[artists]", "albums,genres")
	query.Set("include[music-videos]", "artists,albums,credits,genres")
	query.Set("include[songs]", "artists,albums,composers,credits,music-videos,genres")
	query.Set("l", "zh-Hans-CN")
	query.Set("meta[albums:tracks]", "popularity")
	query.Set("platform", "web")
	req.URL.RawQuery = query.Encode()

	do, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	data := new(struct {
		Errors []Errors `json:"errors,omitempty"`
		Data   []Albums `json:"data,omitempty"`
	})

	if err = json.NewDecoder(do.Body).Decode(&data); err != nil {
		return nil, err
	}
	if len(data.Errors) > 0 {
		return nil, errors.New(data.Errors[0].Detail)
	}
	if len(data.Data) != 1 {
		return nil, ErrEntityNotFound("albums")
	}

	return &data.Data[0], nil
}

func GetLyrics(id string) (string, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		"https://amp-api.music.apple.com/v1/catalog/"+config.Storefront+"/songs/"+id+"/lyrics",
		nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", Authorization)
	req.Header.Set("User-Agent", config.UserAgent)
	req.Header.Set("Origin", config.Origin)
	req.Header.Set("Referer", config.Referer)
	req.Header.Set("Media-User-Token", config.MediaUserToken)

	do, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	var data struct {
		Errors []Errors `json:"errors,omitempty"`
		Data   []Lyrics `json:"Data,omitempty"`
	}

	defer utils.CloseQuietly(do.Body)

	if err = json.NewDecoder(do.Body).Decode(&data); err != nil {
		return "", err
	}
	if len(data.Errors) > 0 {
		return "", errors.New(data.Errors[0].Detail)
	}
	if len(data.Data) == 0 {
		return "", ErrEntityNotFound("lyrics")
	}

	return *data.Data[0].Attributes.Ttml, nil
}

func GetSyllableLyrics(id string) (string, string, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		"https://amp-api.music.apple.com/v1/catalog/cn/songs/"+id+"/syllable-lyrics",
		nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Authorization", Authorization)
	req.Header.Set("User-Agent", config.UserAgent)
	req.Header.Set("Origin", config.Origin)
	req.Header.Set("Referer", config.Referer)
	req.Header.Set("Media-User-Token", config.MediaUserToken)

	query := req.URL.Query()
	query.Set("l[lyrics]", "zh-hans-cn")
	query.Set("l[script]", "zh-Hans")
	query.Set("extend", "ttml,ttmlLocalizations")
	req.URL.RawQuery = query.Encode()

	do, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", err
	}

	var data struct {
		Errors []Errors `json:"errors,omitempty"`
		Data   []Lyrics `json:"data,omitempty"`
	}

	defer utils.CloseQuietly(do.Body)

	if err = json.NewDecoder(do.Body).Decode(&data); err != nil {
		return "", "", err
	}
	if len(data.Errors) > 0 {
		return "", "", errors.New(data.Errors[0].Detail)
	}
	if len(data.Data) == 0 {
		return "", "", ErrEntityNotFound("lyrics")
	}

	return *data.Data[0].Attributes.Ttml, *data.Data[0].Attributes.TtmlLocalizations, nil
}

func GetAllGenres() ([]Genres, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		"https://amp-api.music.apple.com/v1/catalog/"+config.Storefront+"/genres",
		nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", Authorization)
	req.Header.Set("User-Agent", config.UserAgent)
	req.Header.Set("Origin", config.Origin)
	req.Header.Set("Referer", config.Referer)

	do, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	data := new(struct {
		Data []Genres `json:"data,omitempty"`
	})

	defer utils.CloseQuietly(do.Body)

	if err = json.NewDecoder(do.Body).Decode(&data); err != nil {
		return nil, err
	}

	return data.Data, nil
}
