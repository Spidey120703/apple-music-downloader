package applemusic

import (
	"bytes"
	"downloader/consts"
	"downloader/utils"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
)

const StoreFront = "cn"
const MediaUserToken = "Ai00hPjqDQcpdKvILsHxXWeJLNt2miOjjBe7cgSI0uIpZu0U90Fu7DQsovYaMHU+p+gJyOHUKfgA2vbGN19XbGy40oWwO3u+46cEucIzORDAuTaPQsrBvMZidhP2krg5QhPW3jYXuFgK2xUaFWrZ45jrun0MX4KeD3G/Lck8cwACZ+5BHeh4V65fpcTjLa6Sm8Uy7Na+R6bse+iBiuvgnVkirt1FmQdVK22RfyXAX7uJYpaAgw=="

func GetToken() (string, error) {
	const baseUrl = "https://beta.music.apple.com"
	resp, err := http.Get(baseUrl + "/cn")
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

		temp := regex.FindStringSubmatch(string(body))
		if len(temp) != 0 {
			token = temp[1]
			break
		}

		err = resp.Body.Close()
		if err != nil {
			return "", err
		}
	}

	if len(token) == 0 {
		return "", errors.New("unable to get token")
	}

	return token, nil
}

func GetAlbumData(id string, token string) (*Albums, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		"https://amp-api.music.apple.com/v1/catalog/"+StoreFront+"/albums/"+id,
		nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("User-Agent", consts.UserAgent)
	req.Header.Set("Origin", consts.Origin)
	req.Header.Set("Referer", consts.Referer)

	query := req.URL.Query()
	query.Set("extend", "artistBio,bornOrFormed,editorialArtwork,editorialNotes,editorialVideo,extendedAssetUrls,hero,isGroup,offers,origin,plainEditorialNotes,seoDescription,seoTitle,artistUrl,contentRating")
	query.Set("include", "albums,record-labels,artists,persons,bands,composers,credits,lyrics,music-videos,tracks,genres")
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
		Data []Albums `json:"data,omitempty"`
	})

	err = json.NewDecoder(do.Body).Decode(&data)
	if err != nil {
		return nil, err
	}

	//_ = json.NewEncoder(os.Stdout).Encode(data)

	if len(data.Data) != 1 {
		return nil, errors.New("not found")
	}

	return &data.Data[0], nil
}

func GetLyrics(id string, token string) (string, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		"https://amp-api.music.apple.com/v1/catalog/cn/songs/"+id+"/lyrics",
		nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("User-Agent", consts.UserAgent)
	req.Header.Set("Origin", consts.Origin)
	req.Header.Set("Referer", consts.Referer)
	req.Header.Set("Media-User-Token", MediaUserToken)

	do, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	var data struct {
		Data []Lyrics `json:"Data,omitempty"`
	}

	defer utils.CloseQuietly(do.Body)

	err = json.NewDecoder(do.Body).Decode(&data)
	if err != nil {
		return "", err
	}

	return *data.Data[0].Attributes.Ttml, nil
}

func GetSyllableLyrics(id string, token string) (string, string, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		"https://amp-api.music.apple.com/v1/catalog/cn/songs/"+id+"/syllable-lyrics",
		nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("User-Agent", consts.UserAgent)
	req.Header.Set("Origin", consts.Origin)
	req.Header.Set("Referer", consts.Referer)
	req.Header.Set("Media-User-Token", MediaUserToken)

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
		Data []struct {
			ID         string `json:"id,omitempty"`
			Type       string `json:"type,omitempty"`
			Attributes struct {
				PlayParams struct {
					CatalogID   string `json:"catalogId,omitempty"`
					DisplayType int    `json:"displayType,omitempty"`
					ID          string `json:"id,omitempty"`
					Kind        string `json:"kind,omitempty"`
				} `json:"playParams,omitempty"`
				Ttml              string `json:"ttml,omitempty"`
				TtmlLocalizations string `json:"ttmlLocalizations,omitempty"`
			} `json:"attributes,omitempty"`
		} `json:"Data,omitempty"`
	}

	defer utils.CloseQuietly(do.Body)

	err = json.NewDecoder(do.Body).Decode(&data)
	if err != nil {
		return "", "", err
	}

	return data.Data[0].Attributes.Ttml, data.Data[0].Attributes.TtmlLocalizations, nil
}

func GetWebPlayback(id string, token string) (*WebPlaybackSong, error) {
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
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("User-Agent", consts.UserAgent)
	req.Header.Set("Origin", consts.Origin)
	req.Header.Set("Referer", consts.Referer)
	req.Header.Set("X-Apple-Music-User-Token", MediaUserToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	data := new(WebPlaybackResponse)
	err = json.NewDecoder(resp.Body).Decode(&data)

	if len(data.SongList) < 1 {
		return nil, errors.New("song not found")
	}

	return &data.SongList[0], nil
}

func PostWebPlaybackLicense(url string, licenseRequest WebPlaybackLicenseRequest, token string) ([]byte, error) {
	body, err := json.Marshal(licenseRequest)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Apple-Music-User-Token", MediaUserToken)

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
