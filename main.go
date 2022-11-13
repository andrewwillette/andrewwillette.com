package main

import (
	"os"

	"github.com/andrewwillette/willette_api/logging"
	"github.com/andrewwillette/willette_api/server"
)

func main() {
	logging.GlobalLogger.Info().Msg("Starting application.")
	env := os.Getenv("ENV")
	if env == "PROD" {
		server.StartHttpsServer()
	} else {
		server.StartHttpServer()
	}
}
