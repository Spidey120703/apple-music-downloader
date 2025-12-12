package downloader

import (
	"downloader/internal/config"
	"downloader/pkg/utils"
	"encoding/base64"
	"net/url"
	"regexp"
)

var ImageURLPattern = regexp.MustCompile(`^(?P<url_without_ext>https://is[1-5]-ssl\.mzstatic\.com/image/thumb/(?P<img_mode>[0-9A-Za-z]+)/(?:(?:v\d/(?:[0-9a-f]{2}/){3}(?P<uuid>[0-9a-f]{8}-(?:[0-9a-f]{4}-){3}[0-9a-f]{12})/|(?:[0-9a-f]{2}/){3})(?P<original_filename>[^/]+(?P<original_file_ext>\.[0-9a-z]+?))/)?(?P<new_filename>(?:(?P<width_placeholder>\{w})|\d+)x(?:(?P<height_placeholder>\{h})|\d+).+?))(?P<new_file_ext>\.[0-9a-z]+)(?P<query_string>\?[_=+&%\-0-9A-Za-z]+)?$`)

const (
	KeyUrlWithoutExt    = "url_without_ext"
	KeyImgMode          = "img_mode"
	KeyUuid             = "uuid"
	KeyOriginalFilename = "original_filename"
	KeyOriginalFileExt  = "original_file_ext"
	KeyNewFilename      = "new_filename"
	KeyWidthPlace       = "width_placeholder"
	KeyHeightPlace      = "height_placeholder"
	KeyNewFileExt       = "new_file_ext"
	KeyQueryString      = "query_string"
	KeyAvatarText       = "avatar_text"
)

const (
	FilenameFormatUUID             = "{uuid}{original_file_ext}"
	FilenameFormatOriginalFileName = "{original_filename}"
	FilenameFormatNewFileName      = "{new_filename}"
	FilenameFormatAvatarText       = "{avatar_text}"
)

func GetFormattedImageURLName(rawURL, format string, context map[string]string) (finalURL string, filename string) {
	useOriginalExt := config.UseOriginalExt

	finalURL = utils.Format(rawURL, context)
	submatches := utils.FindStringSubmatchMap(ImageURLPattern, finalURL)

	ctx := make(map[string]string)

	// Apple text-avatar generator
	// Example:
	// 	https://is1-ssl.mzstatic.com/image/thumb/gen/{w}x{h}AM.SCM01.jpg?name=XXXX&signature=xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx&vkey=1
	// The extension in the URL is not the real file format, and there is no
	// original extension to preserve. When mode == "gen", always ignore
	// useOriginalExt. The "name" query param (Base64) contains the avatar text.
	if submatches[KeyImgMode] == "gen" {
		useOriginalExt = false
		if queryString, found := submatches[KeyQueryString]; found && len(queryString) > 1 {
			query, err := url.ParseQuery(queryString[1:])
			if err == nil {
				if name, err := base64.StdEncoding.DecodeString(query.Get("name")); err == nil {
					ctx[KeyAvatarText] = string(name)
				}
			}
		}
	}

	if useOriginalExt {
		finalURL = submatches[KeyUrlWithoutExt] + submatches[KeyOriginalFileExt] + submatches[KeyQueryString]
	} else {
		format = format + submatches[KeyNewFileExt]
	}

	for k, v := range submatches {
		ctx[k] = v
	}
	for k, v := range context {
		ctx[k] = v
	}
	return finalURL, utils.Format(format, ctx)
}

func FixLanguageQuery(playlistURL string) string {
	parse, err := url.Parse(playlistURL)
	if err != nil {
		return playlistURL
	}

	lang := config.Language

	query := parse.Query()
	query.Set("l", lang)
	parse.RawQuery = query.Encode()

	return parse.String()
}
