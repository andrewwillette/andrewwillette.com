package aws

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/rs/zerolog/log"

	webCfg "github.com/andrewwillette/andrewwillettedotcom/config"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	s3Client        *s3.Client
	presignedURLTTL = 30 * time.Minute
	cacheTTL        = presignedURLTTL - 1*time.Minute
)

type S3Song struct {
	Name         string
	AudioURL     string
	ImageURL     string
	LastModified time.Time
	Key          string
}

func UploadAudioToS3(filePath string) error {
	log.Debug().Msgf("Uploading audio file %s to S3...", filePath)

	client := getS3Client()

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	key := filepath.Join(webCfg.C.AudioS3BucketPrefix, filepath.Base(filePath))
	contentType := "audio/mpeg"
	if strings.HasSuffix(filePath, ".wav") {
		contentType = "audio/wav"
	}

	uploader := manager.NewUploader(client)

	_, err = uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(webCfg.C.AudioS3BucketName),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(contentType),
		ACL:         types.ObjectCannedACLPublicRead,
	})
	if err != nil {
		return fmt.Errorf("failed to upload to S3: %w", err)
	}

	log.Info().Msgf("Successfully uploaded %s to s3://%s/%s", filePath, webCfg.C.AudioS3BucketName, key)
	return nil
}

func GetAudioFromS3() ([]S3Song, error) {
	log.Debug().Msg("GetS3Songs()")

	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(webCfg.C.AudioS3BucketName),
		Prefix: aws.String(webCfg.C.AudioS3BucketPrefix),
	}

	start := time.Now()
	output, err := getS3Client().ListObjectsV2(context.TODO(), input)
	if err != nil {
		log.Error().Msgf("Failed to list objects in S3: %v", err)
		return nil, err
	}
	log.Debug().Msgf("audioImageData S3 access took %v", time.Since(start))

	wavs := make(map[string]S3Song)
	imgs := make(map[string]string)

	for _, item := range output.Contents {
		if item.Key == nil || *item.Key == webCfg.C.AudioS3BucketPrefix {
			continue
		}
		key := *item.Key
		filetype := wavOrPng(key)
		audioTitle := formatAudioTitle(key)
		log.Debug().Msgf("item.Key: %s, audioTitle: %s", key, audioTitle)

		itemUrl, err := getPresignedURL(key)
		if err != nil {
			log.Error().Msgf("Failed to get URL for %s: %v", key, err)
			continue
		}

		switch filetype {
		case "wav":
			wavs[audioTitle] = S3Song{
				Name:         audioTitle,
				AudioURL:     itemUrl,
				LastModified: aws.ToTime(item.LastModified),
				Key:          *item.Key,
			}
		case "png":
			imgs[audioTitle] = itemUrl
		}
	}

	var songs []S3Song
	backupImageURL, _ := getPresignedURL("audio/webpage_album_cuts_image.png")

	for key, s3Song := range wavs {
		s3Song.ImageURL = imgs[key]
		if s3Song.ImageURL == "" {
			log.Debug().Msgf("No image found for %s", s3Song.Name)
			s3Song.ImageURL = backupImageURL
		}
		songs = append(songs, s3Song)
	}

	sortS3SongsByRecent(songs)
	return songs, nil
}

func DeleteAudioFromS3(key string) error {
	log.Info().Msgf("Deleting audio file %s from S3...", key)

	client := getS3Client()

	if !strings.HasPrefix(key, webCfg.C.AudioS3BucketPrefix) {
		key = filepath.Join(webCfg.C.AudioS3BucketPrefix, key)
	}

	_, err := client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(webCfg.C.AudioS3BucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete object %s from S3: %w", key, err)
	}

	log.Info().Msgf("Successfully deleted %s from S3 bucket %s", key, webCfg.C.AudioS3BucketName)
	return nil
}

func wavOrPng(key string) string {
	if strings.HasSuffix(key, ".wav") {
		return "wav"
	} else if strings.HasSuffix(key, ".png") {
		return "png"
	}
	return ""
}

func formatAudioTitle(filePath string) string {
	base := filepath.Base(filePath)
	name := strings.TrimSuffix(base, filepath.Ext(base))
	name = strings.ReplaceAll(name, "_", " ")
	titleCaser := cases.Title(language.English)
	name = titleCaser.String(name)
	return name
}

func sortS3SongsByRecent(songs []S3Song) {
	sort.Slice(songs, func(i, j int) bool {
		return songs[i].LastModified.After(songs[j].LastModified)
	})
}

const PresignURLExpiry = 60 * time.Minute

func getPresignedURL(key string) (string, error) {
	presigner := s3.NewPresignClient(getS3Client())

	resp, err := presigner.PresignGetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(webCfg.C.AudioS3BucketName),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(PresignURLExpiry))
	if err != nil {
		return "", err
	}

	return resp.URL, nil
}
