package metadata

import (
	"downloader/internal/api/applemusic"
	"downloader/internal/api/itunes"
	"downloader/internal/config"
	"downloader/internal/media/quicktime"
	"encoding/binary"
	"strconv"
	"time"
)

func assign[T any](ptr ...*T) *T {
	for _, p := range ptr {
		if p != nil {
			return p
		}
	}
	return nil
}

func ref[T any](val T) *T {
	var v = val
	return &v
}

func atoi(str string) *uint32 {
	integer, err := strconv.ParseInt(str, 10, 32)
	if err != nil {
		return nil
	}
	return ref(uint32(integer))
}

func genre(id int) (raw []byte) {
	raw = make([]byte, 2)
	binary.BigEndian.PutUint16(raw, uint16(id))
	return
}

type MusicVideoType int

const (
	MusicVideoTypeAlbumRelated MusicVideoType = iota
	MusicVideoTypeSongsRelated
)

type Context struct {
	Type                  MusicVideoType
	WebPlayback           *applemusic.WebPlaybackSong
	AppleMusicSongs       *applemusic.Songs
	AppleMusicMusicVideos *applemusic.MusicVideos
	AppleMusicAlbum       *applemusic.Albums
	ItunesSong            *itunes.Song
	ItunesMusicVideo      *itunes.MusicVideo
	CoverData             []byte
	LyricsData            string
}

