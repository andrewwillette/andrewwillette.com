package server

import (
	"net/http"
	"time"

	"github.com/andrewwillette/andrewwillettedotcom/aws"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

type AudioPageData struct {
	Songs       []aws.S3Song
	CurrentYear int
}

func handleRecordingsPage(c echo.Context) error {
	songs, err := aws.GetCachedAudio()
	if err != nil {
		log.Error().Msgf("Unable to list songs: %v", err)
		return err
	}
	data := AudioPageData{
		Songs:       songs,
		CurrentYear: time.Now().Year(),
	}
	err = c.Render(http.StatusOK, "musicpage", data)
	if err != nil {
		log.Error().Msgf("Unable to render musicpagenew: %v", err)
		return err
	}
	return nil
}
