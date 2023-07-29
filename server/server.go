package server

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"runtime"

	"github.com/andrewwillette/keyofday/key"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/crypto/acme/autocert"
)

const (
	homeEndpoint     = "/"
	musicEndpoint    = "/music"
	resumeEndpoint   = "/resume"
	cssEndpoint      = "/static/main.css"
	cssResource      = "static/main.css"
	keyOfDayEndpoint = "/kod"
	resumeResource   = "https://andrewwillette.s3.us-east-2.amazonaws.com/newdir/resume.pdf"
)

// StartHttpServer starts the web server with http
func StartHttpServer() {
	e := echo.New()
	addRoutes(e)
	e.Logger.Fatal(e.Start(":80"))
}

// StartHttpsServer starts the web server with https certificate
// provided by letsencrypt
func StartHttpsServer() {
	e := echo.New()
	e.Pre(middleware.HTTPSRedirect())
	e.AutoTLSManager.HostPolicy = autocert.HostWhitelist("andrewwillette.com")
	e.AutoTLSManager.Cache = autocert.DirCache("/var/www/.cache")
	addRoutes(e)
	go func(c *echo.Echo) {
		e.Logger.Fatal(e.Start(":80"))
	}(e)
	e.Logger.Fatal(e.StartAutoTLS(":443"))
}

// addRoutes adds routes to the echo webserver
func addRoutes(e *echo.Echo) {
	e.GET(homeEndpoint, handleHomePage)
	e.GET(resumeEndpoint, handleResumePage)
	e.GET(musicEndpoint, handleMusicPage)
	e.GET(keyOfDayEndpoint, handleKeyOfDayPage)
	e.File(cssEndpoint, cssResource)
	e.Renderer = getTemplateRenderer()
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
		log.Println(err)
		return err
	}
	return nil
}

// handleMusicPage handles returning the music template
func handleMusicPage(c echo.Context) error {
	err := c.Render(http.StatusOK, "musicpage", musicData)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// handleKeyOfDayPage handles returning the key of the day
func handleKeyOfDayPage(c echo.Context) error {
	err := c.Render(http.StatusOK, "keyofdaypage", key.GetKeyOfDay())
	if err != nil {
		log.Println(err)
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

type Song struct {
	Title string
	URL   string
}
type MusicPageData struct {
	Songs []Song
}

var (
	_, b, _, _ = runtime.Caller(0)
	basepath   = filepath.Dir(b)
	musicData  = MusicPageData{
		Songs: []Song{
			{
				Title: "Swallowtail Jig",
				URL:   "https://w.soundcloud.com/player/?url=https%3A//api.soundcloud.com/tracks/1511471356&color=%23b0a472&auto_play=false&hide_related=true&show_comments=false&show_user=true&show_reposts=false&show_teaser=false&visual=true",
			},
			{
				Title: "Sherry",
				URL:   "https://w.soundcloud.com/player/?url=https%3A//api.soundcloud.com/tracks/1412600335&color=%233799bb&auto_play=false&hide_related=true&show_comments=false&show_user=true&show_reposts=false&show_teaser=false&visual=true",
			},
			{
				Title: "Raggedy Ann",
				URL:   "https://w.soundcloud.com/player/?url=https%3A//api.soundcloud.com/tracks/1353192850&color=%23e25862&auto_play=false&hide_related=true&show_comments=false&show_user=true&show_reposts=false&show_teaser=false&visual=true",
			},
			{
				Title: "Leather Britches",
				URL:   "https://w.soundcloud.com/player/?url=https%3A//api.soundcloud.com/tracks/1340417752&color=%237c643d&auto_play=false&hide_related=true&show_comments=false&show_user=true&show_reposts=false&show_teaser=false&visual=true",
			},
			{
				Title: "Carrol County Blues",
				URL:   "https://w.soundcloud.com/player/?url=https%3A//api.soundcloud.com/tracks/1299581302&color=%23e25862&auto_play=false&hide_related=true&show_comments=false&show_user=true&show_reposts=false&show_teaser=false&visual=true",
			},
			{
				Title: "Flop Eared Mule",
				URL:   "https://w.soundcloud.com/player/?url=https%3A//api.soundcloud.com/tracks/1279484290&color=%2356bb37&auto_play=false&hide_related=true&show_comments=false&show_user=true&show_reposts=false&show_teaser=false&visual=true",
			},
			{
				Title: "Dixie",
				URL:   "https://w.soundcloud.com/player/?url=https%3A//api.soundcloud.com/tracks/1279435306&color=%23e193db&auto_play=false&hide_related=true&show_comments=false&show_user=true&show_reposts=false&show_teaser=false&visual=true",
			},
			{
				Title: "Polecat Blues",
				URL:   "https://w.soundcloud.com/player/?url=https%3A//api.soundcloud.com/tracks/1244352139&color=%23f3143a&auto_play=false&hide_related=true&show_comments=false&show_user=true&show_reposts=false&show_teaser=false&visual=true",
			},
			{
				Title: "Wednesday Night Waltz",
				URL:   "https://w.soundcloud.com/player/?url=https%3A//api.soundcloud.com/tracks/1241630236&color=%23e6d5d5&auto_play=false&hide_related=true&show_comments=false&show_user=true&show_reposts=false&show_teaser=false&visual=true",
			},
			{
				Title: "Down The Road Somewhere",
				URL:   "https://w.soundcloud.com/player/?url=https%3A//api.soundcloud.com/tracks/1206283588&color=%235a1c1c&auto_play=false&hide_related=true&show_comments=false&show_user=true&show_reposts=false&show_teaser=false&visual=true",
			},
			{
				Title: "Benton's Dream",
				URL:   "https://w.soundcloud.com/player/?url=https%3A//api.soundcloud.com/tracks/1202201206&color=%23ccd7c8&auto_play=false&hide_related=true&show_comments=false&show_user=true&show_reposts=false&show_teaser=false&visual=true",
			},
			{
				Title: "Sugar In The Gourd",
				URL:   "https://w.soundcloud.com/player/?url=https%3A//api.soundcloud.com/tracks/1177441174&color=%23c224c3&auto_play=false&hide_related=true&show_comments=false&show_user=true&show_reposts=false&show_teaser=false&visual=true",
			},
			{
				Title: "Walking In My Sleep",
				URL:   "https://w.soundcloud.com/player/?url=https%3A//api.soundcloud.com/tracks/1144703368&color=%23b0a472&auto_play=false&hide_related=true&show_comments=false&show_user=true&show_reposts=false&show_teaser=false&visual=true",
			},
		},
	}
)

// getTemplateRenderer returns a template renderer for my echo webserver
func getTemplateRenderer() *Template {
	t := &Template{
		templates: template.Must(template.ParseGlob(fmt.Sprintf("%s/templates/*.tmpl", basepath))),
	}
	return t
}
