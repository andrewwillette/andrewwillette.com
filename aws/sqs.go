package aws

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/rs/zerolog/log"
)

// We listen to an SQS queue which receives events when audio files are added or removed from S3
// and update our local cache accordingly.

var (
	sqsClient       *sqs.Client
	sqsURL          = "https://sqs.us-east-2.amazonaws.com/981566818315/s3-audio-event-queue"
	sqsPollInterval = 10 * time.Second
)

func init() {
	StartSQSPoller(sqsURL)
}

func StartSQSPoller(queueURL string) {
	go func() {
		log.Info().Msg("Starting SQS poller...")
		for {
			msgs, err := receiveSQSMessages(queueURL)
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
					deleteSQSMessage(queueURL, *msg.ReceiptHandle)
				}
			}
		}
	}()
}

func receiveSQSMessages(queueURL string) ([]types.Message, error) {
	if sqsClient == nil {
		cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
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

func handleSQSEvent(msg types.Message) bool {
	var payload struct {
		Records []struct {
			EventName string `json:"eventName"`
			S3        struct {
				Object struct {
					Key string `json:"key"`
				} `json:"object"`
			} `json:"s3"`
		} `json:"Records"`
	}

	if err := json.Unmarshal([]byte(*msg.Body), &payload); err != nil {
		log.Error().Msgf("Invalid SQS message format: %v", err)
		return false
	}

	for _, record := range payload.Records {
		if strings.HasPrefix(record.S3.Object.Key, audioBucketPrefix) &&
			(strings.HasPrefix(record.EventName, "ObjectCreated") || strings.HasPrefix(record.EventName, "ObjectRemoved")) {

			log.Info().Msgf("Detected S3 event %s for %s — updating cache", record.EventName, record.S3.Object.Key)
			go cache.updateCache()
			return true
		}
	}

	log.Debug().Msg("SQS message not relevant to audio/, ignoring.")
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
