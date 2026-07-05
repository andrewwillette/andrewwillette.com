package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
	"time"

	webCfg "github.com/andrewwillette/andrewwillettedotcom/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/rs/zerolog/log"
)

type ShowJSONObject struct {
	Title       string `json:"title"`
	Date        string `json:"date"` // YYYY-MM-DD
	Description string `json:"description"`
}

// FormattedDate returns Date as "January 2, 2006", or the raw string if unparseable.
func (s ShowJSONObject) FormattedDate() string {
	t, err := time.Parse("2006-01-02", s.Date)
	if err != nil || s.Date == "" {
		return s.Date
	}
	return t.Format("January 2, 2006")
}

type showsCache struct {
	shows []ShowJSONObject
	mu    sync.Mutex
}

var showsCacheInst = showsCache{}

func GetCachedShows() ([]ShowJSONObject, error) {
	showsCacheInst.mu.Lock()
	defer showsCacheInst.mu.Unlock()
	cp := make([]ShowJSONObject, len(showsCacheInst.shows))
	copy(cp, showsCacheInst.shows)
	return cp, nil
}

func UpdateShowsCache() {
	log.Debug().Msg("Updating shows cache...")
	items, err := ListShowsFromS3()
	if err != nil {
		log.Error().Err(err).Msg("Failed to list shows from S3")
		return
	}
	today := time.Now().Format("2006-01-02")
	// upcoming (>= today) ascending, past (< today) descending, empty last
	sort.Slice(items, func(i, j int) bool {
		di, dj := items[i].Date, items[j].Date
		iEmpty, jEmpty := di == "", dj == ""
		if iEmpty && jEmpty {
			return strings.ToLower(items[i].Title) < strings.ToLower(items[j].Title)
		}
		if iEmpty {
			return false
		}
		if jEmpty {
			return true
		}
		iUpcoming := di >= today
		jUpcoming := dj >= today
		if iUpcoming != jUpcoming {
			return iUpcoming // upcoming before past
		}
		if iUpcoming {
			return di < dj // upcoming: nearest first
		}
		return di > dj // past: most recent first
	})
	showsCacheInst.mu.Lock()
	showsCacheInst.shows = items
	showsCacheInst.mu.Unlock()
	log.Info().Msgf("Shows cache updated, %d entries", len(items))
}

func PutShowJSON(title, date, description string) error {
	slug := slugify(title)
	key := ensureTrailingSlash(webCfg.C.ShowsS3BucketPrefix) + slug + ".json"

	item := ShowJSONObject{
		Title:       strings.TrimSpace(title),
		Date:        strings.TrimSpace(date),
		Description: strings.TrimSpace(description),
	}
	body, err := json.MarshalIndent(item, "", "  ")
	if err != nil {
		return err
	}

	_, err = getS3Client().PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(webCfg.C.ShowsS3BucketName),
		Key:         aws.String(key),
		Body:        strings.NewReader(string(body)),
		ContentType: aws.String("application/json"),
	})
	if err != nil {
		return err
	}

	log.Info().Msgf("Uploaded show JSON: s3://%s/%s", webCfg.C.ShowsS3BucketName, key)
	return nil
}

func ListShowsFromS3() ([]ShowJSONObject, error) {
	client := getS3Client()
	out, err := client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String(webCfg.C.ShowsS3BucketName),
		Prefix: aws.String(ensureTrailingSlash(webCfg.C.ShowsS3BucketPrefix)),
	})
	if err != nil {
		return nil, err
	}

	items := make([]ShowJSONObject, 0, len(out.Contents))
	for _, obj := range out.Contents {
		if obj.Key == nil {
			continue
		}
		key := *obj.Key
		if key == webCfg.C.ShowsS3BucketPrefix || !strings.HasSuffix(strings.ToLower(key), ".json") {
			continue
		}

		item, err := readShowJSONFromS3(client, webCfg.C.ShowsS3BucketName, key)
		if err != nil {
			log.Warn().Err(err).Msgf("Failed reading show %s", key)
			continue
		}
		if strings.TrimSpace(item.Title) == "" {
			item.Title = fallbackNameFromKey(key, webCfg.C.ShowsS3BucketPrefix)
		}
		items = append(items, item)
	}
	return items, nil
}

type ShowAdminObject struct {
	Key   string
	Title string
	Date  string
}

func ListShowObjects() ([]ShowAdminObject, error) {
	client := getS3Client()
	out, err := client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String(webCfg.C.ShowsS3BucketName),
		Prefix: aws.String(ensureTrailingSlash(webCfg.C.ShowsS3BucketPrefix)),
	})
	if err != nil {
		return nil, err
	}

	items := make([]ShowAdminObject, 0, len(out.Contents))
	for _, obj := range out.Contents {
		if obj.Key == nil {
			continue
		}
		key := *obj.Key
		if key == webCfg.C.ShowsS3BucketPrefix || !strings.HasSuffix(strings.ToLower(key), ".json") {
			continue
		}
		item, err := readShowJSONFromS3(client, webCfg.C.ShowsS3BucketName, key)
		if err != nil {
			log.Warn().Err(err).Msgf("Failed reading show %s", key)
			continue
		}
		title := strings.TrimSpace(item.Title)
		if title == "" {
			title = fallbackNameFromKey(key, webCfg.C.ShowsS3BucketPrefix)
		}
		items = append(items, ShowAdminObject{Key: key, Title: title, Date: item.Date})
	}
	sort.Slice(items, func(i, j int) bool {
		return strings.ToLower(items[i].Title) < strings.ToLower(items[j].Title)
	})
	return items, nil
}

func DeleteShowFromS3(key string) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return fmt.Errorf("empty key")
	}
	_, err := getS3Client().DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(webCfg.C.ShowsS3BucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return err
	}
	log.Info().Msgf("Deleted show JSON: s3://%s/%s", webCfg.C.ShowsS3BucketName, key)
	return nil
}

func readShowJSONFromS3(client *s3.Client, bucket, key string) (ShowJSONObject, error) {
	resp, err := client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return ShowJSONObject{}, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return ShowJSONObject{}, err
	}

	var item ShowJSONObject
	if err := json.Unmarshal(b, &item); err != nil {
		return ShowJSONObject{}, err
	}
	return item, nil
}
