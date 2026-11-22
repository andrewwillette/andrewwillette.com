// Package backup mirrors S3 buckets into a single local SQLite file, so the
// site's content survives losing the AWS account and can be re-uploaded
// to any S3-compatible store.
package backup

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/andrewwillette/andrewwillettedotcom/aws"
	"github.com/rs/zerolog/log"
	_ "modernc.org/sqlite"
)

const schema = `
CREATE TABLE IF NOT EXISTS objects (
    bucket        TEXT NOT NULL,
    key           TEXT NOT NULL,
    etag          TEXT NOT NULL,
    size          INTEGER NOT NULL,
    last_modified DATETIME NOT NULL,
    content_type  TEXT,
    data          BLOB NOT NULL,
    synced_at     DATETIME NOT NULL,
    deleted_at    DATETIME, -- set when the object disappears from S3; data is kept
    PRIMARY KEY (bucket, key)
);`

// Result counts what a sync did to the archive, per bucket.
type Result struct {
	Bucket                                 string
	Added, Updated, Unchanged, MarkedGone int
}

// Sync brings the archive at dbPath up to date with each bucket. Objects
// are re-downloaded only when their ETag changed; objects removed from S3
// are marked deleted but kept.
func Sync(dbPath string, buckets []string) ([]Result, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	if _, err := db.Exec(schema); err != nil {
		return nil, fmt.Errorf("creating schema: %w", err)
	}

	var results []Result
	for _, bucket := range buckets {
		r, err := syncBucket(db, bucket)
		if err != nil {
			return results, err
		}
		results = append(results, r)
	}
	return results, nil
}

type archived struct {
	etag    string
	deleted bool
}

func syncBucket(db *sql.DB, bucket string) (Result, error) {
	r := Result{Bucket: bucket}

	existing := map[string]archived{}
	rows, err := db.Query(`SELECT key, etag, deleted_at IS NOT NULL FROM objects WHERE bucket = ?`, bucket)
	if err != nil {
		return r, err
	}
	for rows.Next() {
		var key string
		var a archived
		if err := rows.Scan(&key, &a.etag, &a.deleted); err != nil {
			rows.Close()
			return r, err
		}
		existing[key] = a
	}
	rows.Close()

	objects, err := aws.ListAllObjects(bucket)
	if err != nil {
		return r, err
	}

	now := time.Now().UTC()
	seen := map[string]bool{}
	for _, obj := range objects {
		seen[obj.Key] = true
		prev, ok := existing[obj.Key]
		if ok && prev.etag == obj.ETag && !prev.deleted {
			r.Unchanged++
			continue
		}

		data, contentType, err := aws.GetObjectBytes(bucket, obj.Key)
		if err != nil {
			return r, fmt.Errorf("downloading s3://%s/%s: %w", bucket, obj.Key, err)
		}
		_, err = db.Exec(`
			INSERT INTO objects (bucket, key, etag, size, last_modified, content_type, data, synced_at, deleted_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, NULL)
			ON CONFLICT (bucket, key) DO UPDATE SET
				etag = excluded.etag, size = excluded.size, last_modified = excluded.last_modified,
				content_type = excluded.content_type, data = excluded.data,
				synced_at = excluded.synced_at, deleted_at = NULL`,
			bucket, obj.Key, obj.ETag, obj.Size, obj.LastModified, contentType, data, now)
		if err != nil {
			return r, fmt.Errorf("archiving s3://%s/%s: %w", bucket, obj.Key, err)
		}
		if ok {
			r.Updated++
		} else {
			r.Added++
		}
		log.Debug().Msgf("backup: archived s3://%s/%s (%d bytes)", bucket, obj.Key, len(data))
	}

	for key, prev := range existing {
		if seen[key] || prev.deleted {
			continue
		}
		if _, err := db.Exec(`UPDATE objects SET deleted_at = ? WHERE bucket = ? AND key = ?`, now, bucket, key); err != nil {
			return r, err
		}
		r.MarkedGone++
	}
	return r, nil
}
