package server

import (
	"bytes"
	"encoding/json"
	"io"
)

type MockUserService struct {
	UsersRegistered  []UserJson
	LoginFunc        func(username string, password string) (success bool, bearerToken string)
	IsAuthorizedFunc func(bearerToken string) bool
}

func (m *MockUserService) Login(username, password string) (success bool, bearerToken string) {
	return m.LoginFunc(username, password)
}

func (m *MockUserService) IsAuthorized(bearerToken string) bool {
	return m.IsAuthorizedFunc(bearerToken)
}

type MockSoundcloudUrlService struct {
	AddSoundcloudUrlsFunc   func(s string) error
	DeleteSoundcloudUrlFunc func(s string) error
}

func (m *MockSoundcloudUrlService) AddSoundcloudUrl(s string) error {
	return m.AddSoundcloudUrlsFunc(s)
}

func (m MockSoundcloudUrlService) DeleteSoundcloudUrl(s string) error {
	return m.DeleteSoundcloudUrlFunc(s)
}

func authenticatedSoundcloudUrlToJSON(url SoundcloudUrlJson) io.Reader {
	marshalledUser, _ := json.Marshal(url)
	return bytes.NewReader(marshalledUser)
}

func userToJSON(user UserJson) io.Reader {
	marshalledUser, _ := json.Marshal(user)
	return bytes.NewReader(marshalledUser)
}
