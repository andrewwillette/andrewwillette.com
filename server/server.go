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
)

const (
	homeEndpoint                 = "/"
	musicEndpoint                = "/music"
	resumeEndpoint               = "/resume"
	cssEndpoint                  = "/static/main.css"
	keyOfDayEndpoint             = "/kod"
	updateSoundcloudUrlsEndpoint = "/update-soundcloud-urls"
	port                         = 80
)

func StartServer() {
	e := echo.New()
	e.GET(homeEndpoint, handleHomePage)
	e.GET(resumeEndpoint, handleResumePage)
	e.GET(musicEndpoint, handleMusicPage)
	e.GET(keyOfDayEndpoint, handleKeyOfDay)
	e.File(cssEndpoint, "static/main.css")
	e.Renderer = getTemplateRenderer()
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", port)))
}

func handleHomePage(c echo.Context) error {
	err := c.Render(http.StatusOK, "homepage", nil)
	if err != nil {
		return err
	}
	return nil
}

func handleResumePage(c echo.Context) error {
	err := c.Redirect(http.StatusPermanentRedirect, "https://andrewwillette.s3.us-east-2.amazonaws.com/newdir/resume.pdf")
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func handleMusicPage(c echo.Context) error {
	err := c.Render(http.StatusOK, "musicpage", nil)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func handleKeyOfDay(c echo.Context) error {
	err := c.Render(http.StatusOK, "keyofdaypage", key.GetKeyOfDay())
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

var (
	_, b, _, _ = runtime.Caller(0)
	basepath   = filepath.Dir(b)
)

func getTemplateRenderer() *Template {
	t := &Template{
		templates: template.Must(template.ParseGlob(fmt.Sprintf("%s/templates/*.tmpl", basepath))),
	}
	return t
}
