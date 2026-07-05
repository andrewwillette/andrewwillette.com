package cmd

import (
	"github.com/andrewwillette/andrewwillettedotcom/aws"
	"github.com/andrewwillette/gofzf"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var deleteShowCmd = &cobra.Command{
	Use:   "delete-show",
	Short: "Delete a show entry from S3",
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug().Msg("Deleting show from S3...")
		if err := deleteShowFromS3(); err != nil {
			log.Fatal().Err(err).Msg("Failed to delete show")
		}
		log.Debug().Msg("Show deletion complete!")
	},
}

func init() {
	rootCmd.AddCommand(deleteShowCmd)
}

func deleteShowFromS3() error {
	shows, err := aws.ListShowObjects()
	if err != nil {
		return err
	}
	titles := make([]string, len(shows))
	for i, s := range shows {
		titles[i] = s.Title
	}
	selected, err := gofzf.Select(titles)
	if err != nil {
		return err
	}
	var selectedKey string
	for _, s := range shows {
		if s.Title == selected {
			selectedKey = s.Key
			break
		}
	}
	return aws.DeleteShowFromS3(selectedKey)
}
