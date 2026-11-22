// Package gitproxy forwards requests for the git subdomain to a self-hosted
// Forgejo instance, so the website's Echo server stays the only TLS endpoint.
package gitproxy

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/labstack/echo/v4"
	zlog "github.com/rs/zerolog/log"
)

// transferTimeout replaces the website's short server read/write timeouts for
// proxied requests, so large pushes and clones aren't cut off mid-transfer.
const transferTimeout = 15 * time.Minute

// Middleware returns Pre middleware that hands every request addressed to host
// to the git server at upstream, bypassing the website's routes and middleware.
// Requests for any other host fall through to the website.
func Middleware(host, upstream string) (echo.MiddlewareFunc, error) {
	target, err := url.Parse(upstream)
	if err != nil || target.Scheme == "" || target.Host == "" {
		return nil, fmt.Errorf("invalid git upstream URL %q", upstream)
	}
	proxy := &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetURL(target)
			r.SetXForwarded()
			r.Out.Host = r.In.Host
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			zlog.Error().Err(err).Str("path", r.URL.Path).Msg("git proxy request failed")
			w.WriteHeader(http.StatusBadGateway)
		},
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			if hostname(req.Host) != host {
				return next(c)
			}
			rc := http.NewResponseController(c.Response())
			deadline := time.Now().Add(transferTimeout)
			if err := rc.SetReadDeadline(deadline); err != nil {
				zlog.Warn().Err(err).Msg("git proxy: could not extend read deadline")
			}
			if err := rc.SetWriteDeadline(deadline); err != nil {
				zlog.Warn().Err(err).Msg("git proxy: could not extend write deadline")
			}
			proxy.ServeHTTP(c.Response(), req)
			return nil
		}
	}, nil
}

// hostname strips any port from a Host header value.
func hostname(hostport string) string {
	if h, _, err := net.SplitHostPort(hostport); err == nil {
		return h
	}
	return hostport
}
