package main

import (
	"os"

	_ "github.com/andrewwillette/willette_api/logging"
	"github.com/andrewwillette/willette_api/server"
)

func main() {
	env := os.Getenv("ENV")
	if env == "PROD" {
		sslEnabled := true
		server.StartServer(sslEnabled)
	} else {
		sslEnabled := false
		server.StartServer(sslEnabled)
	}
}
