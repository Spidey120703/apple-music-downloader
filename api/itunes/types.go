package itunes

import (
	"encoding/json"
)

type LookupResponse struct {
	ResultCount int               `json:"resultCount"`
	Results     []json.RawMessage `json:"results"`
}

type IResult interface {
	GetWrapperType() string
}

type Result struct {
	WrapperType string `json:"wrapperType"`
}

func (s Result) GetWrapperType() string {
	return s.WrapperType
}

type Song struct {
	WrapperType            *string  `json:"wrapperType"`
	Kind                   *string  `json:"kind"`
	ArtistID               *int     `json:"artistId"`
	CollectionID           *int     `json:"collectionId"`
	TrackID                *int     `json:"trackId"`
	ArtistName             *string  `json:"artistName"`
	CollectionName         *string  `json:"collectionName"`
	TrackName              *string  `json:"trackName"`
	CollectionCensoredName *string  `json:"collectionCensoredName"`
	TrackCensoredName      *string  `json:"trackCensoredName"`
	ArtistViewURL          *string  `json:"artistViewUrl"`
	CollectionViewURL      *string  `json:"collectionViewUrl"`
	TrackViewURL           *string  `json:"trackViewUrl"`
	PreviewURL             *string  `json:"previewUrl"`
	ArtworkURL30           *string  `json:"artworkUrl30"`
	ArtworkURL60           *string  `json:"artworkUrl60"`
	ArtworkURL100          *string  `json:"artworkUrl100"`
	CollectionPrice        *float64 `json:"collectionPrice,omitempty"`
	TrackPrice             *float64 `json:"trackPrice,omitempty"`
	ReleaseDate            *string  `json:"releaseDate"`
	CollectionExplicitness *string  `json:"collectionExplicitness"`
	TrackExplicitness      *string  `json:"trackExplicitness"`
	DiscCount              *int     `json:"discCount"`
	DiscNumber             *int     `json:"discNumber"`
	TrackCount             *int     `json:"trackCount"`
	TrackNumber            *int     `json:"trackNumber"`
	TrackTimeMillis        *int     `json:"trackTimeMillis"`
	Country                *string  `json:"country"`
	Currency               *string  `json:"currency"`
	PrimaryGenreName       *string  `json:"primaryGenreName"`
	IsStreamable           *bool    `json:"isStreamable"`
}

func (s Song) GetWrapperType() string {
	return *s.WrapperType
}

type MusicVideo struct {
	WrapperType            *string  `json:"wrapperType"`
	Kind                   *string  `json:"kind"`
	ArtistID               *int     `json:"artistId"`
	CollectionID           *int     `json:"collectionId"`
	TrackID                *int     `json:"trackId"`
	ArtistName             *string  `json:"artistName"`
	CollectionName         *string  `json:"collectionName"`
	TrackName              *string  `json:"trackName"`
	CollectionCensoredName *string  `json:"collectionCensoredName"`
	TrackCensoredName      *string  `json:"trackCensoredName"`
	ArtistViewURL          *string  `json:"artistViewUrl"`
	CollectionViewURL      *string  `json:"collectionViewUrl"`
	TrackViewURL           *string  `json:"trackViewUrl"`
	PreviewURL             *string  `json:"previewUrl"`
	ArtworkURL30           *string  `json:"artworkUrl30"`
	ArtworkURL60           *string  `json:"artworkUrl60"`
	ArtworkURL100          *string  `json:"artworkUrl100"`
	CollectionPrice        *float64 `json:"collectionPrice,omitempty"`
	TrackPrice             *float64 `json:"trackPrice,omitempty"`
	ReleaseDate            *string  `json:"releaseDate"`
	CollectionExplicitness *string  `json:"collectionExplicitness"`
	TrackExplicitness      *string  `json:"trackExplicitness"`
	DiscCount              *int     `json:"discCount"`
	DiscNumber             *int     `json:"discNumber"`
	TrackCount             *int     `json:"trackCount"`
	TrackNumber            *int     `json:"trackNumber"`
	TrackTimeMillis        *int     `json:"trackTimeMillis"`
	Country                *string  `json:"country"`
	Currency               *string  `json:"currency"`
	PrimaryGenreName       *string  `json:"primaryGenreName"`
}

func (s MusicVideo) GetWrapperType() string {
	return *s.WrapperType
}

type Album struct {
	WrapperType            *string  `json:"wrapperType"`
	CollectionType         *string  `json:"collectionType"`
	ArtistID               *int     `json:"artistId"`
	CollectionID           *int     `json:"collectionId"`
	AmgArtistID            *int     `json:"amgArtistId"`
	ArtistName             *string  `json:"artistName"`
	CollectionName         *string  `json:"collectionName"`
	CollectionCensoredName *string  `json:"collectionCensoredName"`
	ArtistViewURL          *string  `json:"artistViewUrl"`
	CollectionViewURL      *string  `json:"collectionViewUrl"`
	ArtworkURL60           *string  `json:"artworkUrl60"`
	ArtworkURL100          *string  `json:"artworkUrl100"`
	CollectionPrice        *float64 `json:"collectionPrice,omitempty"`
	CollectionExplicitness *string  `json:"collectionExplicitness"`
	TrackCount             *int     `json:"trackCount"`
	Copyright              *string  `json:"copyright"`
	Country                *string  `json:"country"`
	Currency               *string  `json:"currency"`
	ReleaseDate            *string  `json:"releaseDate"`
	PrimaryGenreName       *string  `json:"primaryGenreName"`
}

func (s Album) GetWrapperType() string {
	return *s.WrapperType
}

type Composer struct {
	WrapperType      *string `json:"wrapperType"`
	ArtistType       *string `json:"artistType"`
	ArtistName       *string `json:"artistName"`
	ArtistLinkURL    *string `json:"artistLinkUrl"`
	ArtistID         *int    `json:"artistId"`
	AmgArtistID      *int    `json:"amgArtistId"`
	PrimaryGenreName *string `json:"primaryGenreName"`
	PrimaryGenreID   *int    `json:"primaryGenreId"`
}

func (s *Composer) GetWrapperType() string {
	return *s.WrapperType
}
