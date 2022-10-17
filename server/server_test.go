package server

import (
	"bytes"
	"encoding/json"
	"io"
)

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
