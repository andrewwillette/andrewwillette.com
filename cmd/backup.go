package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/andrewwillette/andrewwillettedotcom/backup"
	"github.com/andrewwillette/andrewwillettedotcom/config"
	"github.com/spf13/cobra"
)

var (
	backupOutFlag     string
	backupBucketsFlag []string
)

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Mirror all S3 content (including the server's traffic.db snapshot) into a local SQLite file",
	Run: func(cmd *cobra.Command, args []string) {
		if err := runBackup(); err != nil {
			fmt.Fprintln(os.Stderr, "backup failed:", err)
			os.Exit(1)
		}
	},
}

func init() {
	backupCmd.Flags().StringVarP(&backupOutFlag, "out", "o", "", "Archive file (default ~/andrewwillette-backup.db)")
	backupCmd.Flags().StringSliceVarP(&backupBucketsFlag, "bucket", "b", nil, "Bucket(s) to archive (default: every bucket in config)")
	rootCmd.AddCommand(backupCmd)
}

func runBackup() error {
	out := backupOutFlag
	if out == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		out = filepath.Join(home, "andrewwillette-backup.db")
	}

	buckets := backupBucketsFlag
	if len(buckets) == 0 {
		buckets = configuredBuckets()
	}
	if len(buckets) == 0 {
		return fmt.Errorf("no buckets configured; pass --bucket")
	}

	fmt.Printf("Archiving %v into %s\n", buckets, out)
	results, err := backup.Sync(out, buckets)
	for _, r := range results {
		fmt.Printf("%s: added %d, updated %d, unchanged %d, gone from S3 %d\n",
			r.Bucket, r.Added, r.Updated, r.Unchanged, r.MarkedGone)
	}
	return err
}

// configuredBuckets returns the distinct non-empty bucket names from config.
func configuredBuckets() []string {
	var buckets []string
	seen := map[string]bool{}
	for _, b := range []string{
		config.C.AudioS3BucketName,
		config.C.SheetMusicS3BucketName,
		config.C.ShowsS3BucketName,
		config.C.ChordChartsS3BucketName,
		config.C.BackupS3BucketName,
	} {
		if b != "" && !seen[b] {
			seen[b] = true
			buckets = append(buckets, b)
		}
	}
	return buckets
}
