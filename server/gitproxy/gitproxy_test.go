package gitproxy

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

func newTestServer(t *testing.T, upstreamURL string) *echo.Echo {
	t.Helper()
	mw, err := Middleware("git.example.com", upstreamURL)
	require.NoError(t, err)
	e := echo.New()
	e.Pre(mw)
	e.GET("/", func(c echo.Context) error { return c.String(http.StatusOK, "website") })
	return e
}

func TestProxiesGitHost(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "forgejo "+r.Host+" "+r.Header.Get("X-Forwarded-Proto"))
	}))
	defer upstream.Close()
	e := newTestServer(t, upstream.URL)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Host = "git.example.com:443"
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "forgejo git.example.com:443 http", rec.Body.String())
}

func TestOtherHostsReachWebsite(t *testing.T) {
	e := newTestServer(t, "http://127.0.0.1:1")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Host = "example.com"
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "website", rec.Body.String())
}

func TestUpstreamDownReturnsBadGateway(t *testing.T) {
	e := newTestServer(t, "http://127.0.0.1:1")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Host = "git.example.com"
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadGateway, rec.Code)
}

func TestInvalidUpstream(t *testing.T) {
	_, err := Middleware("git.example.com", "not a url")
	require.Error(t, err)
}
