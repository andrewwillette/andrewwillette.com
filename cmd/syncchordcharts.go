package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/andrewwillette/andrewwillettedotcom/aws"
	"github.com/andrewwillette/andrewwillettedotcom/config"
	"github.com/andrewwillette/andrewwillettedotcom/server/chordcharts"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var chordChartsDirFlag string

const defaultChordChartsDir = "server/chordcharts/charts"

var syncChordChartsCmd = &cobra.Command{
	Use:   "sync-chord-charts",
	Short: "Format and validate local .chord files and sync them to S3",
	Run: func(cmd *cobra.Command, args []string) {
		if err := runSyncChordCharts(); err != nil {
			fmt.Fprintln(os.Stderr, "sync-chord-charts failed:", err)
			os.Exit(1)
		}
	},
}

func init() {
	syncChordChartsCmd.Flags().StringVar(&chordChartsDirFlag, "dir", "", "Local directory of .chord files (default: CHORD_CHARTS_DIR config, else "+defaultChordChartsDir+")")
	rootCmd.AddCommand(syncChordChartsCmd)
}

// resolveChordChartsDir picks the --dir flag, then the CHORD_CHARTS_DIR
// config value, then the repo-relative default, expanding a leading "~/".
func resolveChordChartsDir() (string, error) {
	dir := chordChartsDirFlag
	if dir == "" {
		dir = config.C.ChordChartsDir
	}
	if dir == "" {
		dir = defaultChordChartsDir
	}
	if strings.HasPrefix(dir, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("expanding %q: %w", dir, err)
		}
		dir = filepath.Join(home, dir[2:])
	}
	return dir, nil
}

func runSyncChordCharts() error {
	dir, err := resolveChordChartsDir()
	if err != nil {
		return err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("reading dir %q (set --dir or CHORD_CHARTS_DIR): %w", dir, err)
	}

	var items []aws.ChordChartJSONObject
	var drafts []string
	var validationErrs []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".chord") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		raw, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading %s: %w", e.Name(), err)
		}
		formatted := chordcharts.FormatFile(string(raw))
		if formatted != string(raw) {
			if err := os.WriteFile(path, []byte(formatted), 0o644); err != nil {
				return fmt.Errorf("writing %s: %w", e.Name(), err)
			}
			fmt.Println("formatted:", e.Name())
		}
		parsed, err := chordcharts.ParseChordFile(e.Name(), formatted)
		// A draft is skipped even if the rest of its header is broken, as long
		// as the Draft line comes before the problem.
		if parsed.Draft {
			drafts = append(drafts, e.Name())
			continue
		}
		if err != nil {
			validationErrs = append(validationErrs, fmt.Sprintf("%s: %s", e.Name(), err))
			continue
		}
		if errs := chordcharts.Validate(parsed); len(errs) > 0 {
			for _, ve := range errs {
				validationErrs = append(validationErrs, fmt.Sprintf("%s: %s", e.Name(), ve))
			}
			continue
		}
		items = append(items, aws.ChordChartJSONObject{
			Title:         parsed.Title,
			Artist:        parsed.Artist,
			DefaultKey:    parsed.DefaultKey,
			TimeSignature: parsed.TimeSignature,
			Notes:         parsed.Notes,
			URLVal:        parsed.URLVal(),
			FileName:      parsed.FileName,
			Body:          parsed.Body,
		})
	}

	if len(validationErrs) > 0 {
		for _, e := range validationErrs {
			fmt.Fprintln(os.Stderr, "  -", e)
		}
		return fmt.Errorf("%d chord file(s) failed validation; nothing was synced", len(validationErrs))
	}
	if len(drafts) > 0 {
		fmt.Printf("Skipped drafts: %v\n", drafts)
	}
	if len(items) == 0 {
		return fmt.Errorf("no non-draft .chord files found in %q", dir)
	}

	log.Info().Msgf("Syncing %d chord chart(s) from %s to S3...", len(items), dir)
	result, err := aws.SyncChordCharts(items)
	if err != nil {
		return err
	}

	fmt.Printf("Added:     %v\n", result.Added)
	fmt.Printf("Updated:   %v\n", result.Updated)
	fmt.Printf("Unchanged: %v\n", result.Unchanged)
	fmt.Printf("Deleted:   %v\n", result.Deleted)
	return nil
}
