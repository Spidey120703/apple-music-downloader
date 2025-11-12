package widevine

import (
	"downloader/api/applemusic"
	"downloader/drm/widevine/cdm"
	"downloader/log"
	"downloader/utils"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"google.golang.org/protobuf/proto"
)

const KeyFormatWidevine = "urn:uuid:edef8ba9-79d6-4ace-a3c8-27dcd51d21ed"

// GetPSSH - Get PSSH (Protection System Specific Header) of HLS
func GetPSSH(keyIdBase64 string) (string, error) {
	keyId, err := base64.StdEncoding.DecodeString(keyIdBase64)
	if err != nil {
		return "", err
	}
	algorithm := cdm.WidevineCencHeader_AESCTR
	widevineCencHeader := &cdm.WidevineCencHeader{
		Algorithm: &algorithm,
		KeyId:     [][]byte{keyId},
		Provider:  new(string),
		ContentId: []byte(""),
		Policy:    new(string),
	}
	widevineCencHeaderProtobuf, err := proto.Marshal(widevineCencHeader)

	widevineCenc := append([]byte("0123456789abcdef0123456789abcdef"), widevineCencHeaderProtobuf...)
	pssh := base64.StdEncoding.EncodeToString(widevineCenc)
	return pssh, nil
}

func GetKey(pssh string, keyUri string, song *applemusic.WebPlaybackSong, token string) (string, error) {
	initData, err := base64.StdEncoding.DecodeString(pssh)
	if err != nil {
		return "", err
	}

	module, err := cdm.New(initData)
	if err != nil {
		return "", err
	}

	challenge, err := module.Challenge()
	if err != nil {
		return "", err
	}

	license, err := applemusic.PostWebPlaybackLicense(
		song.HlsKeyServerURL,
		applemusic.WebPlaybackLicenseRequest{
			AdamId:        song.SongID,
			IsLibrary:     true,
			UserInitiated: true,
			Challenge:     base64.StdEncoding.EncodeToString(challenge),
			Uri:           keyUri,
			KeySystem:     "com.widevine.alpha",
		},
		token)
	if err != nil {
		return "", err
	}
	keys, err := module.GetLicenseKeys(challenge, license)
	if err != nil {
		return "", err
	}

	for _, key := range keys {
		if key.Type == cdm.License_KeyContainer_CONTENT {
			return hex.EncodeToString(key.Key), nil
		}
	}

	return "", errors.New("content key not found")
}

func Decrypt(files []*os.File, keyUri string, webPlayback *applemusic.WebPlaybackSong, token string) ([]byte, error) {
	encPath := strings.Replace(files[0].Name(), "-0.mp4", ".mp4", 1)
	decPath := strings.Replace(files[0].Name(), "-0.mp4", "-dec.mp4", 1)
	enc, err := os.Create(encPath)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		_, err := file.Seek(0, io.SeekStart)
		if err != nil {
			return nil, err
		}
		_, err = file.WriteTo(enc)
		if err != nil {
			return nil, err
		}
	}
	utils.CloseQuietly(enc)

	pssh, err := GetPSSH(keyUri[len("data:text/plain;base64,"):])
	if err != nil {
		return nil, err
	}

	key, err := GetKey(pssh, keyUri, webPlayback, token)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command("mp4decrypt", "--key", "1:"+key, encPath, decPath)
	cmd.Dir = filepath.Dir(".")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error.Printf("Output: %s", output)
		return nil, err
	}

	dec, err := os.Open(decPath)
	if err != nil {
		return nil, err
	}
	defer utils.CloseQuietly(dec)

	return io.ReadAll(dec)
}
