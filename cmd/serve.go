package cmd

import (
	"os"

	"github.com/andrewwillette/andrewwillettedotcom/server"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the web server",
	Run: func(cmd *cobra.Command, args []string) {
		log.Info().Msg("starting andrewwillette.com server")
		env := os.Getenv("ENV")
		sslEnabled := env == "PROD"
		server.StartServer(sslEnabled)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
