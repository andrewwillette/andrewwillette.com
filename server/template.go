package server

import (
	"html/template"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v4"
	zlog "github.com/rs/zerolog/log"
)

// renderer holds the parsed template map and implements echo.Renderer.
type renderer struct {
	templates map[string]*template.Template
}

// Render executes the named template using the shared "base" layout.
func (t *renderer) Render(w io.Writer, name string, data any, c echo.Context) error {
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

// newTemplateRenderer parses all page templates against the shared base layout
// and returns a renderer ready to assign to echo.Echo.Renderer.
func newTemplateRenderer() *renderer {
	base := template.Must(template.ParseFiles(
		filepath.Join(basepath, "templates/base.tmpl"),
	))

	pages := []string{
		"templates/homepage.tmpl",
		"templates/musicpage.tmpl",
		"templates/keyofdaypage.tmpl",
		"templates/sheetmusicpage.tmpl",
		"templates/blogs/blogspage.tmpl",
		"templates/blogs/singleblogpage.tmpl",
		"templates/chordcharts/chordchartspage.tmpl",
		"templates/chordcharts/singlechordchartpage.tmpl",
		"templates/adminpage.tmpl",
		"templates/showspage.tmpl",
	}

	templates := make(map[string]*template.Template, len(pages))
	for _, page := range pages {
		t := template.Must(base.Clone())
		template.Must(t.ParseFiles(filepath.Join(basepath, page)))
		name := strings.TrimSuffix(filepath.Base(page), ".tmpl")
		templates[name] = t
	}

	return &renderer{templates: templates}
}
