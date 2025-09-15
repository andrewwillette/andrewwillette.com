package aws

import (
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type songCache struct {
	songs []S3Song
	mu    sync.Mutex
}

var cache = songCache{}

func GetCachedAudio() ([]S3Song, error) {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	return cache.songs, nil
}

// UpdateAudioCacheOnPresignExpiry resolves bug where presignedURLs become stale
// if cache not updated
func UpdateAudioCacheOnPresignExpiry() {
	for {
		time.Sleep(PresignURLExpiry - 1*time.Minute)
		UpdateAudioCache()
	}
}

func UpdateAudioCache() {
	log.Debug().Msg("Updating song cache...")
	songs, err := GetAudioFromS3()
	if err != nil {
		log.Error().Msgf("Failed to update cache: %v", err)
		return
	}

	cache.mu.Lock()
	defer cache.mu.Unlock()
	cache.songs = songs
}
