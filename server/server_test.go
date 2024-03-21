package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/andrewwillette/keyofday/key"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

func TestHandleMusicPage(t *testing.T) {
	e := echo.New()
	e.Renderer = getTemplateRenderer()
	req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(""))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := handleMusicPage(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "Andrew Willette")
	require.Contains(t, rec.Body.String(), "© 2024 Andrew Willette. All rights reserved.")
	require.Equal(t, true, false)
}

func TestHandleHomePage(t *testing.T) {
	e := echo.New()
	e.Renderer = getTemplateRenderer()
	req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(""))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := handleHomePage(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "Andrew Willette")
	require.Contains(t, rec.Body.String(), "Madison, Wisconsin")
}

func TestHandleResumePage(t *testing.T) {
	e := echo.New()
	e.Renderer = getTemplateRenderer()
	req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(""))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := handleResumePage(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusPermanentRedirect, rec.Code)
}

func TestHandleKeyOfDayPage(t *testing.T) {
	e := echo.New()
	e.Renderer = getTemplateRenderer()
	req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(""))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := handleKeyOfDayPage(c)
	require.NoError(t, err)
	require.Contains(t, rec.Body.String(), key.TodaysKey())
}

func BenchmarkHandleHomePage(b *testing.B) {
	e := echo.New()
	e.Renderer = getTemplateRenderer()
	req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(""))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	for n := 0; n < b.N; n++ {
		err := handleResumePage(c)
		require.NoError(b, err)
	}
}
