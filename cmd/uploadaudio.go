package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/andrewwillette/andrewwillettedotcom/aws"
	"github.com/andrewwillette/andrewwillettedotcom/images"
	"github.com/andrewwillette/gofzf"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var audioFilePath string
var audioFileDir string
var audioResultDir string

var uploadAudioCmd = &cobra.Command{
	Use:   "upload-audio",
	Short: "Upload audio file to S3",
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug().Msg("Uploading audio file to S3...")
		if audioFileDir != "" {
			if err := uploadAudioFileFromDir(audioFileDir); err != nil {
				log.Fatal().Err(err).Msg("Failed to upload audio from directory")
			}
		} else if audioFilePath != "" {
			if err := uploadAudioWithImage(audioFilePath); err != nil {
				log.Fatal().Err(err).Msg("Failed to upload audio")
			}
		} else {
			_ = cmd.Help()
			os.Exit(1)
		}
	},
}

func init() {
	uploadAudioCmd.Flags().StringVarP(&audioFilePath, "file", "f", "", "Path to audio file")
	uploadAudioCmd.Flags().StringVarP(&audioFileDir, "dir", "d", "", "Directory that contains audio file")
	uploadAudioCmd.Flags().StringVarP(&audioResultDir, "result-dir", "r", "", "Directory to move the uploaded file into (with timestamp suffix)")
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
	return uploadAudioWithImage(fileMap[selected])
}

// moveToResultDir moves src into destDir, inserting a timestamp before the extension.
// e.g. song.mp3 → destDir/song_20260625_153045.mp3
func moveToResultDir(src, destDir string) error {
	if info, err := os.Stat(destDir); err != nil || !info.IsDir() {
		return fmt.Errorf("result-dir %q is not a valid directory", destDir)
	}
	ext := filepath.Ext(src)
	base := strings.TrimSuffix(filepath.Base(src), ext)
	ts := time.Now().Format("20060102_150405")
	dest := filepath.Join(destDir, fmt.Sprintf("%s_%s%s", base, ts, ext))
	if err := os.Rename(src, dest); err != nil {
		return fmt.Errorf("failed to move %s to %s: %w", src, dest, err)
	}
	log.Info().Msgf("Moved %s → %s", src, dest)
	return nil
}

func uploadAudioToS3(audioFile string) error {
	log.Info().Msgf("Uploading audio file %s to S3...", audioFile)
	if !isValidAudioFile(audioFile) {
		return fmt.Errorf("invalid audio file: %s", audioFile)
	}
	return aws.UploadAudioToS3(audioFile)
}

func uploadAudioWithImage(audioFile string) error {
	if err := uploadAudioToS3(audioFile); err != nil {
		return err
	}

	log.Info().Msg("Generating cover art image...")
	imagePath, err := images.GenerateSingleCoverArt(audioFile)
	if err != nil {
		return fmt.Errorf("failed to generate cover art: %w", err)
	}

	log.Info().Msgf("Uploading cover art image %s to S3...", imagePath)
	if err := aws.UploadAudioImageToS3(imagePath); err != nil {
		return fmt.Errorf("failed to upload cover art: %w", err)
	}

	if err := os.Remove(imagePath); err != nil {
		log.Warn().Msgf("Could not delete generated image: %v", err)
	}

	log.Info().Msg("Audio and cover art uploaded successfully!")

	if audioResultDir != "" {
		if err := moveToResultDir(audioFile, audioResultDir); err != nil {
			return err
		}
	}

	return nil
}

func isValidAudioFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return false
	}
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".wav" || ext == ".mp3"
}
