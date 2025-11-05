package aws

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"
	"time"

	webCfg "github.com/andrewwillette/andrewwillettedotcom/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/rs/zerolog/log"
)

var (
	sqsClient       *sqs.Client
	sqsPollInterval = 10 * time.Second
)

func StartSQSPoller() {
	go func() {
		for {
			msgs, err := receiveSQSMessages(webCfg.C.AudioSQSURL)
			if err != nil {
				log.Error().Msgf("Failed to receive SQS messages: %v", err)
				time.Sleep(sqsPollInterval)
				continue
			}

			if len(msgs) == 0 {
				time.Sleep(sqsPollInterval)
				continue
			}

			for _, msg := range msgs {
				handled := handleSQSEvent(msg)
				if handled {
					deleteSQSMessage(webCfg.C.AudioSQSURL, *msg.ReceiptHandle)
				} else {
					log.Info().Msg("SQS message on queue not related to audio, deleting it")
					deleteSQSMessage(webCfg.C.AudioSQSURL, *msg.ReceiptHandle)
				}
			}
		}
	}()
}

func receiveSQSMessages(queueURL string) ([]types.Message, error) {
	if sqsClient == nil {
		cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(webCfg.C.AudioS3Region))
		if err != nil {
			return nil, err
		}
		sqsClient = sqs.NewFromConfig(cfg)
	}

	resp, err := sqsClient.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(queueURL),
		MaxNumberOfMessages: 5,
		WaitTimeSeconds:     0,
	})
	if err != nil {
		return nil, err
	}

	return resp.Messages, nil
}

type S3Event struct {
	Records []struct {
		EventName string `json:"eventName"`
		S3        struct {
			Object struct {
				Key string `json:"key"`
			} `json:"object"`
		} `json:"s3"`
	} `json:"Records"`
}

func handleSQSEvent(msg types.Message) bool {
	var payload S3Event

	if err := json.Unmarshal([]byte(*msg.Body), &payload); err != nil {
		log.Error().Msgf("Invalid SQS message format: %v", err)
		return false
	}

	for _, record := range payload.Records {
		key, _ := url.QueryUnescape(record.S3.Object.Key)
		switch {
		case strings.HasPrefix(key, webCfg.C.AudioS3BucketPrefix):
			log.Info().Msgf("Detected audiodata S3 event %s for %s — updating cache", record.EventName, record.S3.Object.Key)
			go UpdateAudioCache()
			return true
		case strings.HasPrefix(key, webCfg.C.SheetMusicS3BucketPrefix):
			log.Info().Msgf("Detected sheetmusic S3 event %s for %s — updating cache", record.EventName, record.S3.Object.Key)
			go UpdateSheetMusicCache()
			return true
		}
	}

	log.Debug().Msg("SQS message not relevant to audio or sheetmusic, ignoring.")
	return false
}

func deleteSQSMessage(queueURL, receiptHandle string) {
	_, err := sqsClient.DeleteMessage(context.TODO(), &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(queueURL),
		ReceiptHandle: aws.String(receiptHandle),
	})
	if err != nil {
		log.Error().Msgf("Failed to delete SQS message: %v", err)
	}
}
