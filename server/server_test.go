package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

func TestHandleMusicPage(t *testing.T) {
	e := echo.New()
	e.Renderer = getRenderer()
	req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(""))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := handleMusicPage(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "Andrew Willette")
}

func TestHandleHomePage(t *testing.T) {
	e := echo.New()
	e.Renderer = getRenderer()
	req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(""))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
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
	e.Renderer = getRenderer()
	req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(""))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := handleResumePage(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusPermanentRedirect, rec.Code)
}
