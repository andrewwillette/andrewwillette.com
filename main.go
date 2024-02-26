package main

import (
	"os"

	"github.com/andrewwillette/willette_api/server"
)

func main() {
	// logging.GlobalLogger.Info().Msg("Starting application.")
	println("Starting application.")
	env := os.Getenv("ENV")
	if env == "PROD" {
		sslEnabled := true
		server.StartServer(sslEnabled)
	} else {
		sslEnabled := false
		server.StartServer(sslEnabled)
	}
}
