package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/andrewwillette/andrewwillettedotcom/aws"
	"github.com/andrewwillette/andrewwillettedotcom/images"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var uploadImages bool
var keepImages bool

var generateImagesCmd = &cobra.Command{
	Use:   "generate-images",
	Short: "Generate cover art images for all audio files",
	Long:  `Generates per-song cover art images by overlaying song titles on the base album image. Use --upload to also upload to S3.`,
	Run: func(cmd *cobra.Command, args []string) {
		keys, err := aws.GetAudioKeysFromS3()
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to fetch audio keys from S3")
		}

		if len(keys) == 0 {
			log.Info().Msg("No audio files found in S3")
			return
		}

		log.Info().Msgf("Found %d audio files in S3", len(keys))

		if err := images.GenerateSongCoverArt(keys); err != nil {
			log.Fatal().Err(err).Msg("Failed to generate images")
		}

		if uploadImages {
			if err := uploadGeneratedImages(); err != nil {
				log.Fatal().Err(err).Msg("Failed to upload images to S3")
			}
			if !keepImages {
				if err := cleanupGeneratedImages(); err != nil {
					log.Warn().Msgf("Could not clean up generated images: %v", err)
				}
			}
		}

		log.Info().Msg("Image generation complete!")
	},
}

func init() {
	generateImagesCmd.Flags().BoolVarP(&uploadImages, "upload", "u", false, "Upload generated images to S3")
	generateImagesCmd.Flags().BoolVarP(&keepImages, "keep", "k", false, "Keep generated images locally after upload")
	rootCmd.AddCommand(generateImagesCmd)
}

func uploadGeneratedImages() error {
	imagesDir := "images/audioimages"
	entries, err := os.ReadDir(imagesDir)
	if err != nil {
		return fmt.Errorf("failed to read images directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".png" {
			continue
		}

		imagePath := filepath.Join(imagesDir, entry.Name())
		if err := aws.UploadAudioImageToS3(imagePath); err != nil {
			return fmt.Errorf("failed to upload %s: %w", entry.Name(), err)
		}
	}

	return nil
}

func cleanupGeneratedImages() error {
	imagesDir := "images/audioimages"
	entries, err := os.ReadDir(imagesDir)
	if err != nil {
		return fmt.Errorf("failed to read images directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".png" {
			continue
		}

		imagePath := filepath.Join(imagesDir, entry.Name())
		if err := os.Remove(imagePath); err != nil {
			log.Warn().Msgf("Could not delete %s: %v", imagePath, err)
		}
	}

	return nil
}
