package server

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"runtime"

	"github.com/andrewwillette/keyOfDay/key"
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
	e.GET(keyOfDayEndpoint, handleKeyOfDay)
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
	err := c.Render(http.StatusOK, "musicpage", nil)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// handleKeyOfDay handles returning the key of the day
func handleKeyOfDay(c echo.Context) error {
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

var (
	_, b, _, _ = runtime.Caller(0)
	basepath   = filepath.Dir(b)
)

// getTemplateRenderer returns a template renderer for my echo webserver
func getTemplateRenderer() *Template {
	t := &Template{
		templates: template.Must(template.ParseGlob(fmt.Sprintf("%s/templates/*.tmpl", basepath))),
	}
	return t
}
