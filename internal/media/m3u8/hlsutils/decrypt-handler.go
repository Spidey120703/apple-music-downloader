package hlsutils

import (
	"downloader/internal/drm/fairplay"
	"downloader/internal/drm/widevine"
	"downloader/internal/media/mp4/cmaf"
	"downloader/pkg/LOG"
	"downloader/pkg/ansi"
	"downloader/pkg/utils"
	"errors"
	"os"
	"strings"
)

type DecryptHandler struct {
	*Context
}

func (ctx *DecryptHandler) getKeys(entry *MediaPlaylistEntry) (keys [][]byte, drm DrmType, err error) {
	var key []byte
	switch ctx.Type {
	case MediaTypeSong:
		if _, found := entry.KeyURIs[fairplay.KeyFormatString]; found {
			drm = DrmFairPlay
			for _, keyURI := range entry.KeyURIs[fairplay.KeyFormatString] {
				keys = append(keys, []byte(keyURI))
			}
			return
		} else if _, found = entry.KeyURIs[widevine.KeyFormatString]; found {
			// fallthrough
		} else {
			return nil, DrmNone, errors.New("key URI not found")
		}
		fallthrough
	case MediaTypeMusicVideo:
		if _, found := entry.KeyURIs[widevine.KeyFormatString]; !found {
			return nil, DrmNone, errors.New("key URI not found")
		}
		var pssh string
		for _, keyURI := range entry.KeyURIs[widevine.KeyFormatString] {
			var keyBase64 string
			if strings.HasPrefix(keyURI, "data:text/plain;base64,") {
				keyBase64 = keyURI[len("data:text/plain;base64,"):]
			} else if strings.HasPrefix(keyURI, "data:;base64,") {
				keyBase64 = keyURI[len("data:;base64,"):]
			}
			if pssh, err = widevine.GeneratePSSH("", keyBase64); err != nil {
				return
			}
			if key, err = widevine.GetKey(pssh, keyURI, ctx.WebPlayback); err != nil {
				return
			}
			keys = append(keys, key)
		}
		drm = DrmWidevine
	default:
	}
	return
}

func (ctx *DecryptHandler) decryptEntry(entry *MediaPlaylistEntry) (err error) {
	var inputs []*os.File
	if inputs, err = utils.OpenFiles(entry.FilePaths); err != nil {
		return
	}
	defer utils.CloseQuietlyAll(inputs)

	if len(inputs) == 0 {
		return errors.New("empty media playlist")
	}

	var drm DrmType
	var keys [][]byte

	if keys, drm, err = ctx.getKeys(entry); err != nil {
		return
	}

	switch drm {
	case DrmFairPlay:
		LOG.Info.Printf("Decrypting using Apple FairPlay CDM...")
		entry.Decryptor = fairplay.New(ctx.AdamID)
	case DrmWidevine:
		LOG.Info.Printf("Decrypting using Google Widevine CDM...")
		LOG.Info.Printf(ansi.CSIFgBrightBlack + "- Dynamic encryption detected: Progress bar disabled, speed depends on sample count." + ansi.CSIReset)
		entry.Decryptor = widevine.New()
	default:
	}

	if err = entry.Decryptor.Initialize(inputs[0]); err != nil {
		return
	}
	if err = entry.Decryptor.DecryptHeader(keys); err != nil {
		return
	}

	var seg *cmaf.Segment
	for _, input := range inputs[1:] {
		if seg, err = entry.Decryptor.AddSegment(input); err != nil {
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
