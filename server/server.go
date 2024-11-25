package server

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"path/filepath"
	"runtime"
	"time"

	"github.com/andrewwillette/keyofday/key"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/acme/autocert"

	cfg "github.com/andrewwillette/willette_api/cfg"
	"github.com/andrewwillette/willette_api/server/blog"
	"github.com/andrewwillette/willette_api/server/echopprof"
)

const (
	homeEndpoint       = "/"
	musicEndpoint      = "/music"
	resumeEndpoint     = "/resume"
	sheetmusicEndpoint = "/sheet-music"
	blogEndpoint       = "/blog"
	blogRssEndpoint    = "/blog/rss"
	cssEndpoint        = "/static/main.css"
	cssResource        = "static/main.css"
	keyOfDayEndpoint   = "/key-of-the-day"
	resumeResource     = "https://andrewwillette.s3.us-east-2.amazonaws.com/newdir/resume.pdf"
)

var (
	_, b, _, _ = runtime.Caller(0)
	basepath   = filepath.Dir(b)
)

// StartServer start the server with https certificate configurable
func StartServer(sslEnabled bool) {
	e := echo.New()
	e.HideBanner = true
	addRoutes(e)
	addMiddleware(e)
	if cfg.C.PProfEnabled {
		echopprof.Wrap(e)
	}
	e.Renderer = getTemplateRenderer()
	if sslEnabled {
		e.Pre(middleware.HTTPSRedirect())
		e.AutoTLSManager.HostPolicy = autocert.HostWhitelist("andrewwillette.com")
		// getSSLCacheDir return directory for ssl cache
		const sslCacheDir = "/var/www/.cache"
		e.AutoTLSManager.Cache = autocert.DirCache(sslCacheDir)
		go func(c *echo.Echo) {
			e.Logger.Fatal(e.Start(":80"))
		}(e)
		e.Logger.Fatal(e.StartAutoTLS(":443"))
	} else {
		e.Logger.Fatal(e.Start(":80"))
	}
}

// addRoutes adds routes to the echo webserver
func addRoutes(e *echo.Echo) {
	e.GET(homeEndpoint, handleHomePage)
	e.GET(resumeEndpoint, handleResumePage)
	e.GET(musicEndpoint, handleRecordingsPage)
	e.GET(sheetmusicEndpoint, handleSheetmusicPage)
	e.GET(keyOfDayEndpoint, handleKeyOfDayPage)
	e.GET(blogEndpoint, blog.HandleBlogPage)
	e.GET(blogRssEndpoint, blog.HandleRssFeed)
	e.GET("/blog/:blog", blog.HandleIndividualBlogPage)
	e.File(cssEndpoint, cssResource)
}

func addMiddleware(e *echo.Echo) {
	e.Use(logmiddleware)
}

// handleHomePage handles returning the homepage template
func handleHomePage(c echo.Context) error {
	err := c.Render(http.StatusOK, "homepage", nil)
	if err != nil {
		return err
	}
	return nil
}

// handleResumePage handles returning the resume template
func handleResumePage(c echo.Context) error {
	err := c.Redirect(http.StatusPermanentRedirect, resumeResource)
	if err != nil {
		return err
	}
	return nil
}

// handleKeyOfDayPage handles returning the key of the day
func handleKeyOfDayPage(c echo.Context) error {
	err := c.Render(http.StatusOK, "keyofdaypage", key.TodaysKey())
	if err != nil {
		return err
	}
	return nil
}

// Template is the template renderer for my echo webserver
type Template struct {
	templates *template.Template
}

// Render renders the template
func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

// getTemplateRenderer returns a template renderer for my echo webserver
func getTemplateRenderer() *Template {
	templates := template.New("")
	if t, _ := templates.ParseGlob(fmt.Sprintf("%s/templates/blogs/*.tmpl", basepath)); t != nil {
		templates = t
	}
	if t, _ := templates.ParseGlob(fmt.Sprintf("%s/templates/*.tmpl", basepath)); t != nil {
		templates = t
	}
	return &Template{
		templates: templates,
	}

}

func logmiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		log.Info().Msgf("Request to %s with ip %s", c.Path(), c.RealIP())
		if err := next(c); err != nil {
			c.Error(err)
		}
		return nil
	}
}

func initOtelTracer() {
	tracer, err := initTracer()
	if err != nil {
		log.Fatal().Msgf("failed to initialize tracer: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := tracer.Shutdown(ctx); err != nil {
			log.Fatal().Msgf("failed to shutdown tracer: %v", err)
		}
	}()
}
