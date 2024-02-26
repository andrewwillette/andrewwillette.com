package server

import (
	"context"
	"embed"
	_ "embed"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/andrewwillette/keyofday/key"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/crypto/acme/autocert"
)

const (
	homeEndpoint       = "/"
	musicEndpoint      = "/music"
	resumeEndpoint     = "/resume"
	sheetmusicEndpoint = "/sheet-music"
	cssEndpoint        = "/static/main.css"
	cssResource        = "static/main.css"
	keyOfDayEndpoint   = "/kod"
	resumeResource     = "https://andrewwillette.s3.us-east-2.amazonaws.com/newdir/resume.pdf"
)

var (
	_, b, _, _ = runtime.Caller(0)
	basepath   = filepath.Dir(b)
)

// StartServer start the server with https certificate configurable
func StartServer(isHttps bool) {
	e := echo.New()
	e.Renderer = getTemplateRenderer()
	addRoutes(e)
	if isHttps {
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
		server := wrapRouter(e)
		lambda.Start(server)
		// e.Logger.Fatal(e.Start(":80"))
	}
}

func wrapRouter(e *echo.Echo) func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		body := strings.NewReader(request.Body)
		req := httptest.NewRequest(request.HTTPMethod, request.Path, body)
		for k, v := range request.Headers {
			req.Header.Add(k, v)
		}

		q := req.URL.Query()
		for k, v := range request.QueryStringParameters {
			q.Add(k, v)
		}
		req.URL.RawQuery = q.Encode()

		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		res := rec.Result()
		responseBody, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return formatAPIErrorResponse(http.StatusInternalServerError, res.Header, err.Error())
		}

		return formatAPIResponse(res.StatusCode, res.Header, string(responseBody))
	}
}
func formatAPIResponse(statusCode int, headers http.Header, responseData string) (events.APIGatewayProxyResponse, error) {
	responseHeaders := make(map[string]string)

	responseHeaders["Content-Type"] = "application/json"
	for key, value := range headers {
		responseHeaders[key] = ""

		if len(value) > 0 {
			responseHeaders[key] = value[0]
		}
	}

	responseHeaders["Access-Control-Allow-Origin"] = "*"
	responseHeaders["Access-Control-Allow-Headers"] = "origin,Accept,Authorization,Content-Type"

	return events.APIGatewayProxyResponse{
		Body:       responseData,
		Headers:    responseHeaders,
		StatusCode: statusCode,
	}, nil
}

func formatAPIErrorResponse(statusCode int, headers http.Header, err string) (events.APIGatewayProxyResponse, error) {
	responseHeaders := make(map[string]string)

	responseHeaders["Content-Type"] = "application/json"
	for key, value := range headers {
		responseHeaders[key] = ""

		if len(value) > 0 {
			responseHeaders[key] = value[0]
		}
	}

	responseHeaders["Access-Control-Allow-Origin"] = "*"
	responseHeaders["Access-Control-Allow-Headers"] = "origin,Accept,Authorization,Content-Type"

	return events.APIGatewayProxyResponse{
		Body:       err,
		Headers:    responseHeaders,
		StatusCode: statusCode,
	}, nil
}

// addRoutes adds routes to the echo webserver
func addRoutes(e *echo.Echo) {
	e.GET(homeEndpoint, handleHomePage)
	e.GET(resumeEndpoint, handleResumePage)
	e.GET(musicEndpoint, handleMusicPage)
	e.GET(sheetmusicEndpoint, handleSheetmusicPage)
	e.GET(keyOfDayEndpoint, handleKeyOfDayPage)
	e.GET(cssEndpoint, contentHandler)
}

//go:embed static/*
var content embed.FS

var contentHandler = echo.WrapHandler(http.FileServer(http.FS(content)))

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

// handleSheetmusicPage handles returning the transcription template
func handleSheetmusicPage(c echo.Context) error {
	sort.Sort(sheetmusicData.Sheets)
	err := c.Render(http.StatusOK, "sheetmusicpage", sheetmusicData)
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

type DropboxReference struct {
	Name string
	URL  string
}

type Sheets []DropboxReference

type SheetMusicPageData struct {
	Sheets Sheets
}

//go:embed templates/*.tmpl
var templates embed.FS

// getTemplateRenderer returns a template renderer for my echo webserver
func getTemplateRenderer() *Template {
	t := &Template{
		// templates: template.Must(template.ParseGlob(fmt.Sprintf("%s/templates/*.tmpl", basepath))),
		// retrieve templates from embedded filesystem
		template.Must(template.ParseFS(templates, "templates/*.tmpl")),
	}
	return t
}

// Len to implement sort.Interface
func (sheets Sheets) Len() int {
	return len(sheets)
}

// Swap to implement sort.Interface
func (sheets Sheets) Swap(i, j int) {
	sheets[i], sheets[j] = sheets[j], sheets[i]
}

// Less to implement sort.Interface
func (sheets Sheets) Less(i, j int) bool {
	switch strings.Compare(sheets[i].Name, sheets[j].Name) {
	case -1:
		return true
	case 0:
		return false
	case 1:
		return false
	default:
		return false
	}
}
