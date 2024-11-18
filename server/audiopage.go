package server

import (
	"net/http"
	"time"

	"github.com/andrewwillette/willette_api/server/aws"
	// "github.com/aws/aws-sdk-go/aws"
	// "github.com/aws/aws-sdk-go/aws/session"
	// "github.com/aws/aws-sdk-go/service/s3"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

func handleRecordingsPage(c echo.Context) error {
	start := time.Now()
	songs, err := aws.ListSongsWithRandomImage()
	log.Debug().Msgf("listSongsWithRandomImage took %v", time.Since(start))
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
