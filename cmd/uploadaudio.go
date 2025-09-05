package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/andrewwillette/andrewwillettedotcom/server/aws"

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

func uploadAudioToS3(audioFile string) error {
	if !isValidAudioFile(audioFile) {
		return fmt.Errorf("invalid audio file: %s", audioFile)
	}
	return aws.UploadAudioToS3(audioFile)
}

func isValidAudioFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return false
	}
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".wav" || ext == ".mp3"
}
