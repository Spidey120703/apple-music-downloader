package applemusic

import (
	"downloader/internal/api"
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
		"https://amp-api.music.apple.com/v1/catalog/"+config.Get().AppleMusic.Storefront+"/albums/"+id,
		nil)
	if err != nil {
		return nil, err
	}

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

	do, err := api.Client().Do(req)
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

func GetSongData(id string) (*Songs, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		"https://amp-api.music.apple.com/v1/catalog/"+config.Get().AppleMusic.Storefront+"/songs/"+id,
		nil)
	if err != nil {
		return nil, err
	}

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

	do, err := api.Client().Do(req)
	if err != nil {
		return nil, err
	}

	data := new(struct {
		Errors []Errors `json:"errors,omitempty"`
		Data   []Songs  `json:"data,omitempty"`
	})

	if err = json.NewDecoder(do.Body).Decode(&data); err != nil {
		return nil, err
	}
	if len(data.Errors) > 0 {
		return nil, errors.New(data.Errors[0].Detail)
	}
	if len(data.Data) != 1 {
		return nil, ErrEntityNotFound("songs")
	}

	return &data.Data[0], nil
}

func GetMusicVideoData(id string) (*MusicVideos, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		"https://amp-api.music.apple.com/v1/catalog/"+config.Get().AppleMusic.Storefront+"/music-videos/"+id,
		nil)
	if err != nil {
		return nil, err
	}

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

	do, err := api.Client().Do(req)
	if err != nil {
		return nil, err
	}

	data := new(struct {
		Errors []Errors      `json:"errors,omitempty"`
		Data   []MusicVideos `json:"data,omitempty"`
	})

	if err = json.NewDecoder(do.Body).Decode(&data); err != nil {
		return nil, err
	}
	if len(data.Errors) > 0 {
		return nil, errors.New(data.Errors[0].Detail)
	}
	if len(data.Data) != 1 {
		return nil, ErrEntityNotFound("music-videos")
	}

	return &data.Data[0], nil
}

func GetLyrics(id string) (string, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		"https://amp-api.music.apple.com/v1/catalog/"+config.Get().AppleMusic.Storefront+"/songs/"+id+"/lyrics",
		nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Media-User-Token", config.Get().AppleMusic.MediaUserToken)

	do, err := api.Client().Do(req)
	if err != nil {
		return "", err
	}

	var data struct {
		Errors []Errors `json:"errors,omitempty"`
		Data   []Lyrics `json:"data,omitempty"`
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
	req.Header.Set("Media-User-Token", config.Get().AppleMusic.MediaUserToken)

	query := req.URL.Query()
	query.Set("l[lyrics]", "zh-hans-cn")
	query.Set("l[script]", "zh-Hans")
	query.Set("extend", "ttml,ttmlLocalizations")
	req.URL.RawQuery = query.Encode()

	do, err := api.Client().Do(req)
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
		"https://amp-api.music.apple.com/v1/catalog/"+config.Get().AppleMusic.Storefront+"/genres",
		nil)
	if err != nil {
		return nil, err
	}

	do, err := api.Client().Do(req)
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
