package aws

import (
	"context"
	"sync"

	webCfg "github.com/andrewwillette/andrewwillettedotcom/config"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/rs/zerolog/log"
)

var (
	clientsMu sync.Mutex
	clients   = map[string]*s3.Client{} // keyed by AWS region
)

// s3ClientForRegion returns a cached client for region, creating it on first
// use. An empty region falls back to the audio region, the site's default.
func s3ClientForRegion(region string) *s3.Client {
	if region == "" {
		region = webCfg.C.AudioS3Region
	}

	clientsMu.Lock()
	defer clientsMu.Unlock()

	if c, ok := clients[region]; ok {
		return c
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		log.Fatal().Msgf("Failed to load AWS config for region %s: %v", region, err)
	}
	c := s3.NewFromConfig(cfg)
	clients[region] = c
	return c
}

func getAudioS3Client() *s3.Client { return s3ClientForRegion(webCfg.C.AudioS3Region) }

func getSheetMusicS3Client() *s3.Client { return s3ClientForRegion(webCfg.C.SheetMusicS3Region) }

func getShowsS3Client() *s3.Client { return s3ClientForRegion(webCfg.C.ShowsS3Region) }

func getChordChartsS3Client() *s3.Client { return s3ClientForRegion(webCfg.C.ChordChartsS3Region) }

// getBackupS3Client returns the client for backup operations, which have no
// region setting of their own and use the default region.
func getBackupS3Client() *s3.Client { return s3ClientForRegion("") }
