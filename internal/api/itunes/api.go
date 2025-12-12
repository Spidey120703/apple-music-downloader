package itunes

import (
	"downloader/internal/api/applemusic"
	"downloader/internal/config"
	"encoding/json"
	"errors"
	"net/http"
)

func getITunesLookup(id string, entity string) (*LookupResponse, error) {
	req, err := http.NewRequest(
		http.MethodPost,
		"https://itunes.apple.com/lookup",
		nil)
	if err != nil {
		panic(err)
	}

	req.Header.Set("Authorization", applemusic.Authorization)
	req.Header.Set("User-Agent", config.UserAgent)
	req.Header.Set("Origin", config.Origin)
	req.Header.Set("Referer", config.Referer)

	query := req.URL.Query()
	query.Set("id", id)
	query.Set("entity", entity)
	query.Set("country", config.Storefront)
	query.Set("lang", "en_us")
	query.Set("limit", "100")
	req.URL.RawQuery = query.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	data := new(LookupResponse)

	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, err
	}

	if data.ResultCount == 0 {
		return data, errors.New("not found")
	}

	return data, nil
}

func GetITunesInfo[T IResult](id string, entity string) (*T, error) {
	response, err := getITunesLookup(id, entity)
	if err != nil {
		return nil, err
	}

	result := new(Result)
	err = json.Unmarshal(response.Results[0], &result)
	if err != nil {
		return nil, err
	}

	info := new(T)
	err = json.Unmarshal(response.Results[0], &info)
	if err != nil {
		return nil, err
	}

	return info, nil
}
