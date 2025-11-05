package aws

import (
	"context"

	webCfg "github.com/andrewwillette/andrewwillettedotcom/config"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/rs/zerolog/log"
)

func getS3Client() *s3.Client {
	if s3Client == nil {
		s3Client = initS3Session()
	}
	return s3Client
}

func initS3Session() *s3.Client {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(webCfg.C.AudioS3Region))
	if err != nil {
		log.Fatal().Msgf("Failed to load AWS config: %v", err)
	}
	client := s3.NewFromConfig(cfg)
	s3Client = client
	return client
}
