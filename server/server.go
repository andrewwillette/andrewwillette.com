package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"

	"github.com/andrewwillette/keyOfDay/key"
	"github.com/andrewwillette/willette_api/logging"
	"github.com/labstack/echo/v4"
	"github.com/newrelic/go-agent/v3/newrelic"
)

const (
	getSoundcloudAllEndpoint     = "/get-soundcloud-urls"
	addSoundcloudEndpoint        = "/add-soundcloud-url"
	deleteSoundcloudEndpoint     = "/delete-soundcloud-url"
	keyOfDayEndpoint             = "/keyOfDay"
	healthEndpoint               = "/health"
	loginEndpoint                = "/login"
	updateSoundcloudUrlsEndpoint = "/update-soundcloud-urls"
	port                         = 80
)

func StartServer() {
	e := echo.New()
	e.GET("/", handleHomePage)
	e.GET("/resume", handleResumePage)
	e.GET("/music", handleMusicPage)
	e.GET("/kod", handleKeyOfDay)
	e.File("/static/main.css", "static/main.css")
	e.File("/static/illuminati.gif", "static/illuminati.gif")
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

type keyOfDayService interface {
	GetKeyOfDay() string
}

func getNewRelicApp() *newrelic.Application {
	newrelicLicense := os.Getenv("NEW_RELIC_LICENSE")
	app, err := newrelic.NewApplication(
		newrelic.ConfigAppName("go-andrewwillette"),
		newrelic.ConfigLicense(newrelicLicense),
	)
	if err != nil {
		logging.GlobalLogger.Error().Msgf("Failed to start new relic app, newrelic key: %s", newrelicLicense)
	}
	return app
}

// newServer setup server with endpoints
var buildTime = func() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.time" {
				return setting.Value
			}
		}
	}
	return ""
}()

var commit = func() string {
	cm := os.Getenv("GIT_COMMIT")
	if cm != "" {
		return cm
	} else {
		return "GIT_COMMIT not set"
	}
	// if info, ok := debug.ReadBuildInfo(); ok {
	// 	for _, setting := range info.Settings {
	// 		if setting.Key == "vcs.revision" {
	// 			return setting.Value
	// 		}
	// 	}
	// }
	// return ""
}()

type healthCheckResp struct {
	Buildtime string `json:"buildTime"`
	Commit    string `json:"gitCommit"`
}

func getKeyOfDay(c echo.Context) error {
	logging.GlobalLogger.Info().Msg("Calling key of day.")
	c.Response().Header().Set("Content-Type", "application-json")
	c.Response().Header().Set("Access-Control-Allow-Origin", "*")
	if err := json.NewEncoder(c.Response()).Encode(key.GetKeyOfDay()); err != nil {
		const errMsg = "Failed to encode keyOfDay string"
		logging.GlobalLogger.Err(err).Msg(errMsg)
		return c.String(http.StatusInternalServerError, errMsg)
	}
	return c.String(http.StatusOK, "")
}
