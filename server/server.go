package server

import (
	"net/http"
	"path/filepath"
	"runtime"
	"time"

	"github.com/andrewwillette/keyofday/key"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	zlog "github.com/rs/zerolog/log"
	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/time/rate"

	"github.com/andrewwillette/andrewwillettedotcom/aws"
	"github.com/andrewwillette/andrewwillettedotcom/config"
	applog "github.com/andrewwillette/andrewwillettedotcom/log"
	"github.com/andrewwillette/andrewwillettedotcom/server/blog"
	"github.com/andrewwillette/andrewwillettedotcom/server/chordcharts"
	"github.com/andrewwillette/andrewwillettedotcom/server/echopprof"
	"github.com/andrewwillette/andrewwillettedotcom/server/gitproxy"
	"github.com/andrewwillette/andrewwillettedotcom/server/traffic"
)

const (
	homeEndpoint = "/"

	musicEndpoint = "/music"

	resumeEndpoint = "/resume"

	sheetmusicEndpoint = "/sheet-music"

	showsEndpoint = "/shows"

	blogsEndpoint   = "/blog"
	blogEndpoint    = "/blog/:blog"
	blogRssEndpoint = "/blog/rss"

	chordChartsEndpoint = "/chord-charts"
	chordChartEndpoint  = "/chord-charts/:chart"

	cssEndpoint = "/static/main.css"
	cssResource = "server/static/main.css"

	chordChartsJsEndpoint = "/static/chordcharts.js"
	chordChartsJsResource = "server/static/chordcharts.js"
	fuzzySearchJsEndpoint = "/static/fuzzysearch.js"
	fuzzySearchJsResource = "server/static/fuzzysearch.js"

	robotsEndpoint    = "/robots.txt"
	robotsTxtResource = "server/static/robots.txt"

	keyOfDayEndpoint = "/key-of-the-day"

	adminEndpoint = "/admin/traffic"
)

var (
	_, b, _, _ = runtime.Caller(0)
	basepath   = filepath.Dir(b)
)

// StartServer starts the Echo server. Pass sslEnabled=true in production to
// enable AutoTLS; false runs plain HTTP on port 80.
func StartServer(sslEnabled bool) {
	if err := traffic.RestoreFromS3IfEmpty(config.C.TrafficDBPath); err != nil {
		zlog.Error().Err(err).Msg("failed to restore traffic database from S3")
	}
	if err := traffic.InitDB(config.C.TrafficDBPath); err != nil {
		zlog.Error().Err(err).Msg("failed to initialize traffic database")
	}
	traffic.StartBackupJob()

	e := echo.New()
	e.HideBanner = true
	// No reverse proxy sits in front of the server, so never trust
	// client-supplied X-Forwarded-For / X-Real-IP headers (GO-2026-6872).
	e.IPExtractor = echo.ExtractIPDirect()
	e.Logger = applog.NewEchoAdapter(zlog.Logger)
	e.Renderer = newTemplateRenderer()

	addMiddleware(e)
	addRoutes(e)

	if config.C.PProfEnabled {
		echopprof.Wrap(e)
	}

	blog.InitializeBlogs()
	chordcharts.InitializeChordCharts()
	aws.UpdateAudioCache()
	aws.UpdateSheetMusicCache()
	aws.UpdateShowsCache()
	go aws.UpdateAudioCacheOnPresignExpiry()
	aws.StartSQSPoller()
	aws.StartSheetMusicLinkRefreshJob()

	const (
		readTimeout  = 10 * time.Second
		writeTimeout = 30 * time.Second
		idleTimeout  = 120 * time.Second
	)

	if sslEnabled {
		e.Pre(middleware.HTTPSRedirect())
	}
	// Registered after HTTPSRedirect so git traffic is only ever proxied over TLS.
	addGitProxy(e)

	if sslEnabled {
		tlsHosts := []string{"andrewwillette.com"}
		if config.C.GitHost != "" {
			tlsHosts = append(tlsHosts, config.C.GitHost)
		}
		e.AutoTLSManager.HostPolicy = autocert.HostWhitelist(tlsHosts...)
		const sslCacheDir = "/var/www/.cache"
		e.AutoTLSManager.Cache = autocert.DirCache(sslCacheDir)
		e.TLSServer.ReadTimeout = readTimeout
		e.TLSServer.WriteTimeout = writeTimeout
		e.TLSServer.IdleTimeout = idleTimeout
		go func(c *echo.Echo) {
			redirectServer := &http.Server{
				Addr:         ":80",
				ReadTimeout:  readTimeout,
				WriteTimeout: writeTimeout,
				IdleTimeout:  idleTimeout,
			}
			e.Logger.Fatal(e.StartServer(redirectServer))
		}(e)
		e.Logger.Fatal(e.StartAutoTLS(":443"))
	} else {
		s := &http.Server{
			Addr:         ":80",
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
			IdleTimeout:  idleTimeout,
		}
		e.Logger.Fatal(e.StartServer(s))
	}
}

