package aws

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"

	webCfg "github.com/andrewwillette/andrewwillettedotcom/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/rs/zerolog/log"
)

// ChordChartJSONObject is the JSON object stored in S3 for a single chord chart.
type ChordChartJSONObject struct {
	Title         string `json:"title"`
	Artist        string `json:"artist"`
	DefaultKey    string `json:"default_key"`
	TimeSignature string `json:"time_signature,omitempty"`
	Notes         string `json:"notes,omitempty"`
	URLVal        string `json:"url_val"`
	FileName      string `json:"file_name"`
	Body          string `json:"body"`
}

// ChordChartSyncResult reports what SyncChordCharts actually did, classified
// by comparing a content checksum of each incoming chart against what's
// already stored in S3.
type ChordChartSyncResult struct {
	Added     []string
	Updated   []string
	Unchanged []string
	Deleted   []string
}

type chordChartCache struct {
	charts []ChordChartJSONObject
	mu     sync.Mutex
}

var chordChartCacheInst = chordChartCache{}

func GetCachedChordCharts() []ChordChartJSONObject {
	chordChartCacheInst.mu.Lock()
	defer chordChartCacheInst.mu.Unlock()
	cp := make([]ChordChartJSONObject, len(chordChartCacheInst.charts))
	copy(cp, chordChartCacheInst.charts)
	return cp
}

func UpdateChordChartsCache() {
	log.Debug().Msg("Updating chord chart cache...")
	items, err := listChordChartsFromS3()
	if err != nil {
		log.Error().Err(err).Msg("Failed to list chord charts from S3")
		return
	}
	sort.Slice(items, func(i, j int) bool {
		return strings.ToLower(items[i].Title) < strings.ToLower(items[j].Title)
	})
	chordChartCacheInst.mu.Lock()
	chordChartCacheInst.charts = items
	chordChartCacheInst.mu.Unlock()
	log.Info().Msgf("Chord chart cache updated, %d entries", len(items))
}

// chordChartChecksum hashes the fields that matter for display, so a sync can
// tell an unchanged chart apart from one whose content actually differs
// without re-uploading every chart on every sync.
func chordChartChecksum(item ChordChartJSONObject) string {
	h := sha256.Sum256([]byte(item.Title + "\x00" + item.Artist + "\x00" + item.DefaultKey + "\x00" + item.TimeSignature + "\x00" + item.Notes + "\x00" + item.Body))
	return hex.EncodeToString(h[:])
}

// SyncChordCharts makes S3 match items exactly: charts whose content checksum
// differs from (or is absent from) what's stored are uploaded, charts whose
// checksum already matches are left alone, and any chart in S3 that isn't in
// items is deleted.
func SyncChordCharts(items []ChordChartJSONObject) (ChordChartSyncResult, error) {
	var result ChordChartSyncResult

	existing, err := listChordChartsFromS3()
	if err != nil {
		return result, err
	}
	existingByURLVal := make(map[string]ChordChartJSONObject, len(existing))
	for _, e := range existing {
		existingByURLVal[e.URLVal] = e
	}

	client := getS3Client()
	seen := make(map[string]bool, len(items))
	for _, item := range items {
		seen[item.URLVal] = true
		old, existed := existingByURLVal[item.URLVal]
		if existed && chordChartChecksum(old) == chordChartChecksum(item) {
			result.Unchanged = append(result.Unchanged, item.URLVal)
			continue
		}
		if err := putChordChart(client, item); err != nil {
			return result, fmt.Errorf("uploading %s: %w", item.URLVal, err)
		}
		if existed {
			result.Updated = append(result.Updated, item.URLVal)
		} else {
			result.Added = append(result.Added, item.URLVal)
		}
	}

	for urlVal := range existingByURLVal {
		if seen[urlVal] {
			continue
		}
		if err := deleteChordChart(client, urlVal); err != nil {
			return result, fmt.Errorf("deleting %s: %w", urlVal, err)
		}
		result.Deleted = append(result.Deleted, urlVal)
	}

	log.Info().Msgf("Synced chord charts to S3: %d added, %d updated, %d unchanged, %d deleted",
		len(result.Added), len(result.Updated), len(result.Unchanged), len(result.Deleted))
	return result, nil
}

func putChordChart(client *s3.Client, item ChordChartJSONObject) error {
	body, err := json.MarshalIndent(item, "", "  ")
	if err != nil {
		return err
	}
	_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(webCfg.C.ChordChartsS3BucketName),
		Key:         aws.String(chordChartKey(item.URLVal)),
		Body:        strings.NewReader(string(body)),
		ContentType: aws.String("application/json"),
	})
	return err
}

func deleteChordChart(client *s3.Client, urlVal string) error {
	_, err := client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(webCfg.C.ChordChartsS3BucketName),
		Key:    aws.String(chordChartKey(urlVal)),
	})
	return err
}

func chordChartKey(urlVal string) string {
	return ensureTrailingSlash(webCfg.C.ChordChartsS3BucketPrefix) + urlVal + ".json"
}

func listChordChartsFromS3() ([]ChordChartJSONObject, error) {
	client := getS3Client()
	out, err := client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String(webCfg.C.ChordChartsS3BucketName),
		Prefix: aws.String(ensureTrailingSlash(webCfg.C.ChordChartsS3BucketPrefix)),
	})
	if err != nil {
		return nil, err
	}

	items := make([]ChordChartJSONObject, 0, len(out.Contents))
	for _, obj := range out.Contents {
		if obj.Key == nil {
			continue
		}
		key := *obj.Key
		if key == webCfg.C.ChordChartsS3BucketPrefix || !strings.HasSuffix(strings.ToLower(key), ".json") {
			continue
		}

		resp, err := client.GetObject(context.TODO(), &s3.GetObjectInput{
			Bucket: aws.String(webCfg.C.ChordChartsS3BucketName),
			Key:    aws.String(key),
		})
		if err != nil {
			log.Warn().Err(err).Msgf("Failed reading %s", key)
			continue
		}
		b, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			log.Warn().Err(err).Msgf("Failed reading body of %s", key)
			continue
		}

		var item ChordChartJSONObject
		if err := json.Unmarshal(b, &item); err != nil {
			log.Warn().Err(err).Msgf("Failed unmarshaling %s", key)
			continue
		}
		items = append(items, item)
	}
	return items, nil
}
