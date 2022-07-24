package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/andrewwillette/willette_api/persistence"
	"github.com/stretchr/testify/require"
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
	GetAllSoundcloudUrlsFunc     func() ([]persistence.SoundcloudUrl, error)
	AddSoundcloudUrlsFunc        func(s string) error
	DeleteSoundcloudUrlFunc      func(s string) error
	UpdateSoundcloudUiOrdersFunc func([]persistence.SoundcloudUrl) error
	SoundcloudUrls               []persistence.SoundcloudUrl
}

func (m *MockSoundcloudUrlService) GetAllSoundcloudUrls() ([]persistence.SoundcloudUrl, error) {
	return m.GetAllSoundcloudUrlsFunc()
}

func (m *MockSoundcloudUrlService) AddSoundcloudUrl(s string) error {
	return m.AddSoundcloudUrlsFunc(s)
}

func (m MockSoundcloudUrlService) DeleteSoundcloudUrl(s string) error {
	return m.DeleteSoundcloudUrlFunc(s)
}

func (m MockSoundcloudUrlService) UpdateSoundcloudUiOrders(urls []persistence.SoundcloudUrl) error {
	return m.UpdateSoundcloudUiOrdersFunc(urls)
}

func TestLogin_InvalidUser(t *testing.T) {
	body := UserJson{Username: "hello", Password: "passwordWorld"}
	var users []UserJson
	userService := &MockUserService{
		UsersRegistered: users,
		LoginFunc: func(username, password string) (success bool, bearerToken string) {
			return false, ""
		},
		IsAuthorizedFunc: func(bearerToken string) bool {
			return true
		},
	}
	var soundcloudUrls []persistence.SoundcloudUrl
	soundcloudUrlService := &MockSoundcloudUrlService{
		SoundcloudUrls: soundcloudUrls,
	}
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, loginEndpoint, userToJSON(body))
	webServices := newWebServices(userService, soundcloudUrlService)
	e := newServer(*webServices)
	e.ServeHTTP(response, request)
	require.Equal(t, 401, response.Code)
}

func TestLogin_SuccessfulLogin(t *testing.T) {
	var users []UserJson
	testBearerToken := "testBearerToken"
	userService := &MockUserService{
		UsersRegistered: users,
		LoginFunc: func(username, password string) (success bool, bearerToken string) {
			return true, testBearerToken
		},
		IsAuthorizedFunc: func(bearerToken string) bool {
			return true
		},
	}
	var soundcloudUrls []persistence.SoundcloudUrl
	soundcloudUrlService := &MockSoundcloudUrlService{
		SoundcloudUrls: soundcloudUrls,
	}
	webServices := newWebServices(userService, soundcloudUrlService)
	response := httptest.NewRecorder()
	body := UserJson{Username: "hello", Password: "passwordWorld"}
	request := httptest.NewRequest(http.MethodPost, loginEndpoint, userToJSON(body))

	e := newServer(*webServices)
	e.ServeHTTP(response, request)
	require.Equal(t, 200, response.Code)
	require.Contains(t, response.Body.String(), testBearerToken)
}

func TestLogin_InvalidHttpRequests(t *testing.T) {
	var users []UserJson
	testBearerToken := "testBearerToken"
	userService := &MockUserService{
		UsersRegistered: users,
		LoginFunc: func(username, password string) (success bool, bearerToken string) {
			return true, testBearerToken
		},
		IsAuthorizedFunc: func(bearerToken string) bool {
			return true
		},
	}
	var soundcloudUrls []persistence.SoundcloudUrl
	soundcloudUrlService := &MockSoundcloudUrlService{
		SoundcloudUrls: soundcloudUrls,
	}
	webServices := newWebServices(userService, soundcloudUrlService)
	body := UserJson{Username: "hello", Password: "passwordWorld"}
	e := newServer(*webServices)
	request := httptest.NewRequest(http.MethodPut, loginEndpoint, userToJSON(body))
	response := httptest.NewRecorder()
	e.ServeHTTP(response, request)
	require.Equal(t, 405, response.Code)
}

