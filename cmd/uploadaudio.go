package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/andrewwillette/andrewwillettedotcom/aws"
	"github.com/andrewwillette/gofzf"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var audioFilePath string
var audioFileDir string

var uploadAudioCmd = &cobra.Command{
	Use:   "upload-audio",
	Short: "Upload audio file to S3",
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug().Msg("Uploading audio file to S3...")
		if audioFileDir != "" {
			if err := uploadAudioFileFromDir(audioFileDir); err != nil {
				log.Fatal().Err(err).Msg("Failed to upload audio from directory")
			}
		} else if audioFilePath != "" { // other
			if err := uploadAudioToS3(audioFilePath); err != nil {
				log.Fatal().Err(err).Msg("Failed to upload audio")
			}
			log.Info().Msg("Upload complete!")
		} else {
			_ = cmd.Help()
			os.Exit(1)
		}
	},
}

func init() {
	uploadAudioCmd.Flags().StringVarP(&audioFilePath, "file", "f", "", "Path to audio file")
	uploadAudioCmd.Flags().StringVarP(&audioFileDir, "dir", "d", "", "Directory that contains audio file")
	rootCmd.AddCommand(uploadAudioCmd)
}

func uploadAudioFileFromDir(dir string) error {
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("invalid directory: %s", dir)
	}
	files, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}
	fileList := []string{}
	fileMap := make(map[string]string) // basename -> full path
	for _, file := range files {
		if !file.IsDir() {
			fullPath := filepath.Join(dir, file.Name())
			if isValidAudioFile(fullPath) {
				fileList = append(fileList, file.Name())
				fileMap[file.Name()] = fullPath
			}
		}
	}
	selected, err := gofzf.Select(fileList)
	if err != nil {
		return fmt.Errorf("failed to select file: %w", err)
	}
	return aws.UploadAudioToS3(fileMap[selected])
}

func uploadAudioToS3(audioFile string) error {
	log.Info().Msgf("Uploading audio file %s to S3...", audioFile)
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
