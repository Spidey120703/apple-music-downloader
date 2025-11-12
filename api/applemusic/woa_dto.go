package applemusic

import "time"

type WebPlaybackResponse struct {
	SongList []WebPlaybackSong `json:"songList"`
	Status   int               `json:"status"`
}

type WebPlaybackSong struct {
	HlsKeyCertURL  string `json:"hls-key-cert-url"`
	HlsPlaylistURL string `json:"hls-playlist-url"`
	ArtworkUrls    struct {
		Default struct {
			URL string `json:"url"`
		} `json:"default"`
		Default2X struct {
			URL string `json:"url"`
		} `json:"default@2x"`
		ImageType string `json:"image-type"`
	} `json:"artwork-urls"`
	Assets []struct {
		Flavor      string `json:"flavor"`
		URL         string `json:"URL"`
		DownloadKey string `json:"downloadKey"`
		ArtworkURL  string `json:"artworkURL"`
		FileSize    int    `json:"file-size"`
		Md5         string `json:"md5"`
		Chunks      struct {
			ChunkSize int      `json:"chunkSize"`
			Hashes    []string `json:"hashes"`
		} `json:"chunks"`
		Metadata struct {
			ComposerID          string    `json:"composerId"`
			GenreID             int       `json:"genreId"`
			Copyright           string    `json:"copyright"`
			Year                int       `json:"year"`
			SortArtist          string    `json:"sort-artist"`
			IsMasteredForItunes bool      `json:"isMasteredForItunes"`
			VendorID            int       `json:"vendorId"`
			ArtistID            string    `json:"artistId"`
			Duration            int       `json:"duration"`
			DiscNumber          int       `json:"discNumber"`
			ItemName            string    `json:"itemName"`
			TrackCount          int       `json:"trackCount"`
			Xid                 string    `json:"xid"`
			BitRate             int       `json:"bitRate"`
			FileExtension       string    `json:"fileExtension"`
			SortAlbum           string    `json:"sort-album"`
			Genre               string    `json:"genre"`
			Rank                int       `json:"rank"`
			SortName            string    `json:"sort-name"`
			PlaylistID          string    `json:"playlistId"`
			SortComposer        string    `json:"sort-composer"`
			TrackNumber         int       `json:"trackNumber"`
			ReleaseDate         time.Time `json:"releaseDate"`
			Kind                string    `json:"kind"`
			Work                string    `json:"work"`
			PlaylistArtistName  string    `json:"playlistArtistName"`
			Gapless             bool      `json:"gapless"`
			ComposerName        string    `json:"composerName"`
			DiscCount           int       `json:"discCount"`
			SampleRate          int       `json:"sampleRate"`
			PlaylistName        string    `json:"playlistName"`
			Explicit            int       `json:"explicit"`
			ItemID              string    `json:"itemId"`
			S                   int       `json:"s"`
			Compilation         bool      `json:"compilation"`
			ArtistName          string    `json:"artistName"`
		} `json:"metadata"`
	} `json:"assets"`
	WidevineCertURL string `json:"widevine-cert-url"`
	FormerIds       []int  `json:"formerIds"`
	SongID          string `json:"songId"`
	IsItunesStream  bool   `json:"is-itunes-stream"`
	HlsKeyServerURL string `json:"hls-key-server-url"`
}

type WebPlaybackLicenseResponse struct {
	License    string `json:"license"`
	ErrorCode  int    `json:"errorCode"`
	RenewAfter int    `json:"renew-after"`
	Status     int    `json:"status"`
}
