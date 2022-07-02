package main

import (
	"github.com/andrewwillette/willette_api/logging"
	"github.com/andrewwillette/willette_api/server"
)

func main() {
	logging.GlobalLogger.Info().Msg("Starting application.")
	server.StartServer()
}
