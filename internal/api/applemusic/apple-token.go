package applemusic

import (
	"downloader/internal/config"
	"downloader/pkg/utils"
	"errors"
	"io"
	"net/http"
	"net/url"
	"regexp"
)

var Authorization string

func SetAuthorization(token string) {
	Authorization = "Bearer " + token
}

func GetToken() (string, error) {
	const baseUrl = "https://beta.music.apple.com"
	index, _ := url.JoinPath(baseUrl, config.Storefront)
	resp, err := http.Get(index)
	if err != nil {
		return "", err
	}

	defer utils.CloseQuietly(resp.Body)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	regex := regexp.MustCompile(`src="(/assets/index-legacy[-~][0-9a-f]{8,10}\.js)"`)
	submatcheds := regex.FindAllSubmatch(body, 2)

	var token = ""

	regex = regexp.MustCompile(`="(eyJh[0-9A-Za-z\-_]+={0,2}\.[0-9A-Za-z\-_]+={0,2}\.[0-9A-Za-z\-_]+={0,2})"`)
	for _, submatched := range submatcheds {
		resp, err = http.Get(baseUrl + string(submatched[1]))
		if err != nil {
			return "", err
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}

		if temp := regex.FindStringSubmatch(string(body)); len(temp) != 0 {
			token = temp[1]
			break
		}

		if err = resp.Body.Close(); err != nil {
			return "", err
		}
	}

	if len(token) == 0 {
		return "", errors.New("unable to get token")
	}

	return token, nil
}

func RefreshToken() {
	token, err := GetToken()
	if err != nil {
		panic(err)
	}
	SetAuthorization(token)
}
