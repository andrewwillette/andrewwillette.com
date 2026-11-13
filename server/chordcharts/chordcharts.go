package chordcharts

import (
	"html/template"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/andrewwillette/andrewwillettedotcom/aws"
	"github.com/labstack/echo/v4"
)

type ChordChart struct {
	Title         string
	Artist        string
	DefaultKey    string
	TimeSignature string
	Notes         string
	URLVal        string
	FileName      string
	Body          string
	BodyHTML      template.HTML
	CurrentYear   int
}

type ChordChartsPageData struct {
	ChordCharts []ChordChart
	CurrentYear int
}

// InitializeChordCharts (re)loads the chord chart cache from S3. Call it at
// startup, and again any time S3 content changes (e.g. after a sync), so the
// site reflects new charts without a redeploy or restart.
func InitializeChordCharts() {
	aws.UpdateChordChartsCache()
}

func chordChartFromJSON(j aws.ChordChartJSONObject) ChordChart {
	return ChordChart{
		Title:         j.Title,
		Artist:        j.Artist,
		DefaultKey:    j.DefaultKey,
		TimeSignature: j.TimeSignature,
		Notes:         j.Notes,
		URLVal:        j.URLVal,
		FileName:      j.FileName,
		Body:          j.Body,
	}
}

// GetChordChart returns the chord chart matching urlval, rendered to HTML.
func GetChordChart(urlval string) ChordChart {
	for _, j := range aws.GetCachedChordCharts() {
		if j.URLVal == urlval {
			c := chordChartFromJSON(j)
			c.BodyHTML = RenderBody(c.Body)
			c.CurrentYear = time.Now().Year()
			return c
		}
	}
	return ChordChart{}
}

// GetChordChartsPageData returns chord chart data for rendering in the list template.
func GetChordChartsPageData() ChordChartsPageData {
	cached := aws.GetCachedChordCharts()
	charts := make([]ChordChart, 0, len(cached))
	for _, j := range cached {
		charts = append(charts, chordChartFromJSON(j))
	}
	sort.Slice(charts, func(i, k int) bool {
		return strings.ToLower(charts[i].Title) < strings.ToLower(charts[k].Title)
	})
	return ChordChartsPageData{
		ChordCharts: charts,
		CurrentYear: time.Now().Year(),
	}
}

// HandleChordChartsPage handles returning the chord charts list page
func HandleChordChartsPage(c echo.Context) error {
	return c.Render(http.StatusOK, "chordchartspage", GetChordChartsPageData())
}

// HandleChordChartPage handles returning an individual chord chart page
func HandleChordChartPage(c echo.Context) error {
	return c.Render(http.StatusOK, "singlechordchartpage", GetChordChart(c.Param("chart")))
}
