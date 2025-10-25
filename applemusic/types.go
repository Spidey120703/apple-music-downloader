package applemusic

import "encoding/json"

type Resource struct {
	ID   *string `json:"id"`
	Type *string `json:"type"`
	Href *string `json:"href,omitempty"`
}

type Artwork struct {
	BgColor  *string `json:"bgColor,omitempty"`
	Gradient *struct {
		Color *string  `json:"color,omitempty"`
		Y2    *float64 `json:"y2,omitempty"`
	} `json:"gradient,omitempty"`
	HasP3                *bool    `json:"hasP3,omitempty"`
	Height               *int     `json:"height,omitempty"`
	ImageTraits          []string `json:"imageTraits,omitempty"`
	RecommendedCropCodes []string `json:"recommendedCropCodes,omitempty"`
	TextColor1           *string  `json:"textColor1,omitempty"`
	TextColor2           *string  `json:"textColor2,omitempty"`
	TextColor3           *string  `json:"textColor3,omitempty"`
	TextColor4           *string  `json:"textColor4,omitempty"`
	TextGradient         []string `json:"textGradient,omitempty"`
	URL                  *string  `json:"url,omitempty"`
	Width                *int     `json:"width,omitempty"`
}

type MotionVideo struct {
	PreviewFrame *Artwork `json:"previewFrame,omitempty"`
	Video        *string  `json:"video,omitempty"`
}

type Offers struct {
	BuyParams      *string  `json:"buyParams,omitempty"`
	Price          *float64 `json:"price,omitempty"`
	PriceFormatted *string  `json:"priceFormatted,omitempty"`
	Type           *string  `json:"type,omitempty"`
}

type PlayParameters struct {
	ID   *string `json:"id"`
	Kind *string `json:"kind"`
}

type EditorialNotes struct {
	Name     *string `json:"name,omitempty"`
	Short    *string `json:"short,omitempty"`
	Standard *string `json:"standard,omitempty"`
	Tagline  *string `json:"tagline,omitempty"`
}

type Preview struct {
	Artwork *Artwork `json:"artwork,omitempty"`
	HlsURL  *string  `json:"hlsUrl,omitempty"`
	URL     *string  `json:"url"`
}

type Relationship[T any] struct {
	Href *string         `json:"href,omitempty"`
	Next *string         `json:"next,omitempty"`
	Data []T             `json:"data"`
	Meta json.RawMessage `json:"meta,omitempty"`
}

type Relationships struct {
	Albums       *Relationship[Albums]       `json:"albums,omitempty"`
	Artists      *Relationship[Artists]      `json:"artists,omitempty"`
	Composers    *Relationship[Artists]      `json:"composers,omitempty"`
	Credits      *Relationship[Credits]      `json:"credits,omitempty"`
	Genres       *Relationship[Genres]       `json:"genres,omitempty"`
	MusicVideos  *Relationship[MusicVideos]  `json:"music-videos,omitempty"`
	RecordLabels *Relationship[RecordLabels] `json:"record-labels,omitempty"`
	Tracks       *Relationship[Songs]        `json:"tracks,omitempty"`
}

type Meta struct {
	ContentVersion *struct {
		MZINDEXER *int64 `json:"MZ_INDEXER"`
		RTCI      *int64 `json:"RTCI,omitempty"`
	} `json:"contentVersion,omitempty"`
	FormerIds  []string `json:"formerIds,omitempty"`
	Order      []string `json:"order,omitempty"`
	Popularity *float64 `json:"popularity,omitempty"`
}

type EditorialArtwork struct {
	BannerUber            *Artwork `json:"bannerUber,omitempty"`
	OriginalFlowcaseBrick *Artwork `json:"originalFlowcaseBrick,omitempty"`
	StaticDetailSquare    *Artwork `json:"staticDetailSquare,omitempty"`
	StaticDetailTall      *Artwork `json:"staticDetailTall,omitempty"`
	StoreFlowcase         *Artwork `json:"storeFlowcase,omitempty"`
	SubscriptionHero      *Artwork `json:"subscriptionHero,omitempty"`
	SuperHeroTall         *Artwork `json:"superHeroTall,omitempty"`
}

