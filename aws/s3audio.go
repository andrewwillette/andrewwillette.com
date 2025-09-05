package aws

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/rs/zerolog/log"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	bucketName        = "andrewwillette"
	region            = "us-east-2"
	audioBucketPrefix = "audio/"
)

var (
	s3Client        *s3.Client
	cache           songCache
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

type songCache struct {
	songs []S3Song
	mu    sync.Mutex
}

func init() {
	go cache.updateCache()
	go cache.startAutoUpdate()
}

func UploadAudioToS3(filePath string) error {
	log.Debug().Msgf("Uploading audio file %s to S3...", filePath)

	client := getS3Client()

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	key := filepath.Join(audioBucketPrefix, filepath.Base(filePath))
	contentType := "audio/mpeg"
	if strings.HasSuffix(filePath, ".wav") {
		contentType = "audio/wav"
	}

	uploader := manager.NewUploader(client)

	_, err = uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(contentType),
		ACL:         types.ObjectCannedACLPublicRead,
	})
	if err != nil {
		return fmt.Errorf("failed to upload to S3: %w", err)
	}

	log.Info().Msgf("Successfully uploaded %s to s3://%s/%s", filePath, bucketName, key)
	go cache.updateCache()
	return nil
}

func GetSongsFromCache() ([]S3Song, error) {
	go cache.updateCache()
	return cache.songs, nil
}

func (*songCache) updateCache() {
	log.Debug().Msg("updating the S3 cache")
	cache.mu.Lock()
	songs, err := GetAudioFromS3()
	if err != nil {
		log.Error().Msgf("Unable to get songs from S3: %v", err)
		cache.mu.Unlock()
		return
	} else {
		cache.songs = songs
		cache.mu.Unlock()
	}
}

func (*songCache) startAutoUpdate() {
	ticker := time.NewTicker(cacheTTL)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			log.Debug().Msg("Updating song cache with auto-update")
			cache.updateCache()
		}
	}
}

func GetAudioFromS3() ([]S3Song, error) {
	log.Debug().Msg("GetS3Songs()")

	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
		Prefix: aws.String(audioBucketPrefix),
	}

	start := time.Now()
	output, err := getS3Client().ListObjectsV2(context.TODO(), input)
	if err != nil {
		return nil, err
	}
	log.Debug().Msgf("audioImageData S3 access took %v", time.Since(start))

	wavs := make(map[string]S3Song)
	imgs := make(map[string]string)

	for _, item := range output.Contents {
		if item.Key == nil || *item.Key == audioBucketPrefix {
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
	backupImageURL, _ := getPresignedURL("audio/unknown.png")

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

	if !strings.HasPrefix(key, audioBucketPrefix) {
		key = filepath.Join(audioBucketPrefix, key)
	}

	_, err := client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete object %s from S3: %w", key, err)
	}

	log.Info().Msgf("Successfully deleted %s from S3 bucket %s", key, bucketName)

	go cache.updateCache()
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

func getPresignedURL(key string) (string, error) {
	presigner := s3.NewPresignClient(getS3Client())

	resp, err := presigner.PresignGetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(30*time.Minute))
	if err != nil {
		return "", err
	}

	return resp.URL, nil
}

func getS3Client() *s3.Client {
	if s3Client == nil {
		s3Client = initS3Session()
	}
	return s3Client
}

func initS3Session() *s3.Client {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		log.Fatal().Msgf("Failed to load AWS config: %v", err)
	}
	client := s3.NewFromConfig(cfg)
	s3Client = client
	return client
}
