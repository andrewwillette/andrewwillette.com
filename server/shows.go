package server

import (
	"net/http"
	"time"

	"github.com/andrewwillette/andrewwillettedotcom/aws"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

type ShowsPageData struct {
	Shows       []aws.ShowJSONObject
	CurrentYear int
}

func handleShowsPage(c echo.Context) error {
	shows, err := aws.GetCachedShows()
	if err != nil {
		log.Error().Msgf("Unable to get shows: %v", err)
		return err
	}
	data := ShowsPageData{
		Shows:       shows,
		CurrentYear: time.Now().Year(),
	}
	err = c.Render(http.StatusOK, "showspage", data)
	if err != nil {
		log.Error().Msgf("Unable to render showspage: %v", err)
		return err
	}
	return nil
}
