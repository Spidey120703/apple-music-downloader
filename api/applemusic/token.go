package applemusic

var Authorization string

func SetAuthorization(token string) {
	Authorization = "Bearer " + token
}
