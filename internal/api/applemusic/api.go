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
	"regexp"
)

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

func RefreshToken() {
	token, err := GetToken()
	if err != nil {
		panic(err)
	}
	SetAuthorization(token)
}

func GetAlbumData(id string) (*Albums, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		"https://amp-api.music.apple.com/v1/catalog/"+config.StoreFront+"/albums/"+id,
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

func GetLyrics(id string) (string, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		"https://amp-api.music.apple.com/v1/catalog/cn/songs/"+id+"/lyrics",
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
		Data []Lyrics `json:"Data,omitempty"`
	}

	defer utils.CloseQuietly(do.Body)

	err = json.NewDecoder(do.Body).Decode(&data)
	if err != nil {
		return "", err
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
		Data []Lyrics `json:"Data,omitempty"`
	}

	defer utils.CloseQuietly(do.Body)

	err = json.NewDecoder(do.Body).Decode(&data)
	if err != nil {
		return "", "", err
	}

	return *data.Data[0].Attributes.Ttml, *data.Data[0].Attributes.TtmlLocalizations, nil
}

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
