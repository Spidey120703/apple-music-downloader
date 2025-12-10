package widevine

import (
	"downloader/api/applemusic"
	"downloader/drm/widevine/cdm"
	"encoding/base64"
	"errors"

	"google.golang.org/protobuf/proto"
)

const (
	SystemStringPrefix = "com.widevine.alpha"
	KeyFormatString    = "urn:uuid:edef8ba9-79d6-4ace-a3c8-27dcd51d21ed"
)

// GetPSSH - Get PSSH (Protection System Specific Header) of HLS
func GetPSSH(contentId, keyIdBase64 string) (string, error) {
	keyId, err := base64.StdEncoding.DecodeString(keyIdBase64)
	if err != nil {
		return "", err
	}
	algorithm := cdm.WidevineCencHeader_AESCTR
	widevineCencHeader := &cdm.WidevineCencHeader{
		Algorithm: &algorithm,
		KeyId:     [][]byte{keyId},
		Provider:  new(string),
		ContentId: []byte(contentId),
		Policy:    new(string),
	}
	widevineCencHeaderProtobuf, err := proto.Marshal(widevineCencHeader)

	widevineCenc := append([]byte("0123456789abcdef0123456789abcdef"), widevineCencHeaderProtobuf...)
	pssh := base64.StdEncoding.EncodeToString(widevineCenc)
	return pssh, nil
}

func GetKey(pssh string, keyURI string, song *applemusic.WebPlaybackSong) ([]byte, error) {
	initData, err := base64.StdEncoding.DecodeString(pssh)
	if err != nil {
		return nil, err
	}

	module, err := cdm.New(initData)
	if err != nil {
		return nil, err
	}

	challenge, err := module.Challenge()
	if err != nil {
		return nil, err
	}

	license, err := applemusic.PostWebPlaybackLicense(
		song.HlsKeyServerURL,
		applemusic.WebPlaybackLicenseRequest{
			AdamId:        song.SongID,
			IsLibrary:     true,
			UserInitiated: true,
			Challenge:     base64.StdEncoding.EncodeToString(challenge),
			Uri:           keyURI,
			KeySystem:     "com.widevine.alpha",
		})
	if err != nil {
		return nil, err
	}
	keys, err := module.GetLicenseKeys(challenge, license)
	if err != nil {
		return nil, err
	}

	for _, key := range keys {
		if key.Type == cdm.License_KeyContainer_CONTENT {
			return key.Key, nil
		}
	}

	return nil, errors.New("content key not found")
}