func TestAddSouncloudUrl_Success(t *testing.T) {
	var users []UserJson
	testBearerToken := "testBearerToken"
	userService := &MockUserService{
		UsersRegistered: users,
		LoginFunc: func(username, password string) (success bool, bearerToken string) {
			return true, testBearerToken
		},
		IsAuthorizedFunc: func(bearerToken string) bool {
			return true
		},
	}
	soundcloudUrls := []persistence.SoundcloudUrl{{Url: "urlone.com", UiOrder: 1, Id: 1},
		{Url: "urltwo.com", Id: 2, UiOrder: 3}}
	soundcloudUrlService := &MockSoundcloudUrlService{
		SoundcloudUrls: soundcloudUrls,
		GetAllSoundcloudUrlsFunc: func() ([]persistence.SoundcloudUrl, error) {
			return soundcloudUrls, nil
		},
		AddSoundcloudUrlsFunc: func(s string) error {
			soundcloudUrl := persistence.SoundcloudUrl{Url: s}
			soundcloudUrls = append(soundcloudUrls, soundcloudUrl)
			return nil
		},
		DeleteSoundcloudUrlFunc: func(s string) error {
			return nil
		},
	}
	webServices := newWebServices(userService, soundcloudUrlService)
	newSoundcloudUrl := "testsoundcloudurl.com"
	body := SoundcloudUrlJson{Url: newSoundcloudUrl}
	addSdcldUrlReq := httptest.NewRequest(http.MethodPut, addSoundcloudEndpoint, authenticatedSoundcloudUrlToJSON(body))
	addSdcldUrlResp := httptest.NewRecorder()
	e := newServer(*webServices)
	e.ServeHTTP(addSdcldUrlResp, addSdcldUrlReq)
	require.Equal(t, 200, addSdcldUrlResp.Code)
	responseTwo := httptest.NewRecorder()
	requestTwo := httptest.NewRequest(http.MethodGet, getSoundcloudAllEndpoint, nil)
	e.ServeHTTP(responseTwo, requestTwo)
	decoder := json.NewDecoder(responseTwo.Body)
	var soundcloudData []persistence.SoundcloudUrl
	err := decoder.Decode(&soundcloudData)
	if err != nil {
		t.Log("failed to decode soundcloud data")
		t.Fail()
	}
	require.ElementsMatch(t, soundcloudData, []persistence.SoundcloudUrl{
		{Id: 0, Url: "urlone.com", UiOrder: 1},
		{Id: 0, Url: "urltwo.com", UiOrder: 3},
		{Id: 0, Url: "testsoundcloudurl.com", UiOrder: 0}})
}

func TestGetAllSoundcloudUrls_ServiceErr(t *testing.T) {
	var users []UserJson
	testBearerToken := "testBearerToken"
	userService := &MockUserService{
		UsersRegistered: users,
		LoginFunc: func(username, password string) (success bool, bearerToken string) {
			return true, testBearerToken
		},
		IsAuthorizedFunc: func(bearerToken string) bool {
			return true
		},
	}
	soundcloudUrls := []persistence.SoundcloudUrl{{Url: "urlone.com", UiOrder: 1, Id: 1},
		{Url: "urltwo.com", Id: 2, UiOrder: 3}}
	soundcloudUrlService := &MockSoundcloudUrlService{
		SoundcloudUrls: soundcloudUrls,
		GetAllSoundcloudUrlsFunc: func() ([]persistence.SoundcloudUrl, error) {
			return nil, errors.New("Failed to get soundcloudUrls")
		},
		AddSoundcloudUrlsFunc: func(s string) error {
			soundcloudUrl := persistence.SoundcloudUrl{Url: s}
			soundcloudUrls = append(soundcloudUrls, soundcloudUrl)
			return nil
		},
		DeleteSoundcloudUrlFunc: func(s string) error {
			return nil
		},
	}
	webServices := newWebServices(userService, soundcloudUrlService)
	newSoundcloudUrl := "testsoundcloudurl.com"
	body := SoundcloudUrlJson{Url: newSoundcloudUrl}
	addSdcldUrlReq := httptest.NewRequest(http.MethodPut, addSoundcloudEndpoint, authenticatedSoundcloudUrlToJSON(body))
	addSdcldUrlResp := httptest.NewRecorder()
	e := newServer(*webServices)
	e.ServeHTTP(addSdcldUrlResp, addSdcldUrlReq)
	require.Equal(t, 200, addSdcldUrlResp.Code)

	responseTwo := httptest.NewRecorder()
	requestTwo := httptest.NewRequest(http.MethodGet, getSoundcloudAllEndpoint, nil)
	e.ServeHTTP(responseTwo, requestTwo)
	require.Equal(t, 500, responseTwo.Code)
	require.Equal(t, "Failed to get soundcloud urls from service.", responseTwo.Body.String())
}

func authenticatedSoundcloudUrlToJSON(url SoundcloudUrlJson) io.Reader {
	marshalledUser, _ := json.Marshal(url)
	return bytes.NewReader(marshalledUser)
}

func userToJSON(user UserJson) io.Reader {
	marshalledUser, _ := json.Marshal(user)
	return bytes.NewReader(marshalledUser)
}