func addRoutes(e *echo.Echo) {
	e.GET(homeEndpoint, handleHomePage)
	e.GET(resumeEndpoint, handleResumePage)
	e.GET(musicEndpoint, handleRecordingsPage)
	e.GET(sheetmusicEndpoint, handleSheetmusicPage)
	e.GET(showsEndpoint, handleShowsPage)
	e.GET(keyOfDayEndpoint, handleKeyOfDayPage)
	e.GET(blogsEndpoint, blog.HandleBlogPage)
	e.GET(blogRssEndpoint, blog.HandleRssFeed)
	e.GET(blogEndpoint, blog.HandleIndividualBlogPage)
	e.GET(chordChartsEndpoint, chordcharts.HandleChordChartsPage)
	e.GET(chordChartEndpoint, chordcharts.HandleChordChartPage)
	e.File(cssEndpoint, cssResource)
	e.File(chordChartsJsEndpoint, chordChartsJsResource)
	e.File(fuzzySearchJsEndpoint, fuzzySearchJsResource)
	e.File(robotsEndpoint, robotsTxtResource)
	e.GET(adminEndpoint, traffic.HandleAdminPage, adminRateLimiter(), traffic.BasicAuthMiddleware())
}

func addGitProxy(e *echo.Echo) {
	if config.C.GitHost == "" {
		return
	}
	mw, err := gitproxy.Middleware(config.C.GitHost, config.C.GitUpstreamURL)
	if err != nil {
		zlog.Error().Err(err).Msg("git proxy disabled")
		return
	}
	e.Pre(mw)
	zlog.Info().Msgf("proxying %s to %s", config.C.GitHost, config.C.GitUpstreamURL)
}

// adminRateLimiter returns a rate limiter scoped to the admin route.
// Allows 5 requests per minute per IP to mitigate brute-force attacks.
func adminRateLimiter() echo.MiddlewareFunc {
	cfg := middleware.RateLimiterConfig{
		Skipper: middleware.DefaultSkipper,
		Store: middleware.NewRateLimiterMemoryStoreWithConfig(
			middleware.RateLimiterMemoryStoreConfig{
				Rate:      rate.Limit(5.0 / 60.0),
				Burst:     5,
				ExpiresIn: 5 * time.Minute,
			},
		),
		IdentifierExtractor: func(c echo.Context) (string, error) {
			return c.RealIP(), nil
		},
		ErrorHandler: func(c echo.Context, err error) error {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "rate limit exceeded"})
		},
		DenyHandler: func(c echo.Context, identifier string, err error) error {
			return c.JSON(http.StatusTooManyRequests, map[string]string{"error": "too many requests"})
		},
	}
	return middleware.RateLimiterWithConfig(cfg)
}

func addMiddleware(e *echo.Echo) {
	e.Use(logMiddleware)
	e.Use(traffic.TrackingMiddleware)
	e.Use(middleware.Secure())
}

type HomePageData struct {
	CurrentYear int
}

func handleHomePage(c echo.Context) error {
	return c.Render(http.StatusOK, "homepage", HomePageData{CurrentYear: time.Now().Year()})
}

func handleResumePage(c echo.Context) error {
	return c.Redirect(http.StatusPermanentRedirect, config.C.HomePageImageS3URL)
}

type KeyOfDayPage struct {
	KeyOfDay    string
	CurrentYear int
}

func handleKeyOfDayPage(c echo.Context) error {
	data := KeyOfDayPage{
		KeyOfDay:    key.TodaysKey(),
		CurrentYear: time.Now().Year(),
	}
	if err := c.Render(http.StatusOK, "keyofdaypage", data); err != nil {
		zlog.Error().Err(err).Msg("handleKeyOfDayPage render failed")
		return err
	}
	return nil
}

func logMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		zlog.Info().Msgf("Request to registered path %s with ip %s", c.Path(), c.RealIP())
		if err := next(c); err != nil {
			c.Error(err)
		}
		return nil
	}
}