type Albums struct {
	Resource
	Attributes *struct {
		ArtistName       *string           `json:"artistName"`
		ArtistUrl        *string           `json:"artistUrl,omitempty"`
		Artwork          *Artwork          `json:"artwork"`
		AudioTraits      []string          `json:"audioTraits,omitempty"`
		ContentRating    *string           `json:"contentRating,omitempty"`
		Copyright        *string           `json:"copyright,omitempty"`
		EditorialArtwork *EditorialArtwork `json:"editorialArtwork,omitempty"`
		EditorialNotes   *EditorialNotes   `json:"editorialNotes,omitempty"`
		EditorialVideo   *struct {
			MotionDetailSquare   *MotionVideo `json:"motionDetailSquare,omitempty"`
			MotionDetailTall     *MotionVideo `json:"motionDetailTall,omitempty"`
			MotionSquareVideo1X1 *MotionVideo `json:"motionSquareVideo1x1,omitempty"`
		} `json:"editorialVideo,omitempty"`
		GenreNames          []string        `json:"genreNames"`
		IsCompilation       *bool           `json:"isCompilation"`
		IsComplete          *bool           `json:"isComplete"`
		IsMasteredForItunes *bool           `json:"isMasteredForItunes"`
		IsPrerelease        *bool           `json:"isPrerelease"`
		IsSingle            *bool           `json:"isSingle"`
		Name                *string         `json:"name"`
		Offers              []Offers        `json:"offers,omitempty"`
		PlainEditorialNotes *EditorialNotes `json:"plainEditorialNotes,omitempty"`
		PlayParams          *PlayParameters `json:"playParams,omitempty"`
		RecordLabel         *string         `json:"recordLabel,omitempty"`
		ReleaseDate         *string         `json:"releaseDate,omitempty"`
		TrackCount          *int            `json:"trackCount"`
		Upc                 *string         `json:"upc,omitempty"`
		URL                 *string         `json:"url"`
	} `json:"attributes,omitempty"`
	Relationships *Relationships `json:"relationships,omitempty"`
	Meta          *Meta          `json:"meta,omitempty"`
}

type Artists struct {
	Resource
	Attributes *struct {
		ArtistBio        *string           `json:"artistBio,omitempty"`
		Artwork          *Artwork          `json:"artwork,omitempty"`
		BornOrFormed     *string           `json:"bornOrFormed,omitempty"`
		EditorialArtwork *EditorialArtwork `json:"editorialArtwork,omitempty"`
		EditorialNotes   *EditorialNotes   `json:"editorialNotes,omitempty"`
		EditorialVideo   *struct {
			MotionArtistFullscreen16X9 *MotionVideo `json:"motionArtistFullscreen16x9,omitempty"`
			MotionArtistSquare1X1      *MotionVideo `json:"motionArtistSquare1x1,omitempty"`
			MotionArtistWide16X9       *MotionVideo `json:"motionArtistWide16x9,omitempty"`
		} `json:"editorialVideo,omitempty"`
		GenreNames []string `json:"genreNames"`
		Hero       []struct {
			Content []struct {
				Artwork *Artwork `json:"artwork,omitempty"`
			} `json:"content,omitempty"`
		} `json:"hero,omitempty"`
		IsGroup *bool   `json:"isGroup"`
		Name    *string `json:"name"`
		URL     *string `json:"url"`
	} `json:"attributes,omitempty"`
	Relationships *Relationships `json:"relationships,omitempty"`
	Meta          *Meta          `json:"meta,omitempty"`
}

