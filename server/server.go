package server

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/andrewwillette/keyofday/key"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	"golang.org/x/crypto/acme/autocert"

	"github.com/andrewwillette/andrewwillettedotcom/aws"
	"github.com/andrewwillette/andrewwillettedotcom/config"
	"github.com/andrewwillette/andrewwillettedotcom/server/blog"
	"github.com/andrewwillette/andrewwillettedotcom/server/echopprof"
	"github.com/andrewwillette/andrewwillettedotcom/server/traffic"
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

	adminEndpoint = "/admin/traffic"
)

var (
	_, b, _, _ = runtime.Caller(0)
	basepath   = filepath.Dir(b)
)

// StartServer start the server with https certificate configurable
func StartServer(sslEnabled bool) {
	if err := traffic.InitDB(config.C.TrafficDBPath); err != nil {
		zlog.Error().Err(err).Msg("failed to initialize traffic database")
	}
	e := echo.New()
	e.HideBanner = true
	e.Logger = newZerologAdapter(zlog.Logger)
	addRoutes(e)
	addMiddleware(e)
	if config.C.PProfEnabled {
		echopprof.Wrap(e)
	}
	e.Renderer = getTemplateRenderer()
	blog.InitializeBlogs()
	aws.UpdateAudioCache()
	aws.UpdateSheetMusicCache()
	go aws.UpdateAudioCacheOnPresignExpiry()
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
	e.GET(adminEndpoint, traffic.HandleAdminPage, traffic.BasicAuthMiddleware())
}

func addMiddleware(e *echo.Echo) {
	e.Use(logmiddleware)
	e.Use(traffic.TrackingMiddleware)
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
		zlog.Error().Msgf("handleKeyOfDyPage error: %v", err)
		return err
	}
	return nil
}

// Template is the template renderer for my echo webserver
type Template struct {
	templates map[string]*template.Template
}

// Render renders the template
func (t *Template) Render(w io.Writer, name string, data any, c echo.Context) error {
	tmpl, ok := t.templates[name]
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "template not found: "+name)
	}
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		zlog.Error().Err(err).Str("template", name).Msg("template execution failed")
		return err
	}
	return nil
}

// getTemplateRenderer returns a template renderer for my echo webserver
func getTemplateRenderer() *Template {
	// Parse base layout
	base := template.Must(template.ParseFiles(
		filepath.Join(basepath, "templates/base.tmpl"),
	))

	templates := make(map[string]*template.Template)

	// Pages that use the base layout
	pages := []string{
		"templates/homepage.tmpl",
		"templates/musicpage.tmpl",
		"templates/keyofdaypage.tmpl",
		"templates/sheetmusicpage.tmpl",
		"templates/blogs/blogspage.tmpl",
		"templates/blogs/singleblogpage.tmpl",
		"templates/adminpage.tmpl",
	}

	for _, page := range pages {
		// Clone base so each page gets its own copy
		pageTemplate := template.Must(base.Clone())
		template.Must(pageTemplate.ParseFiles(filepath.Join(basepath, page)))

		// Store with page name as key (e.g., "homepage")
		name := strings.TrimSuffix(filepath.Base(page), ".tmpl")
		templates[name] = pageTemplate
	}

	return &Template{templates: templates}
}

func logmiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		zlog.Info().Msgf("Request to registered path %s with ip %s", c.Path(), c.RealIP())
		if err := next(c); err != nil {
			c.Error(err)
		}
		return nil
	}
}

// zerologAdapter adapts zerolog to Echo's Logger interface
type zerologAdapter struct {
	logger zerolog.Logger
	prefix string
	level  log.Lvl
}

func newZerologAdapter(logger zerolog.Logger) *zerologAdapter {
	return &zerologAdapter{logger: logger, level: log.INFO}
}

func (z *zerologAdapter) Output() io.Writer                       { return z.logger }
func (z *zerologAdapter) SetOutput(w io.Writer)                   { z.logger = z.logger.Output(w) }
func (z *zerologAdapter) Prefix() string                          { return z.prefix }
func (z *zerologAdapter) SetPrefix(p string)                      { z.prefix = p }
func (z *zerologAdapter) Level() log.Lvl                          { return z.level }
func (z *zerologAdapter) SetLevel(l log.Lvl)                      { z.level = l }
func (z *zerologAdapter) SetHeader(h string)                      {}
func (z *zerologAdapter) Print(i ...interface{})                  { z.logger.Info().Msg(fmt.Sprint(i...)) }
func (z *zerologAdapter) Printf(format string, i ...interface{})  { z.logger.Info().Msgf(format, i...) }
func (z *zerologAdapter) Printj(j log.JSON)                       { z.logger.Info().Fields(j).Msg("") }
func (z *zerologAdapter) Debug(i ...interface{})                  { z.logger.Debug().Msg(fmt.Sprint(i...)) }
func (z *zerologAdapter) Debugf(format string, i ...interface{})  { z.logger.Debug().Msgf(format, i...) }
func (z *zerologAdapter) Debugj(j log.JSON)                       { z.logger.Debug().Fields(j).Msg("") }
func (z *zerologAdapter) Info(i ...interface{})                   { z.logger.Info().Msg(fmt.Sprint(i...)) }
func (z *zerologAdapter) Infof(format string, i ...interface{})   { z.logger.Info().Msgf(format, i...) }
func (z *zerologAdapter) Infoj(j log.JSON)                        { z.logger.Info().Fields(j).Msg("") }
func (z *zerologAdapter) Warn(i ...interface{})                   { z.logger.Warn().Msg(fmt.Sprint(i...)) }
func (z *zerologAdapter) Warnf(format string, i ...interface{})   { z.logger.Warn().Msgf(format, i...) }
func (z *zerologAdapter) Warnj(j log.JSON)                        { z.logger.Warn().Fields(j).Msg("") }
func (z *zerologAdapter) Error(i ...interface{})                  { z.logger.Error().Msg(fmt.Sprint(i...)) }
func (z *zerologAdapter) Errorf(format string, i ...interface{})  { z.logger.Error().Msgf(format, i...) }
func (z *zerologAdapter) Errorj(j log.JSON)                       { z.logger.Error().Fields(j).Msg("") }
func (z *zerologAdapter) Fatal(i ...interface{})                  { z.logger.Fatal().Msg(fmt.Sprint(i...)) }
func (z *zerologAdapter) Fatalf(format string, i ...interface{})  { z.logger.Fatal().Msgf(format, i...) }
func (z *zerologAdapter) Fatalj(j log.JSON)                       { z.logger.Fatal().Fields(j).Msg("") }
func (z *zerologAdapter) Panic(i ...interface{})                  { z.logger.Panic().Msg(fmt.Sprint(i...)) }
func (z *zerologAdapter) Panicf(format string, i ...interface{})  { z.logger.Panic().Msgf(format, i...) }
func (z *zerologAdapter) Panicj(j log.JSON)                       { z.logger.Panic().Fields(j).Msg("") }
