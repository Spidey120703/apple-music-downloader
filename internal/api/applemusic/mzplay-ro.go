package applemusic

type WebPlaybackRequest struct {
	SalableAdamId string `json:"salableAdamId"`
}

type WebPlaybackLicenseRequest struct {
	AdamId        string `json:"adamId"`
	IsLibrary     bool   `json:"isLibrary"`
	UserInitiated bool   `json:"user-initiated"`
	Challenge     string `json:"challenge"`
	Uri           string `json:"uri"`
	KeySystem     string `json:"key-system"`
}
