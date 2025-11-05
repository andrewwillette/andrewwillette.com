package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	webCfg "github.com/andrewwillette/andrewwillettedotcom/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/rs/zerolog/log"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// JSON object stored in S3
type SheetMusicJSONObject struct {
	DisplayName string `json:"display_name"`
	DropboxURL  string `json:"url"`
}

// SheetMusicAdminObject used for managing the S3 state with various commands
type SheetMusicAdminObject struct {
	Key         string
	DisplayName string
}

type sheetCache struct {
	sheets []SheetMusicJSONObject
	mu     sync.Mutex
}

var sheetCacheInst = sheetCache{}

func GetCachedSheetMusic() ([]SheetMusicJSONObject, error) {
	sheetCacheInst.mu.Lock()
	defer sheetCacheInst.mu.Unlock()
	cp := make([]SheetMusicJSONObject, len(sheetCacheInst.sheets))
	// shallow copy the cache slice
	copy(cp, sheetCacheInst.sheets)
	return cp, nil
}

func UpdateSheetMusicCache() {
	log.Debug().Msg("Updating sheet music cache...")
	items, err := ListSheetMusicFromS3()
	if err != nil {
		log.Error().Err(err).Msg("Failed to list sheet music from S3")
		return
	}
	sort.Slice(items, func(i, j int) bool {
		return strings.ToLower(items[i].DisplayName) < strings.ToLower(items[j].DisplayName)
	})
	sheetCacheInst.mu.Lock()
	sheetCacheInst.sheets = items
	sheetCacheInst.mu.Unlock()
	log.Info().Msgf("Sheet music cache updated, %d entries", len(items))
}

func PutSheetJSON(displayName, dropboxURL string) error {
	// slug is the name of object in S3 bucket, derived from display name. (eg derived_display_name.json)
	slug := slugify(displayName)
	key := ensureTrailingSlash(webCfg.C.SheetMusicS3BucketPrefix) + slug + ".json"

	item := SheetMusicJSONObject{
		DisplayName: strings.TrimSpace(displayName),
		DropboxURL:  normalizeDropboxURL(strings.TrimSpace(dropboxURL)),
	}
	body, err := json.MarshalIndent(item, "", "  ")
	if err != nil {
		return err
	}

	_, err = getS3Client().PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(webCfg.C.SheetMusicS3BucketName),
		Key:         aws.String(key),
		Body:        strings.NewReader(string(body)),
		ContentType: aws.String("application/json"),
	})
	if err != nil {
		return err
	}

	log.Info().Msgf("Uploaded sheet JSON: s3://%s/%s", webCfg.C.SheetMusicS3BucketName, key)
	return nil
}

func DeleteSheetMusicFromS3(key string) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return fmt.Errorf("empty key")
	}
	if !strings.HasPrefix(key, ensureTrailingSlash(webCfg.C.SheetMusicS3BucketPrefix)) {
		key = ensureTrailingSlash(webCfg.C.SheetMusicS3BucketPrefix) + key
	}
	if !strings.HasSuffix(strings.ToLower(key), ".json") {
		key = key + ".json"
	}

	_, err := getS3Client().DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(webCfg.C.SheetMusicS3BucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return err
	}

	log.Info().Msgf("Deleted sheet JSON: s3://%s/%s", webCfg.C.SheetMusicS3BucketName, key)
	return nil
}

func DeleteSheetMusicByDisplayName(displayName string) error {
	slug := slugify(displayName)
	key := ensureTrailingSlash(webCfg.C.SheetMusicS3BucketPrefix) + slug + ".json"
	return DeleteSheetMusicFromS3(key)
}

func ListSheetMusicFromS3() ([]SheetMusicJSONObject, error) {
	rows, err := listSheetJSONRaw()
	if err != nil {
		return nil, err
	}
	out := make([]SheetMusicJSONObject, 0, len(rows))
	for _, r := range rows {
		item := r.JSONItem
		if strings.TrimSpace(item.DisplayName) == "" {
			item.DisplayName = fallbackNameFromKey(r.Key, webCfg.C.SheetMusicS3BucketPrefix)
		}
		item.DropboxURL = normalizeDropboxURL(item.DropboxURL)
		out = append(out, item)
	}
	return out, nil
}

