package images

import (
	"fmt"
	"image"
	"image/png"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/fogleman/gg"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// getBaseImage returns S3 image URL for the base album art
func getBaseImage() string {
	return "https://andrewwillette.s3.us-east-2.amazonaws.com/audio/webpage_album_cuts_image.png"
}

// downloadImage fetches an image from a URL and returns it
func downloadImage(url string) (image.Image, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	return img, nil
}

// formatForDisplay converts snake_case filename to Title Case display name
func formatForDisplay(snakeCase string) string {
	name := strings.ReplaceAll(snakeCase, "_", " ")
	titleCaser := cases.Title(language.English)
	return titleCaser.String(name)
}

// GenerateSongCoverArt generates per-song cover art images.
// songKeys should be S3 keys like "audio/billy_in_the_lowground.wav".
// Images are saved to images/audioimages/ directory.
func GenerateSongCoverArt(songKeys []string) error {
	baseImageURL := getBaseImage()

	fmt.Println("Downloading base image...")
	baseImage, err := downloadImage(baseImageURL)
	if err != nil {
		return err
	}

	bounds := baseImage.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	fmt.Printf("Base image size: %dx%d\n", width, height)

	if err := os.MkdirAll("images/audioimages", 0755); err != nil {
		return fmt.Errorf("failed to create images directory: %w", err)
	}

	for _, key := range songKeys {
		// Extract base filename without extension
		base := filepath.Base(key)
		nameWithoutExt := strings.TrimSuffix(base, filepath.Ext(base))

		// Display title for overlay (e.g., "Billy In The Lowground")
		displayTitle := formatForDisplay(nameWithoutExt)

		// Output filename stays as snake_case to match audio files
		outputFilename := nameWithoutExt + ".png"

		fmt.Printf("Generating album art for: %s\n", displayTitle)

		// Create a new context with the base image
		dc := gg.NewContextForImage(baseImage)

		// Configure text style
		fontSize := float64(width) / 22
		err := dc.LoadFontFace(
			"/System/Library/Fonts/Supplemental/Courier New Bold.ttf",
			fontSize,
		)
		if err != nil {
			return fmt.Errorf("failed to load font: %w", err)
		}

		// Draw text - almost black, slightly gray
		maxWidth := float64(width) * 0.44
		dc.SetRGB(0.1, 0.1, 0.1)
		dc.DrawStringWrapped(displayTitle, float64(width)*0.03+maxWidth/2, float64(height)*0.60, 0.5, 0, maxWidth, 1.8, gg.AlignCenter)

		imagePath := filepath.Join("images", "audioimages", outputFilename)

		// Save the image
		outFile, err := os.Create(imagePath)
		if err != nil {
			return fmt.Errorf("failed to create output file for %q: %w", displayTitle, err)
		}

		if err := png.Encode(outFile, dc.Image()); err != nil {
			outFile.Close()
			return fmt.Errorf("failed to encode image for %q: %w", displayTitle, err)
		}
		outFile.Close()

		fmt.Printf("Saved: %s\n", imagePath)
	}

	fmt.Println("Done generating album art!")
	return nil
}

// GenerateSingleCoverArt generates cover art for a single audio file.
// audioFilePath should be a local path like "/path/to/billy_in_the_lowground.wav".
// Returns the path to the generated image.
func GenerateSingleCoverArt(audioFilePath string) (string, error) {
	baseImageURL := getBaseImage()

	fmt.Println("Downloading base image...")
	baseImage, err := downloadImage(baseImageURL)
	if err != nil {
		return "", err
	}

	bounds := baseImage.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	if err := os.MkdirAll("images/audioimages", 0755); err != nil {
		return "", fmt.Errorf("failed to create images directory: %w", err)
	}

	// Extract base filename without extension
	base := filepath.Base(audioFilePath)
	nameWithoutExt := strings.TrimSuffix(base, filepath.Ext(base))

	// Display title for overlay (e.g., "Billy In The Lowground")
	displayTitle := formatForDisplay(nameWithoutExt)

	// Output filename stays as snake_case to match audio files
	outputFilename := nameWithoutExt + ".png"

	fmt.Printf("Generating album art for: %s\n", displayTitle)

	// Create a new context with the base image
	dc := gg.NewContextForImage(baseImage)

	// Configure text style
	fontSize := float64(width) / 22
	err = dc.LoadFontFace(
		"/System/Library/Fonts/Supplemental/Courier New Bold.ttf",
		fontSize,
	)
	if err != nil {
		return "", fmt.Errorf("failed to load font: %w", err)
	}

	// Draw text - almost black, slightly gray
	maxWidth := float64(width) * 0.44
	dc.SetRGB(0.1, 0.1, 0.1)
	dc.DrawStringWrapped(displayTitle, float64(width)*0.03+maxWidth/2, float64(height)*0.60, 0.5, 0, maxWidth, 1.8, gg.AlignCenter)

	imagePath := filepath.Join("images", "audioimages", outputFilename)

	// Save the image
	outFile, err := os.Create(imagePath)
	if err != nil {
		return "", fmt.Errorf("failed to create output file for %q: %w", displayTitle, err)
	}

	if err := png.Encode(outFile, dc.Image()); err != nil {
		outFile.Close()
		return "", fmt.Errorf("failed to encode image for %q: %w", displayTitle, err)
	}
	outFile.Close()

	fmt.Printf("Saved: %s\n", imagePath)
	return imagePath, nil
}
