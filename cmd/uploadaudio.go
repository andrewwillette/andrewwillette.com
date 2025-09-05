package cmd

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var audioFilePath string

var uploadAudioCmd = &cobra.Command{
	Use:   "upload-audio",
	Short: "Upload audio file to S3",
	Run: func(cmd *cobra.Command, args []string) {
		log.Info().Msg("Uploading audio file to S3...")
		if audioFilePath == "" {
			log.Fatal().Msg("Please provide a path using --file")
		}
		if err := uploadAudioToS3(audioFilePath); err != nil {
			log.Fatal().Err(err).Msg("Failed to upload audio")
		}
		log.Info().Msg("Upload complete!")
	},
}

func init() {
	uploadAudioCmd.Flags().StringVarP(&audioFilePath, "file", "f", "", "Path to audio file")
	rootCmd.AddCommand(uploadAudioCmd)
}

func uploadAudioToS3(filePath string) error {
	// TODO: your actual S3 upload logic goes here
	fmt.Println("Uploading file:", filePath)
	return nil
}