func LoadSongMetadata(ctx Context) (meta *Metadata) {
	meta = &Metadata{}

	if len(ctx.WebPlayback.Assets) > 0 {
		assetMetadata := ctx.WebPlayback.Assets[0].Metadata

		meta.Title = &assetMetadata.ItemName
		meta.ArtistName = &assetMetadata.ArtistName
		meta.PlaylistArtist = &assetMetadata.PlaylistArtistName
		meta.ComposerName = &assetMetadata.ComposerName
		meta.AlbumName = &assetMetadata.PlaylistName
		meta.Work = assetMetadata.Work
		meta.Genre = genre(assetMetadata.GenreID)
		meta.Track = &Track{
			TrackNumber: uint32(assetMetadata.TrackNumber),
			TrackCount:  uint16(assetMetadata.TrackCount),
		}
		meta.DiskNumber = &Disk{
			DiskNumber: uint32(assetMetadata.DiscNumber),
			DiskCount:  uint16(assetMetadata.DiscCount),
		}
		meta.Compilation = ref(func() uint8 {
			if assetMetadata.Compilation {
				return 1
			} else {
				return 0
			}
		}())
		meta.PlayGap = ref(uint8(0))
		meta.ReleaseDate = ref(assetMetadata.ReleaseDate.Format(time.RFC3339))
		meta.AppleID = nil
		meta.Owner = nil
		meta.Copyright = &assetMetadata.Copyright
		meta.ItemID = atoi(assetMetadata.ItemID)
		meta.ArtistID = atoi(assetMetadata.ArtistID)
		meta.Rating = ref(uint8(0))
		meta.ComposerID = atoi(assetMetadata.ComposerID)
		meta.PlaylistID = atoi(assetMetadata.PlaylistID)
		meta.GenreID = ref(uint32(assetMetadata.GenreID))
		meta.StorefrontID = ref(uint32(quicktime.GetStorefrontID(config.Storefront)))
		meta.MediaType = ref(uint8(quicktime.MediaTypeNormal_Music))
		meta.PurchaseDate = nil
		meta.SortName = &assetMetadata.SortName
		meta.SortAlbum = &assetMetadata.SortAlbum
		meta.SortArtist = &assetMetadata.SortArtist
		meta.SortComposer = &assetMetadata.SortComposer
		meta.XID = &assetMetadata.Xid
		meta.Cover = ctx.CoverData
		if *ctx.AppleMusicSongs.Attributes.HasLyrics && len(ctx.LyricsData) != 0 {
			meta.Lyrics = &ctx.LyricsData
		}
	} else {
		meta.Title = assign(ctx.AppleMusicSongs.Attributes.Name, ctx.ItunesSong.TrackName)
		meta.ArtistName = assign(ctx.AppleMusicSongs.Attributes.ArtistName, ctx.ItunesSong.ArtistName)
		meta.PlaylistArtist = assign(ctx.AppleMusicAlbum.Attributes.ArtistName)
		meta.ComposerName = assign(ctx.AppleMusicSongs.Attributes.ComposerName)
		meta.AlbumName = assign(ctx.AppleMusicSongs.Attributes.AlbumName, ctx.AppleMusicAlbum.Attributes.Name, ctx.ItunesSong.CollectionName)
		meta.Work = assign(ctx.AppleMusicSongs.Attributes.WorkName)
		meta.Genre = genre(quicktime.GetGenreID(ctx.AppleMusicSongs.Attributes.GenreNames))
		meta.Track = &Track{
			TrackNumber: uint32(*assign(ctx.AppleMusicSongs.Attributes.TrackNumber, ctx.ItunesSong.TrackNumber)),
			TrackCount:  uint16(*ctx.ItunesSong.TrackCount),
		}
		meta.DiskNumber = &Disk{
			DiskNumber: uint32(*assign(ctx.AppleMusicSongs.Attributes.DiscNumber, ctx.ItunesSong.DiscNumber)),
			DiskCount:  uint16(*ctx.ItunesSong.DiscCount),
		}
		meta.Compilation = nil
		meta.PlayGap = ref(uint8(0))
		meta.ReleaseDate = assign(ctx.ItunesSong.ReleaseDate, ctx.AppleMusicSongs.Attributes.ReleaseDate)
		meta.AppleID = nil
		meta.Owner = nil
		meta.Copyright = assign(ctx.AppleMusicAlbum.Attributes.Copyright)
		meta.ItemID = assign(atoi(*ctx.AppleMusicSongs.ID), ref(uint32(*ctx.ItunesSong.TrackID)))
		meta.ArtistID = assign(ref(uint32(*ctx.ItunesSong.ArtistID)), atoi(*ctx.AppleMusicSongs.Relationships.Artists.Data[0].ID))
		meta.Rating = ref(uint8(0))
		meta.ComposerID = assign(atoi(*ctx.AppleMusicSongs.Relationships.Composers.Data[0].ID))
		meta.PlaylistID = assign(ref(uint32(*ctx.ItunesSong.CollectionID)), atoi(*ctx.AppleMusicAlbum.ID))
		meta.GenreID = assign(ref(uint32(quicktime.GetGenreID(ctx.AppleMusicSongs.Attributes.GenreNames))))
		meta.StorefrontID = ref(uint32(quicktime.GetStorefrontID(config.Storefront)))
		meta.MediaType = ref(uint8(quicktime.MediaTypeNormal_Music))
		meta.PurchaseDate = nil
		meta.SortName = assign(ctx.AppleMusicSongs.Attributes.Name, ctx.ItunesSong.TrackName)
		meta.SortAlbum = assign(ctx.AppleMusicSongs.Attributes.AlbumName, ctx.AppleMusicAlbum.Attributes.Name, ctx.ItunesSong.CollectionName)
		meta.SortArtist = assign(ctx.AppleMusicSongs.Attributes.ArtistName, ctx.ItunesSong.ArtistName)
		meta.SortComposer = assign(ctx.AppleMusicSongs.Attributes.ComposerName)
		meta.XID = nil
		meta.Cover = ctx.CoverData
		if *ctx.AppleMusicSongs.Attributes.HasLyrics && len(ctx.LyricsData) != 0 {
			meta.Lyrics = &ctx.LyricsData
		}
	}
	return
}

