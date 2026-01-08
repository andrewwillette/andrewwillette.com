package cmd

import (
	"path/filepath"
	"strings"

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
	if err != nil {
		return err
	}
	var selectedKey string
	for _, song := range songs {
		if song.Name == selected {
			selectedKey = song.Key
			break
		}
	}

	// Delete audio file
	if err := aws.DeleteAudioFromS3(selectedKey); err != nil {
		return err
	}

	// Delete corresponding image (replace .wav/.mp3 with .png)
	ext := filepath.Ext(selectedKey)
	imageKey := strings.TrimSuffix(selectedKey, ext) + ".png"
	log.Info().Msgf("Deleting corresponding image %s...", imageKey)
	if err := aws.DeleteAudioFromS3(imageKey); err != nil {
		log.Warn().Msgf("Could not delete image (may not exist): %v", err)
	}

	return nil
}