// ListSheetMusicObjects admin-friendly key + display name (for delete/rename)
func ListSheetMusicObjects() ([]SheetMusicAdminObject, error) {
	rows, err := listSheetJSONRaw()
	if err != nil {
		return nil, err
	}
	out := make([]SheetMusicAdminObject, 0, len(rows))
	for _, r := range rows {
		name := strings.TrimSpace(r.JSONItem.DisplayName)
		if name == "" {
			name = fallbackNameFromKey(r.Key, webCfg.C.SheetMusicS3BucketPrefix)
		}
		out = append(out, SheetMusicAdminObject{
			Key:         r.Key,
			DisplayName: name,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		return strings.ToLower(out[i].DisplayName) < strings.ToLower(out[j].DisplayName)
	})
	return out, nil
}

// sheetMusicS3Object the S3 key and associated JSON object for the key
type sheetMusicS3Object struct {
	Key      string
	JSONItem SheetMusicJSONObject
}

// listSheetJSONRaw main entry point for retrieving
func listSheetJSONRaw() ([]sheetMusicS3Object, error) {
	client := getS3Client()
	out, err := client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String(webCfg.C.SheetMusicS3BucketName),
		Prefix: aws.String(ensureTrailingSlash(webCfg.C.SheetMusicS3BucketPrefix)),
	})
	if err != nil {
		return nil, err
	}

	rows := make([]sheetMusicS3Object, 0, len(out.Contents))
	for _, obj := range out.Contents {
		if obj.Key == nil {
			continue
		}
		key := *obj.Key

		if key == webCfg.C.SheetMusicS3BucketPrefix || !strings.HasSuffix(strings.ToLower(key), ".json") {
			continue
		}

		// we need to read the individual json objects after getting list of all the keys
		sheetMusicJSON, err := readSheetMusicJSONFromS3(client, webCfg.C.SheetMusicS3BucketName, key)
		if err != nil {
			log.Warn().Err(err).Msgf("Failed reading %s", key)
			continue
		}
		rows = append(rows, sheetMusicS3Object{Key: key, JSONItem: sheetMusicJSON})
	}
	return rows, nil
}

func readSheetMusicJSONFromS3(client *s3.Client, bucket, key string) (SheetMusicJSONObject, error) {
	resp, err := client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return SheetMusicJSONObject{}, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return SheetMusicJSONObject{}, err
	}

	var item SheetMusicJSONObject
	if err := json.Unmarshal(b, &item); err != nil {
		return SheetMusicJSONObject{}, err
	}
	return item, nil
}

func slugify(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, "-", "_")
	out := make([]rune, 0, len(s))
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			out = append(out, r)
		}
	}
	s = string(out)
	return strings.Trim(s, "_")
}

func fallbackNameFromKey(key, prefix string) string {
	base := strings.TrimPrefix(key, ensureTrailingSlash(prefix))
	base = strings.TrimSuffix(base, filepath.Ext(base))
	base = strings.ReplaceAll(base, "_", " ")
	base = strings.ReplaceAll(base, "-", " ")
	c := cases.Title(language.AmericanEnglish)
	return c.String(base)
}

func ensureTrailingSlash(p string) string {
	if p == "" || strings.HasSuffix(p, "/") {
		return p
	}
	return p + "/"
}

func normalizeDropboxURL(URL string) string {
	if URL == "" {
		return URL
	}
	loweredURL := strings.ToLower(URL)
	if !strings.Contains(loweredURL, "dropbox.com") {
		return URL
	}
	// remove automatic download references if they exist
	if strings.Contains(URL, "?dl=") {
		return reparam(URL, "dl", "0")
	}
	if strings.Contains(URL, "?") {
		return URL + "&dl=0"
	}
	return URL + "?dl=0"
}

func reparam(u, key, val string) string {
	k1 := key + "="
	if i := strings.Index(u, k1); i != -1 {
		j := i + len(k1)
		if end := strings.Index(u[j:], "&"); end == -1 {
			u = u[:i]
		} else {
			u = u[:i] + u[j+end+1:]
		}
		u = strings.TrimSuffix(u, "?")
		u = strings.TrimSuffix(u, "&")
	}
	sep := "?"
	if strings.Contains(u, "?") {
		sep = "&"
	}
	return u + sep + key + "=" + val
}