func LoadMusicVideoMetadata(ctx Context) (meta *Metadata) {
	meta = &Metadata{}
	meta.Title = assign(ctx.AppleMusicMusicVideos.Attributes.Name, ctx.ItunesMusicVideo.TrackName)
	meta.ArtistName = assign(ctx.AppleMusicMusicVideos.Attributes.ArtistName, ctx.ItunesMusicVideo.ArtistName)
	if ctx.Type == MusicVideoTypeAlbumRelated {
		meta.PlaylistArtist = assign(ctx.AppleMusicAlbum.Attributes.ArtistName)
		meta.AlbumName = assign(ctx.AppleMusicMusicVideos.Attributes.AlbumName, ctx.AppleMusicAlbum.Attributes.Name, ctx.ItunesMusicVideo.CollectionName)
	}
	meta.Work = assign(ctx.AppleMusicMusicVideos.Attributes.WorkName)
	meta.Genre = genre(quicktime.GetGenreID(ctx.AppleMusicMusicVideos.Attributes.GenreNames))
	if ctx.Type == MusicVideoTypeAlbumRelated {
		meta.Track = &Track{
			TrackNumber: uint32(*assign(ctx.AppleMusicMusicVideos.Attributes.TrackNumber, ctx.ItunesMusicVideo.TrackNumber)),
			TrackCount:  uint16(*ctx.ItunesMusicVideo.TrackCount),
		}
		meta.DiskNumber = &Disk{
			DiskNumber: uint32(*assign(ctx.AppleMusicMusicVideos.Attributes.DiscNumber, ctx.ItunesMusicVideo.DiscNumber)),
			DiskCount:  uint16(*ctx.ItunesMusicVideo.DiscCount),
		}
	}
	meta.Compilation = nil
	meta.PlayGap = ref(uint8(0))
	meta.ReleaseDate = assign(ctx.ItunesMusicVideo.ReleaseDate, ctx.AppleMusicMusicVideos.Attributes.ReleaseDate)
	meta.AppleID = nil
	meta.Owner = nil
	if ctx.Type == MusicVideoTypeAlbumRelated {
		meta.Copyright = assign(ctx.AppleMusicAlbum.Attributes.Copyright)
	}
	meta.ItemID = assign(atoi(*ctx.AppleMusicMusicVideos.ID), ref(uint32(*ctx.ItunesMusicVideo.TrackID)))
	meta.ArtistID = assign(ref(uint32(*ctx.ItunesMusicVideo.ArtistID)), atoi(*ctx.AppleMusicMusicVideos.Relationships.Artists.Data[0].ID))
	meta.Rating = ref(uint8(0))
	if ctx.Type == MusicVideoTypeAlbumRelated {
		meta.PlaylistID = assign(ref(uint32(*ctx.ItunesMusicVideo.CollectionID)), atoi(*ctx.AppleMusicAlbum.ID))
	}
	meta.GenreID = assign(ref(uint32(quicktime.GetGenreID(ctx.AppleMusicMusicVideos.Attributes.GenreNames))))
	meta.StorefrontID = ref(uint32(quicktime.GetStorefrontID(config.Storefront)))
	meta.HDVideo = nil
	meta.MediaType = ref(uint8(quicktime.MediaTypeMusicVideo))
	meta.PurchaseDate = nil
	meta.SortName = assign(ctx.AppleMusicMusicVideos.Attributes.Name, ctx.ItunesMusicVideo.TrackName)
	if ctx.Type == MusicVideoTypeAlbumRelated {
		meta.SortAlbum = assign(ctx.AppleMusicMusicVideos.Attributes.AlbumName, ctx.AppleMusicAlbum.Attributes.Name, ctx.ItunesMusicVideo.CollectionName)
	}
	meta.SortArtist = assign(ctx.AppleMusicMusicVideos.Attributes.ArtistName, ctx.ItunesMusicVideo.ArtistName)
	meta.XID = nil
	meta.Flavor = nil
	meta.Cover = ctx.CoverData
	return
}
