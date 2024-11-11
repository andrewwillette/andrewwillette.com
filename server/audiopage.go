package server

import (
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var s3Client *s3.S3

func getS3Client() *s3.S3 {
	if s3Client == nil {
		s3Client = initS3Session()
	}
	return s3Client
}

func initS3Session() *s3.S3 {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		log.Fatal().Msgf("Failed to create S3 session: %v", err)
	}
	return s3.New(sess)
}

func handleRecordingsPage(c echo.Context) error {
	// time the listSongsWithRandomImage function and log the duration
	start := time.Now()
	songs, err := listSongsWithRandomImage()
	log.Debug().Msgf("listSongsWithRandomImage took %v", time.Since(start))
	if err != nil {
		// http.Error(w, "Unable to list songs", http.StatusInternalServerError)
		log.Error().Msgf("Unable to list songs: %v", err)
		return err
	}
	err = c.Render(http.StatusOK, "musicpagenew", songs)
	if err != nil {
		log.Error().Msgf("Unable to render musicpagenew: %v", err)
		return err
	}
	return nil
}

type S3Song struct {
	Name     string
	AudioURL string
	ImageURL string
}

const (
	bucketName  = "andrewwillette"
	region      = "us-east-2" // adjust based on your S3 region
	audioPrefix = "audio/"    // Prefix for audio files
	imagePrefix = "audioimages/"
)

func listSongsWithRandomImage() ([]S3Song, error) {
	log.Debug().Msg("listing songs with random image")
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
		Prefix: aws.String(audioPrefix),
	}
	now := time.Now()
	audioImageData, err := getS3Client().ListObjectsV2(input)
	if err != nil {
		return nil, err
	}
	log.Debug().Msgf("audioImageData S3 access took %v", time.Since(now))
	wavs := make(map[string]string)
	imgs := make(map[string]string)
	for _, item := range audioImageData.Contents {
		if *item.Key == audioPrefix { // skip the folder itself
			continue
		}
		wavorpng := func(key string) string { // returns "wav" or "png"
			if strings.HasSuffix(key, ".wav") {
				return "wav"
			} else if strings.HasSuffix(key, ".png") {
				return "png"
			}
			return ""
		}
		filetype := wavorpng(*item.Key)
		mapsKey := formatAudioTitle(*item.Key)
		log.Debug().Msgf("item.Key: %s", *item.Key)
		itemUrl, err := getPresignedURL(getS3Client(), *item.Key)
		if err != nil {
			log.Error().Msgf("Failed to get URL for %s: %v", *item.Key, err)
		}
		if filetype == "wav" {
			wavs[mapsKey] = itemUrl
		} else if filetype == "png" {
			imgs[mapsKey] = itemUrl
		}
	}
	toReturn := []S3Song{}
	backupImageURL, err := getPresignedURL(getS3Client(), "audio/unknown.png")
	if err != nil {
		log.Error().Msgf("Failed to get URL for unknown.png: %v", err)
	}
	for key, songURL := range wavs {
		song := S3Song{AudioURL: songURL, ImageURL: imgs[key], Name: key}
		if song.ImageURL == "" {
			log.Warn().Msgf("No image found for %s", song.Name)
			song.ImageURL = backupImageURL
		}
		toReturn = append(toReturn, song)
	}
	return toReturn, nil
}

// getPresignedURL returns a presigned URL for the given key
// used by the browser client to download the file
func getPresignedURL(svc *s3.S3, key string) (string, error) {
	req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	})
	urlStr, err := req.Presign(30 * time.Minute) // URL expires in 15 minutes
	if err != nil {
		return "", err
	}
	return urlStr, nil
}

func formatAudioTitle(filePath string) string {
	base := filepath.Base(filePath)
	name := strings.TrimSuffix(base, filepath.Ext(base))
	name = strings.ReplaceAll(name, "_", " ")
	titleCaser := cases.Title(language.English)
	name = titleCaser.String(name)
	return name
}
