package server

type SoundcloudUrlUiOrderJson struct {
	Url     string `json:"url"`
	UiOrder int    `json:"uiOrder"`
}

type UserJson struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	BearerToken string `json:"bearerToken"`
}

type BearerTokenJson struct {
	BearerToken string `json:"bearerToken"`
}

type SoundcloudUrlJson struct {
	Url string `json:"url"`
}
