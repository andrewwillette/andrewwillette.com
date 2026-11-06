package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/andrewwillette/andrewwillettedotcom/server/chordcharts"
	"github.com/spf13/cobra"
)

var (
	formatChordChartsDirFlag string
	formatChordChartsCheck   bool
)

var formatChordChartsCmd = &cobra.Command{
	Use:   "format-chord-charts",
	Short: "Line up the bars in local .chord files",
	Long: `Rewrites each .chord file so the bars on its chord lines line up in
columns, with one space between each bar and the chord after it. Headers,
section labels, and blank lines are left alone.

With --check, nothing is written. The command lists the files that need
formatting and exits with status 1 if there are any.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runFormatChordCharts(); err != nil {
			fmt.Fprintln(os.Stderr, "format-chord-charts failed:", err)
			os.Exit(1)
		}
	},
}

func init() {
	formatChordChartsCmd.Flags().StringVar(&formatChordChartsDirFlag, "dir", "", "Local directory of .chord files (default: CHORD_CHARTS_DIR config, else "+defaultChordChartsDir+")")
	formatChordChartsCmd.Flags().BoolVar(&formatChordChartsCheck, "check", false, "List files that need formatting without changing them; exit 1 if any do")
	rootCmd.AddCommand(formatChordChartsCmd)
}

func runFormatChordCharts() error {
	chordChartsDirFlag = formatChordChartsDirFlag
	dir, err := resolveChordChartsDir()
	if err != nil {
		return err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("reading dir %q (set --dir or CHORD_CHARTS_DIR): %w", dir, err)
	}

	var changed []string
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
		if formatted == string(raw) {
			continue
		}
		changed = append(changed, e.Name())
		if formatChordChartsCheck {
			continue
		}
		if err := os.WriteFile(path, []byte(formatted), 0o644); err != nil {
			return fmt.Errorf("writing %s: %w", e.Name(), err)
		}
	}

	if formatChordChartsCheck {
		for _, name := range changed {
			fmt.Println("needs formatting:", name)
		}
		if len(changed) > 0 {
			return fmt.Errorf("%d chord file(s) need formatting", len(changed))
		}
		return nil
	}
	for _, name := range changed {
		fmt.Println("formatted:", name)
	}
	fmt.Printf("%d chord file(s) formatted\n", len(changed))
	return nil
}
