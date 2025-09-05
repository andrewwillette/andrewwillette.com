package cmd

import (
	"github.com/andrewwillette/andrewwillettedotcom/aws"
	"github.com/andrewwillette/gofzf"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var deleteAudioCmd = &cobra.Command{
	Use:   "delete-audio",
	Short: "delete an audio file from S3",
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug().Msg("Deleting audio file from S3...")
		if err := deleteAudioFromS3(); err != nil {
			log.Fatal().Err(err).Msg("Failed to delete audio")
		}
		log.Debug().Msg("Audio deletion complete!")
	},
}

func init() {
	rootCmd.AddCommand(deleteAudioCmd)
}

func deleteAudioFromS3() error {
	songs, err := aws.GetAudioFromS3()
	if err != nil {
		return err
	}
	toselect := make([]string, len(songs))
	for i, song := range songs {
		toselect[i] = song.Name
	}
	selected, err := gofzf.Select(toselect)
	var selectedKey string
	for _, song := range songs {
		if song.Name == selected {
			selectedKey = song.Key
			break
		}
	}
	return aws.DeleteAudioFromS3(selectedKey)
}
