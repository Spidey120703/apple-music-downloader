package applemusic

import (
	"bytes"
	"downloader/internal/config"
	"downloader/pkg/utils"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

func GetWebPlayback(id string) (*WebPlaybackSong, error) {
	reqBody := WebPlaybackRequest{SalableAdamId: id}
	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		"https://play.music.apple.com/WebObjects/MZPlay.woa/wa/webPlayback",
		bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", Authorization)
	req.Header.Set("User-Agent", config.UserAgent)
	req.Header.Set("Origin", config.Origin)
	req.Header.Set("Referer", config.Referer)
	req.Header.Set("X-Apple-Music-User-Token", config.MediaUserToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer utils.CloseQuietly(resp.Body)

	data := new(WebPlaybackResponse)
	err = json.NewDecoder(resp.Body).Decode(&data)

	if len(data.SongList) < 1 {
		if data.ErrorMessage != nil {
			return nil, errors.New(*data.ErrorMessage)
		}
		if data.CustomerMessage != nil {
			return nil, errors.New(*data.CustomerMessage)
		}
		return nil, errors.New("song not found")
	}

	return &data.SongList[0], nil
}

func PostWebPlaybackLicense(url string, licenseRequest WebPlaybackLicenseRequest) ([]byte, error) {
	body, err := json.Marshal(licenseRequest)
	if err != nil {
		return nil, err
	}

	if len(url) == 0 {
		url = "https://play.itunes.apple.com/WebObjects/MZPlay.woa/wa/acquireWebPlaybackLicense"
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", Authorization)
	req.Header.Set("User-Agent", config.UserAgent)
	req.Header.Set("Origin", config.Origin)
	req.Header.Set("Referer", config.Referer)
	req.Header.Set("X-Apple-Music-User-Token", config.MediaUserToken)

	do, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	response, err := io.ReadAll(do.Body)
	if err != nil {
		return nil, err
	}

	var licenseResponse WebPlaybackLicenseResponse
	err = json.Unmarshal(response, &licenseResponse)
	if err != nil {
		return nil, err
	}

	if licenseResponse.Status != 0 || licenseResponse.ErrorCode != 0 {
		return nil, errors.New("something wrong during WebPlayback licensing")
	}

	return base64.StdEncoding.DecodeString(licenseResponse.License)
}
