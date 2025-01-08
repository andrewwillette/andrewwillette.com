package aws

import (
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
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
	s3Client        *s3.S3
	cache           songCache
	presignedURLTTL = 30 * time.Minute
	cacheTTL        = presignedURLTTL - 1*time.Minute
)

type S3Song struct {
	Name         string
	AudioURL     string
	ImageURL     string
	LastModified time.Time
}

type songCache struct {
	songs []S3Song
	mu    sync.Mutex
}

func init() {
	go cache.updateCache()
	go cache.startAutoUpdate()
}

func GetSongs() ([]S3Song, error) {
	go cache.updateCache()
	return cache.songs, nil
}

func (*songCache) updateCache() {
	cache.mu.Lock()
	songs, err := getS3Songs()
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

func getS3Songs() ([]S3Song, error) {
	log.Debug().Msg("GetS3Songs()")
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
		Prefix: aws.String(audioBucketPrefix),
	}
	now := time.Now()
	audioImageData, err := getS3Client().ListObjectsV2(input)
	if err != nil {
		return nil, err
	}
	log.Debug().Msgf("audioImageData S3 access took %v", time.Since(now))
	wavs := make(map[string]S3Song)
	imgs := make(map[string]string)
	for _, item := range audioImageData.Contents {
		if *item.Key == audioBucketPrefix { // skip the folder itself
			continue
		}
		filetype := wavOrPng(*item.Key)
		audioTitle := formatAudioTitle(*item.Key)
		log.Debug().Msgf("item.Key: %s, audioTitle: %s", *item.Key, audioTitle)
		itemUrl, err := getPresignedURL(*item.Key)
		if err != nil {
			log.Error().Msgf("Failed to get URL for %s: %v", *item.Key, err)
		}
		if filetype == "wav" {
			wavs[audioTitle] = S3Song{
				Name:         audioTitle,
				AudioURL:     itemUrl,
				LastModified: *item.LastModified,
			}
		} else if filetype == "png" {
			imgs[audioTitle] = itemUrl
		}
	}
	toReturn := []S3Song{}
	backupImageURL, err := getPresignedURL("audio/unknown.png")
	if err != nil {
		log.Error().Msgf("Failed to get URL for unknown.png: %v", err)
	}
	for key, s3Song := range wavs {
		s3Song.ImageURL = imgs[key]
		if s3Song.ImageURL == "" {
			log.Debug().Msgf("No image found for %s", s3Song.Name)
			s3Song.ImageURL = backupImageURL
		}
		toReturn = append(toReturn, s3Song)
	}
	sortS3SongsByRecent(toReturn)
	return toReturn, nil
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
	req, _ := getS3Client().GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	})
	urlStr, err := req.Presign(30 * time.Minute) // URL expires in 15 minutes
	if err != nil {
		return "", err
	}
	return urlStr, nil
}

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
