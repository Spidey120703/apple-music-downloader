package quicktime

type MediaType uint8

const (
	MediaTypeMovie_old MediaType = iota
	MediaTypeNormal_Music
	MediaTypeAudiobook
	_
	_
	MediaTypeWhackedBookmark
	MediaTypeMusicVideo
	_
	_
	MediaTypeMovie
	MediaTypeTVShow
	MediaTypeBooklet
	_
	_
	MediaTypeRingtone
	MediaTypePodcast = iota + 6
	_
	MediaTypeiTunesU
)
