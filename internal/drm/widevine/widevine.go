package widevine

import (
	"downloader/internal/api/applemusic"
	"downloader/internal/drm/widevine/cdm"
	"encoding/base64"
	"errors"
	"fmt"

	"google.golang.org/protobuf/proto"
)

const (
	SystemStringPrefix = "com.widevine.alpha"
	KeyFormatString    = "urn:uuid:edef8ba9-79d6-4ace-a3c8-27dcd51d21ed"
)

var SystemID = []byte("0123456789abcdef0123456789abcdef")

// GeneratePSSH generates a Widevine PSSH (Protection System Specific Header) box.
// This is required to initialize the CDM when playing HLS content encrypted
// with Widevine, allowing the CDM to construct the necessary license challenge.
func GeneratePSSH(contentId, keyIdBase64 string) (string, error) {
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

	widevineCenc := append(SystemID, widevineCencHeaderProtobuf...)
	pssh := base64.StdEncoding.EncodeToString(widevineCenc)
	return pssh, nil
}

// GetKey performs a license exchange with an Apple-hosted Widevine license server.
// It initiates the CDM, generates a Widevine-formatted license challenge,
// and transmits it to the specified Apple HLS key server. Finally, it
// extracts and returns the content decryption key from the server's response.
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
			KeySystem:     SystemStringPrefix,
		})
	if err != nil {
		return nil, fmt.Errorf("failed to request license from server: %w", err)
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
