package main

import (
	"os"

	_ "github.com/andrewwillette/willette_api/logging"
	"github.com/andrewwillette/willette_api/server"
	"github.com/rs/zerolog/log"
)

func main() {
	log.Info().Msg("starting andrewillette.com server")
	env := os.Getenv("ENV")
	var sslEnabled bool
	if env == "PROD" {
		sslEnabled = true
	} else {
		sslEnabled = false
	}
	server.StartServer(sslEnabled)
}
