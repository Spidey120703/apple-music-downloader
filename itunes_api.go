package main

import (
	"downloader/api/applemusic"
	"downloader/api/itunes"
	"downloader/consts"
	"encoding/json"
	"errors"
	"net/http"
)

func getITunesLookup(id string, entity string, token string) (*itunes.LookupResponse, error) {
	req, err := http.NewRequest(
		http.MethodPost,
		"https://itunes.apple.com/lookup",
		nil)
	if err != nil {
		panic(err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("User-Agent", consts.UserAgent)
	req.Header.Set("Origin", consts.Origin)
	req.Header.Set("Referer", consts.Referer)

	query := req.URL.Query()
	query.Set("id", id)
	query.Set("entity", entity)
	query.Set("country", applemusic.StoreFront)
	query.Set("lang", "en_us")
	query.Set("limit", "100")
	req.URL.RawQuery = query.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	data := new(itunes.LookupResponse)

	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, err
	}

	if data.ResultCount == 0 {
		return data, errors.New("not found")
	}

	return data, nil
}

func GetITunesInfo[T itunes.IResult](id string, entity string, token string) (*T, error) {
	response, err := getITunesLookup(id, entity, token)
	if err != nil {
		return nil, err
	}

	result := new(itunes.Result)
	err = json.Unmarshal(response.Results[0], &result)
	if err != nil {
		return nil, err
	}

	/*
		if result.GetWrapperType() != "track" {
			return nil, errors.New(fmt.Sprintf("invalid wrapper type: %s", result.GetWrapperType()))
		}
	*/

	info := new(T)
	err = json.Unmarshal(response.Results[0], &info)
	if err != nil {
		return nil, err
	}

	return info, nil
}
