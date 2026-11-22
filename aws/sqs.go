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

var sqsClient *sqs.Client

// StartSQSPoller starts a background goroutine that long-polls the SQS queue
// and triggers in-memory cache refreshes when S3 content changes.
func StartSQSPoller() {
	go pollSQS()
}

func pollSQS() {
	for {
		msgs, err := receiveSQSMessages(webCfg.C.AudioSQSURL)
		if err != nil {
			log.Error().Msgf("Failed to receive SQS messages: %v", err)
			time.Sleep(10 * time.Second)
			continue
		}

		for _, msg := range msgs {
			handleSQSMessage(msg)
		}
	}
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
		WaitTimeSeconds:     20, // long-polling: reduces API calls and latency
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

// contentType ties an S3 key prefix to the cache that mirrors it.
type contentType struct {
	name    string
	prefix  string
	refresh func()
}

func contentTypes() []contentType {
	return []contentType{
		{"audio", webCfg.C.AudioS3BucketPrefix, UpdateAudioCache},
		{"sheet music", webCfg.C.SheetMusicS3BucketPrefix, UpdateSheetMusicCache},
		{"shows", webCfg.C.ShowsS3BucketPrefix, UpdateShowsCache},
		{"chord charts", webCfg.C.ChordChartsS3BucketPrefix, UpdateChordChartsCache},
	}
}

// handleSQSMessage refreshes affected caches and deletes the message. Messages
// that can't be parsed are deleted too, since retrying would never succeed.
func handleSQSMessage(msg types.Message) {
	matched, err := matchContentTypes(aws.ToString(msg.Body), contentTypes())
	if err != nil {
		log.Error().Msgf("Invalid SQS message format: %v", err)
	} else if len(matched) == 0 {
		log.Info().Msg("SQS message not related to any known content prefix, deleting")
	}
	for _, ct := range matched {
		log.Info().Msgf("Detected %s S3 event, updating cache", ct.name)
		go ct.refresh()
	}
	deleteSQSMessage(webCfg.C.AudioSQSURL, aws.ToString(msg.ReceiptHandle))
}

// matchContentTypes returns each content type touched by any record in body,
// once. Unset prefixes never match, since they would claim every key.
func matchContentTypes(body string, candidates []contentType) ([]contentType, error) {
	var payload S3Event
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		return nil, err
	}

	var matched []contentType
	seen := make(map[string]bool, len(candidates))
	for _, record := range payload.Records {
		// S3 URL-encodes keys in event notifications.
		key, err := url.QueryUnescape(record.S3.Object.Key)
		if err != nil {
			key = record.S3.Object.Key
		}
		for _, ct := range candidates {
			if ct.prefix != "" && !seen[ct.name] && strings.HasPrefix(key, ct.prefix) {
				seen[ct.name] = true
				matched = append(matched, ct)
				break
			}
		}
	}
	return matched, nil
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
