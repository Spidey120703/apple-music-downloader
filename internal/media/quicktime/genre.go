package quicktime

import (
	"downloader/internal/api/applemusic"
	"downloader/pkg/LOG"
	"strconv"
)

type Genre struct {
	ID       int
	Name     string
	Children []Genre
}

// https://api.music.apple.com/v1/catalog/us/genres

var GenresRoot = Genre{
	ID:   34,
	Name: "Music",
	Children: []Genre{
		{ID: 20, Name: "Alternative"},
		{ID: 2, Name: "Blues"},
		{ID: 22, Name: "Christian"},
		{ID: 5, Name: "Classical"},
		{ID: 6, Name: "Country"},
		{ID: 17, Name: "Dance"},
		{ID: 7, Name: "Electronic"},
		{ID: 18, Name: "Hip-Hop/Rap"},
		{ID: 8, Name: "Holiday"},
		{ID: 11, Name: "Jazz"},
		{ID: 4, Name: "Children's Music"},
		{ID: 12, Name: "Latin"},
		{ID: 15, Name: "R&B/Soul"},
		{ID: 24, Name: "Reggae"},
		{ID: 10, Name: "Singer/Songwriter"},
		{ID: 16, Name: "Soundtrack"},
		{ID: 19, Name: "Worldwide"},
		{ID: 14, Name: "Pop", Children: []Genre{
			{ID: 51, Name: "K-Pop"},
		}},
		{ID: 21, Name: "Rock", Children: []Genre{
			{ID: 1153, Name: "Metal"},
		}},
	},
}

var Genres = []Genre{
	{ID: 34, Name: "Music"},
	{ID: 20, Name: "Alternative"},
	{ID: 2, Name: "Blues"},
	{ID: 22, Name: "Christian"},
	{ID: 5, Name: "Classical"},
	{ID: 6, Name: "Country"},
	{ID: 17, Name: "Dance"},
	{ID: 7, Name: "Electronic"},
	{ID: 18, Name: "Hip-Hop/Rap"},
	{ID: 8, Name: "Holiday"},
	{ID: 11, Name: "Jazz"},
	{ID: 51, Name: "K-Pop"},
	{ID: 4, Name: "Children's Music"},
	{ID: 12, Name: "Latin"},
	{ID: 1153, Name: "Metal"},
	{ID: 14, Name: "Pop"},
	{ID: 15, Name: "R&B/Soul"},
	{ID: 24, Name: "Reggae"},
	{ID: 21, Name: "Rock"},
	{ID: 10, Name: "Singer/Songwriter"},
	{ID: 16, Name: "Soundtrack"},
	{ID: 19, Name: "Worldwide"},
}

func LoadCurrentStorefrontGenres() {
	genres, err := applemusic.GetAllGenres()
	if err != nil {
		LOG.Error.Printf("failed to get all genres: %v", err)
		return
	}
	for _, g := range genres {
		id, _ := strconv.Atoi(g.ID)
		Genres = append(Genres, Genre{ID: id, Name: g.Attributes.Name})
	}
}

func GetGenreID(genreNames []string) int {
	if len(genreNames) == 0 {
		return 0
	}
	for _, genre := range Genres {
		if genre.Name == genreNames[0] {
			return genre.ID
		}
	}
	return 0
}
