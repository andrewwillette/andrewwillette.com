package cmd

import (
	"bufio"
	"context"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/andrewwillette/andrewwillettedotcom/aws"
	webCfg "github.com/andrewwillette/andrewwillettedotcom/config"
	"github.com/andrewwillette/andrewwillettedotcom/dropbox"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	sheetNameFlag     string
	sheetURLFlag      string
	sheetOverwriteFlg bool
)

var uploadSheetMusicCmd = &cobra.Command{
	Use:   "upload-sheet-music",
	Short: "Upload a Dropbox sheet-music link (JSON) to S3",
	Run: func(cmd *cobra.Command, args []string) {
		if err := runUploadSheetMusic(); err != nil {
			log.Fatal().Err(err).Msg("upload-sheet-music failed")
		}
	},
}

func init() {
	uploadSheetMusicCmd.Flags().StringVarP(&sheetNameFlag, "name", "n", "", "Display name (e.g., \"Jerusalem Ridge\")")
	uploadSheetMusicCmd.Flags().StringVarP(&sheetURLFlag, "url", "u", "", "Dropbox URL to the PDF")
	uploadSheetMusicCmd.Flags().BoolVarP(&sheetOverwriteFlg, "overwrite", "y", false, "Overwrite if an entry with the same slug already exists")
	rootCmd.AddCommand(uploadSheetMusicCmd)
}

func runUploadSheetMusic() error {
	reader := bufio.NewReader(os.Stdin)

	dropboxURL := strings.TrimSpace(sheetURLFlag)
	if dropboxURL == "" {
		var err error
		dropboxURL, err = prompt(reader, "Dropbox URL: ")
		if err != nil {
			return err
		}
		dropboxURL = strings.TrimSpace(dropboxURL)
	}
	if err := validateDropboxURL(dropboxURL); err != nil {
		return fmt.Errorf("invalid dropbox url: %w", err)
	}

	displayName := strings.TrimSpace(sheetNameFlag)
	if displayName == "" {
		def := defaultNameFromURL(dropboxURL)
		var err error
		displayName, err = promptDefault(reader, "Display name", def)
		if err != nil {
			return err
		}
		displayName = strings.TrimSpace(displayName)
		if displayName == "" {
			displayName = def
		}
	}

	slug := slugify(displayName)
	key := ensureTrailingSlash(webCfg.C.SheetMusicS3BucketPrefix) + slug + ".json"

	exists, err := keyExistsInS3(key)
	if err != nil {
		return err
	}
	if exists && !sheetOverwriteFlg {
		ans, err := promptDefault(reader, fmt.Sprintf("Entry exists at %s; overwrite?", key), "n")
		if err != nil {
			return err
		}
		if strings.ToLower(strings.TrimSpace(ans)) != "y" {
			log.Info().Msg("Aborted by user")
			return nil
		}
	}

	fileID := resolveDropboxFileID(dropboxURL)

	log.Info().Msgf("Uploading sheet JSON: name=%q url=%q fileID=%q (key=%s)", displayName, dropboxURL, fileID, key)
	if err := aws.PutSheetJSON(displayName, dropboxURL, fileID); err != nil {
		return err
	}
	log.Info().Msg("Upload complete")
	return nil
}

// resolveDropboxFileID looks up the stable Dropbox file ID backing a shared
// URL, so the link refresh job can find this file later even if it's renamed
// or moved. Returns "" (logging a warning) if Dropbox isn't configured yet or
// the lookup fails — the entry is still saved, just without that safety net.
func resolveDropboxFileID(dropboxURL string) string {
	dbx := dropbox.NewClientFromConfig()
	if dbx == nil {
		log.Warn().Msg("Dropbox API not configured (DROPBOX_REFRESH_TOKEN unset); saving entry without a file ID, so it won't be covered by the automatic link refresh job")
		return ""
	}
	fileID, err := dbx.ResolveSharedLinkFileID(context.Background(), dropboxURL)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to resolve Dropbox file ID for this URL; saving entry without one")
		return ""
	}
	return fileID
}

func prompt(r *bufio.Reader, label string) (string, error) {
	fmt.Fprint(os.Stdout, label)
	s, err := r.ReadString('\n')
	return strings.TrimRight(s, "\r\n"), err
}

func promptDefault(r *bufio.Reader, label, def string) (string, error) {
	p := fmt.Sprintf("%s [%s]: ", label, def)
	fmt.Fprint(os.Stdout, p)
	s, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	s = strings.TrimRight(s, "\r\n")
	if s == "" {
		return def, nil
	}
	return s, nil
}

func validateDropboxURL(u string) error {
	log.Debug().Msgf("Validating dropbox URL: %s", u)
	parsed, err := url.Parse(u)
	if err != nil {
		return err
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return fmt.Errorf("url scheme must be http/https, scheme identified as %v", parsed.Scheme)
	}
	host := strings.ToLower(parsed.Host)
	if !strings.Contains(host, "dropbox.com") {
		log.Warn().Msg("URL host is not dropbox.com; continuing anyway")
	}
	if strings.TrimSpace(parsed.Path) == "" || parsed.Path == "/" {
		return fmt.Errorf("url path looks empty")
	}
	return nil
}

func defaultNameFromURL(u string) string {
	parsed, err := url.Parse(u)
	if err != nil {
		return "Untitled"
	}
	base := path.Base(parsed.Path)
	base = strings.TrimSuffix(base, filepath.Ext(base))
	base = strings.ReplaceAll(base, "_", " ")
	base = strings.ReplaceAll(base, "-", " ")
	base = strings.ReplaceAll(base, "%20", " ")
	base = strings.TrimSpace(base)
	if base == "" {
		return "Untitled"
	}
	return toTitleWords(base)
}

func toTitleWords(s string) string {
	parts := strings.Fields(s)
	for i := range parts {
		if len(parts[i]) == 0 {
			continue
		}
		runes := []rune(parts[i])
		runes[0] = []rune(strings.ToUpper(string(runes[0])))[0]
		parts[i] = string(runes)
	}
	return strings.Join(parts, " ")
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
	return strings.Trim(string(out), "_")
}

func ensureTrailingSlash(p string) string {
	if p == "" || strings.HasSuffix(p, "/") {
		return p
	}
	return p + "/"
}

func keyExistsInS3(fullKey string) (bool, error) {
	objs, err := aws.ListSheetMusicObjects()
	if err != nil {
		return false, err
	}
	for _, o := range objs {
		if o.Key == fullKey {
			return true, nil
		}
	}
	return false, nil
}
