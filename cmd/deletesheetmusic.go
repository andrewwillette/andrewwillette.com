package cmd

import (
	"github.com/andrewwillette/andrewwillettedotcom/aws"
	"github.com/andrewwillette/gofzf"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var deleteSheetMusicCmd = &cobra.Command{
	Use:   "delete-sheet-music",
	Short: "delete sheet music entry in S3",
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug().Msg("Deleting sheet music from S3...")
		if err := deleteSheetMusicFromS3(); err != nil {
			log.Fatal().Err(err).Msg("Failed to delete sheet music")
		}
		log.Debug().Msg("SheetMusic deletion complete!")
	},
}

func init() {
	rootCmd.AddCommand(deleteSheetMusicCmd)
}

func deleteSheetMusicFromS3() error {
	sheetmusicBlobs, err := aws.ListSheetMusicObjects()
	if err != nil {
		return err
	}
	toselect := make([]string, len(sheetmusicBlobs))
	for i, song := range sheetmusicBlobs {
		toselect[i] = song.DisplayName
	}
	selected, err := gofzf.Select(toselect)
	var selectedKey string
	for _, sheetmusic := range sheetmusicBlobs {
		if sheetmusic.DisplayName == selected {
			selectedKey = sheetmusic.Key
			break
		}
	}
	return aws.DeleteSheetMusicFromS3(selectedKey)
}
