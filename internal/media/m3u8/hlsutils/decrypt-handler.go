package hlsutils

import (
	"downloader/internal/drm/fairplay"
	"downloader/internal/drm/widevine"
	"downloader/internal/media/mp4/cmaf"
	"downloader/pkg/LOG"
	"downloader/pkg/utils"
	"errors"
	"os"
	"path"
	"strings"
)

type DecryptHandler struct {
	*Context
}

func loadCaches(tempDir string, entry *MediaPlaylistEntry) (files []*os.File, err error) {
	var fileList []string
	for _, uri := range entry.URIs {
		filePath := path.Join(tempDir, uri[strings.LastIndex(uri, "/")+1:])
		fileList = append(fileList, filePath)
	}
	return utils.OpenFiles(fileList)
}

func (ctx *DecryptHandler) getKeys(entry *MediaPlaylistEntry) (keys [][]byte, err error) {
	var key []byte
	switch ctx.Type {
	case MediaTypeSong:
		if _, found := entry.KeyURIs[fairplay.KeyFormatString]; !found {
			return nil, errors.New("key URI not found")
		}
		for _, keyURI := range entry.KeyURIs[fairplay.KeyFormatString] {
			keys = append(keys, []byte(keyURI))
		}
	case MediaTypeMusicVideo:
		if _, found := entry.KeyURIs[widevine.KeyFormatString]; !found {
			return nil, errors.New("key URI not found")
		}
		var pssh string
		for _, keyURI := range entry.KeyURIs[widevine.KeyFormatString] {
			if pssh, err = widevine.GetPSSH("", keyURI[len("data:text/plain;base64,"):]); err != nil {
				return
			}
			if key, err = widevine.GetKey(pssh, keyURI, ctx.WebPlayback); err != nil {
				return
			}
			keys = append(keys, key)
		}
	default:
	}
	return
}

func (ctx *DecryptHandler) decryptEntry(entry *MediaPlaylistEntry) (err error) {
	switch ctx.Type {
	case MediaTypeSong:
		LOG.Info.Printf("Decrypting using Apple FairPlay CDM...")
		entry.Decryptor = fairplay.New(ctx.AdamID)
	case MediaTypeMusicVideo:
		LOG.Info.Printf("Decrypting using Google Widevine CDM... (dynamic encryption detected: progress bar disabled, speed depends on sample count)")
		entry.Decryptor = widevine.New()
	default:
	}

	var inputs []*os.File
	if inputs, err = loadCaches(ctx.TempDir, entry); err != nil {
		return
	}
	defer utils.CloseQuietlyAll(inputs)

	if len(inputs) == 0 {
		return errors.New("empty media playlist")
	}

	var keys [][]byte
	if keys, err = ctx.getKeys(entry); err != nil {
		return
	}

	if err = entry.Decryptor.Initialize(inputs[0]); err != nil {
		return
	}
	if err = entry.Decryptor.DecryptHeader(keys); err != nil {
		return
	}

	var seg *cmaf.Segment
	for _, input := range inputs[1:] {
		if seg, err = entry.Decryptor.MergeSegment(input); err != nil {
			return
		}
		if err = entry.Decryptor.DecryptSegment(seg, keys); err != nil {
			return
		}
	}

	LOG.Info.Println("Decryption completed.")
	return
}

func (ctx *DecryptHandler) decryptSegments() (err error) {
	for idx, entry := range ctx.MediaPlaylistEntries {
		LOG.Info.Printf("Starting decryption for track %d", idx+1)
		if err = ctx.decryptEntry(entry); err != nil {
			return
		}
		LOG.Info.Println()
	}
	return err
}

func (ctx *DecryptHandler) Execute() (err error) {
	if !ctx.IsEncrypted {
		return
	}
	return ctx.decryptSegments()
}
