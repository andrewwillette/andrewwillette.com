package traffic

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/andrewwillette/andrewwillettedotcom/aws"
	"github.com/andrewwillette/andrewwillettedotcom/config"
	"github.com/rs/zerolog/log"
)

const backupInterval = time.Hour

// backupKey is the S3 key holding the latest traffic DB snapshot,
// e.g. "backups/traffic.db".
func backupKey() string {
	prefix := config.C.BackupS3BucketPrefix
	if prefix != "" && prefix[len(prefix)-1] != '/' {
		prefix += "/"
	}
	return prefix + "traffic.db"
}

func backupEnabled() bool {
	return config.C.BackupS3BucketName != ""
}

// RestoreFromS3IfEmpty pulls the latest snapshot from S3 when the local
// database is missing or empty (as on a fresh instance, where the deploy
// script has just touched the file). Must run before InitDB.
func RestoreFromS3IfEmpty(dbPath string) error {
	if !backupEnabled() {
		return nil
	}
	if info, err := os.Stat(dbPath); err == nil && info.Size() > 0 {
		return nil
	}

	// Download to a temp file first so a failed download can't leave a
	// half-written database behind.
	tmp, err := os.CreateTemp("", "traffic-restore-*.db")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	found, err := aws.GetObjectToWriter(config.C.BackupS3BucketName, backupKey(), tmp)
	tmp.Close()
	if err != nil {
		return fmt.Errorf("downloading s3://%s/%s: %w", config.C.BackupS3BucketName, backupKey(), err)
	}
	if !found {
		log.Info().Msg("traffic: no backup found in S3, starting with an empty database")
		return nil
	}

	// Copy contents into dbPath rather than renaming over it: in prod the
	// path is a bind-mounted file, which can't be replaced by rename.
	b, err := os.ReadFile(tmp.Name())
	if err != nil {
		return err
	}
	if err := os.WriteFile(dbPath, b, 0o644); err != nil {
		return err
	}
	log.Info().Msgf("traffic: restored %d bytes from s3://%s/%s", len(b), config.C.BackupS3BucketName, backupKey())
	return nil
}

// StartBackupJob uploads a consistent snapshot of the traffic database to
// S3 every backupInterval.
func StartBackupJob() {
	if !backupEnabled() {
		log.Info().Msg("traffic: BACKUP_S3_BUCKET_NAME not set, S3 backups disabled")
		return
	}
	go func() {
		for {
			time.Sleep(backupInterval)
			if err := backupToS3(); err != nil {
				log.Error().Err(err).Msg("traffic: backup to S3 failed")
			}
		}
	}()
}

func backupToS3() error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}
	dir, err := os.MkdirTemp("", "traffic-backup-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	// VACUUM INTO writes a consistent, compacted copy without blocking writers.
	snapshot := filepath.Join(dir, "traffic.db")
	if _, err := db.Exec("VACUUM INTO ?", snapshot); err != nil {
		return fmt.Errorf("snapshotting database: %w", err)
	}
	if err := aws.PutObjectFromFile(config.C.BackupS3BucketName, backupKey(), snapshot, "application/vnd.sqlite3"); err != nil {
		return fmt.Errorf("uploading snapshot: %w", err)
	}
	log.Debug().Msgf("traffic: backed up to s3://%s/%s", config.C.BackupS3BucketName, backupKey())
	return nil
}
