package main

import (
	"os"

	_ "github.com/andrewwillette/willette_api/logging"
	"github.com/andrewwillette/willette_api/server"
	"github.com/rs/zerolog/log"
)

func main() {
	log.Info().Msg("Starting server")
	env := os.Getenv("ENV")
	if env == "PROD" {
		sslEnabled := true
		server.StartServer(sslEnabled)
	} else {
		sslEnabled := false
		server.StartServer(sslEnabled)
	}
}