type Songs struct {
	Resource
	Attributes *struct {
		AlbumName         *string           `json:"albumName"`
		ArtistName        *string           `json:"artistName"`
		ArtistUrl         *string           `json:"artistUrl,omitempty"`
		Artwork           *Artwork          `json:"artwork"`
		Attribution       *string           `json:"attribution,omitempty"`
		AudioLocale       *string           `json:"audioLocale,omitempty"`
		AudioTraits       []string          `json:"audioTraits,omitempty"`
		ComposerName      *string           `json:"composerName,omitempty"`
		DiscNumber        *int              `json:"discNumber,omitempty"`
		DurationInMillis  *int              `json:"durationInMillis"`
		EditorialArtwork  *EditorialArtwork `json:"editorialArtwork,omitempty"`
		ExtendedAssetUrls *struct {
			EnhancedHls      *string `json:"enhancedHls,omitempty"`
			Lightweight      *string `json:"lightweight,omitempty"`
			LightweightPlus  *string `json:"lightweightPlus,omitempty"`
			Plus             *string `json:"plus,omitempty"`
			SuperLightweight *string `json:"superLightweight,omitempty"`
		} `json:"extendedAssetUrls,omitempty"`
		GenreNames                []string        `json:"genreNames"`
		HasLyrics                 *bool           `json:"hasLyrics"`
		HasTimeSyncedLyrics       *bool           `json:"hasTimeSyncedLyrics"`
		IsAppleDigitalMaster      *bool           `json:"isAppleDigitalMaster"`
		IsMasteredForItunes       *bool           `json:"isMasteredForItunes"`
		IsVocalAttenuationAllowed *bool           `json:"isVocalAttenuationAllowed"`
		Isrc                      *string         `json:"isrc,omitempty"`
		MovementCount             *int            `json:"movementCount,omitempty"`
		MovementName              *string         `json:"movementName,omitempty"`
		MovementNumber            *int            `json:"movementNumber,omitempty"`
		Name                      *string         `json:"name"`
		Offers                    []Offers        `json:"offers,omitempty"`
		PlayParams                *PlayParameters `json:"playParams,omitempty"`
		Previews                  []Preview       `json:"previews"`
		ReleaseDate               *string         `json:"releaseDate,omitempty"`
		TrackNumber               *int            `json:"trackNumber,omitempty"`
		URL                       *string         `json:"url"`
		WorkName                  *string         `json:"workName,omitempty"`
	} `json:"attributes,omitempty"`
	Relationships *Relationships `json:"relationships,omitempty"`
	Meta          *Meta          `json:"meta,omitempty"`
}

type Credits struct {
	Resource
	Attributes *struct {
		Kind  *string `json:"kind,omitempty"`
		Title *string `json:"title,omitempty"`
	} `json:"attributes,omitempty"`
	Relationships *struct {
		CreditArtists *struct {
			Data []struct {
				ID         *string `json:"id,omitempty"`
				Type       *string `json:"type,omitempty"`
				Attributes *struct {
					Artwork   *Artwork `json:"artwork,omitempty"`
					Name      *string  `json:"name,omitempty"`
					RoleNames []string `json:"roleNames,omitempty"`
				} `json:"attributes,omitempty"`
			} `json:"data,omitempty"`
		} `json:"credit-artists,omitempty"`
	} `json:"relationships,omitempty"`
}

type Genres struct {
	Resource
	Attributes *struct {
		Name       *string `json:"name,omitempty"`
		ParentId   *string `json:"parentId,omitempty"`
		ParentName *string `json:"parentName,omitempty"`
		URL        *string `json:"url,omitempty"`
	} `json:"attributes,omitempty"`
}

type MusicVideos struct {
	Resource
	Attributes *struct {
		AlbumName        *string           `json:"albumName,omitempty"`
		ArtistName       *string           `json:"artistName"`
		ArtistUrl        *string           `json:"artistUrl,omitempty"`
		Artwork          *Artwork          `json:"artwork"`
		DiscNumber       *int              `json:"discNumber,omitempty"`
		DurationInMillis *int              `json:"durationInMillis"`
		EditorialArtwork *EditorialArtwork `json:"editorialArtwork,omitempty"`
		GenreNames       []string          `json:"genreNames"`
		Has4K            *bool             `json:"has4K"`
		HasHDR           *bool             `json:"hasHDR"`
		Isrc             *string           `json:"isrc,omitempty"`
		Name             *string           `json:"name"`
		Offers           []Offers          `json:"offers,omitempty"`
		PlayParams       *PlayParameters   `json:"playParams,omitempty"`
		Previews         []Preview         `json:"previews"`
		ReleaseDate      *string           `json:"releaseDate,omitempty"`
		TrackNumber      *int              `json:"trackNumber,omitempty"`
		URL              *string           `json:"url"`
		VideoTraits      []string          `json:"videoTraits,omitempty"`
		VideoSubType     *string           `json:"videoSubType,omitempty"`
		WorkId           *string           `json:"workId,omitempty"`
		WorkName         *string           `json:"workName,omitempty"`
	} `json:"attributes,omitempty"`
}

type RecordLabels struct {
	Resource
	Attributes *struct {
		Artwork     *Artwork `json:"artwork"`
		Description *struct {
			Short    *string `json:"short,omitempty"`
			Standard *string `json:"standard"`
		} `json:"description,omitempty"`
		Name *string `json:"name"`
		URL  *string `json:"url"`
	} `json:"attributes,omitempty"`
}
