package server

import (
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

	"github.com/andrewwillette/andrewwillettedotcom/aws"
	"github.com/andrewwillette/andrewwillettedotcom/config"
	"github.com/andrewwillette/andrewwillettedotcom/server/blog"
	"github.com/andrewwillette/andrewwillettedotcom/server/echopprof"
)

const (
	homeEndpoint = "/"

	musicEndpoint = "/music"

	resumeEndpoint = "/resume"

	sheetmusicEndpoint = "/sheet-music"

	blogsEndpoint   = "/blog"
	blogEndpoint    = "/blog/:blog"
	blogRssEndpoint = "/blog/rss"

	cssEndpoint = "/static/main.css"
	cssResource = "server/static/main.css"

	robotsEndpoint    = "/robots.txt"
	robotsTxtResource = "server/static/robots.txt"

	keyOfDayEndpoint = "/key-of-the-day"
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
	if config.C.PProfEnabled {
		echopprof.Wrap(e)
	}
	e.Renderer = getTemplateRenderer()
	blog.InitializeBlogs()
	aws.UpdateAudioCache()
	go aws.StartSQSPoller()
	if sslEnabled {
		e.Pre(middleware.HTTPSRedirect())
		e.AutoTLSManager.HostPolicy = autocert.HostWhitelist("andrewwillette.com")
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

// addRoutes add routes to the echo webserver
func addRoutes(e *echo.Echo) {
	e.GET(homeEndpoint, handleHomePage)
	e.GET(resumeEndpoint, handleResumePage)
	e.GET(musicEndpoint, handleRecordingsPage)
	e.GET(sheetmusicEndpoint, handleSheetmusicPage)
	e.GET(keyOfDayEndpoint, handleKeyOfDayPage)
	e.GET(blogsEndpoint, blog.HandleBlogPage)
	e.GET(blogRssEndpoint, blog.HandleRssFeed)
	e.GET(blogEndpoint, blog.HandleIndividualBlogPage)
	e.File(cssEndpoint, cssResource)
	e.File(robotsEndpoint, robotsTxtResource)
}

func addMiddleware(e *echo.Echo) {
	e.Use(logmiddleware)
}

type HomePageData struct {
	CurrentYear int
}

// handleHomePage handles returning the homepage template
func handleHomePage(c echo.Context) error {
	data := HomePageData{
		CurrentYear: time.Now().Year(),
	}
	err := c.Render(http.StatusOK, "homepage", data)
	if err != nil {
		return err
	}
	return nil
}

// handleResumePage handles returning the resume template
func handleResumePage(c echo.Context) error {
	err := c.Redirect(http.StatusPermanentRedirect, config.C.HomePageImageS3URL)
	if err != nil {
		return err
	}
	return nil
}

type KeyOfDayPage struct {
	KeyOfDay    string
	CurrentYear int
}

// handleKeyOfDayPage handles returning the key of the day
func handleKeyOfDayPage(c echo.Context) error {
	data := KeyOfDayPage{
		KeyOfDay:    key.TodaysKey(),
		CurrentYear: time.Now().Year(),
	}
	err := c.Render(http.StatusOK, "keyofdaypage", data)
	if err != nil {
		log.Error().Msgf("handleKeyOfDyPage error: %v", err)
		return err
	}
	return nil
}

// Template is the template renderer for my echo webserver
type Template struct {
	templates *template.Template
}

// Render renders the template
func (t *Template) Render(w io.Writer, name string, data any, c echo.Context) error {
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
		log.Info().Msgf("Request to registered path %s with ip %s", c.Path(), c.RealIP())
		if err := next(c); err != nil {
			c.Error(err)
		}
		return nil
	}
}
