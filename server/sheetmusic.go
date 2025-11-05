package server

import (
	"net/http"
	"time"

	awslib "github.com/andrewwillette/andrewwillettedotcom/aws"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

type SheetMusicPageData struct {
	Sheets      []awslib.SheetMusicJSONObject
	CurrentYear int
}

func handleSheetmusicPage(c echo.Context) error {
	sheets, err := awslib.GetCachedSheetMusic()
	if err != nil {
		log.Error().Err(err).Msg("Unable to get cached sheet music")
		return err
	}
	log.Info().Msgf("Serving sheet music page with %d entries", len(sheets))

	data := SheetMusicPageData{
		Sheets:      sheets,
		CurrentYear: time.Now().Year(),
	}
	return c.Render(http.StatusOK, "sheetmusicpage", data)
}
