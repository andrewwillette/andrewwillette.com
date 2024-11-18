package server

import (
	"net/http"

	"github.com/andrewwillette/willette_api/server/aws"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

func handleRecordingsPage(c echo.Context) error {
	songs, err := aws.GetSongs()
	if err != nil {
		log.Error().Msgf("Unable to list songs: %v", err)
		return err
	}
	err = c.Render(http.StatusOK, "musicpage", songs)
	if err != nil {
		log.Error().Msgf("Unable to render musicpagenew: %v", err)
		return err
	}
	return nil
}
