package cmd

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/andrewwillette/andrewwillettedotcom/dropbox"
	"github.com/spf13/cobra"
)

var (
	dropboxAppKeyFlag    string
	dropboxAppSecretFlag string
)

var dropboxAuthCmd = &cobra.Command{
	Use:   "dropbox-auth",
	Short: "One-time interactive setup to obtain a Dropbox refresh token",
	Run: func(cmd *cobra.Command, args []string) {
		if err := runDropboxAuth(); err != nil {
			fmt.Fprintln(os.Stderr, "dropbox-auth failed:", err)
			os.Exit(1)
		}
	},
}

func init() {
	dropboxAuthCmd.Flags().StringVar(&dropboxAppKeyFlag, "app-key", "", "Dropbox app key (from the App Console)")
	dropboxAuthCmd.Flags().StringVar(&dropboxAppSecretFlag, "app-secret", "", "Dropbox app secret (from the App Console)")
	rootCmd.AddCommand(dropboxAuthCmd)
}

func runDropboxAuth() error {
	reader := bufio.NewReader(os.Stdin)

	appKey := strings.TrimSpace(dropboxAppKeyFlag)
	if appKey == "" {
		var err error
		appKey, err = prompt(reader, "Dropbox app key: ")
		if err != nil {
			return err
		}
		appKey = strings.TrimSpace(appKey)
	}

	appSecret := strings.TrimSpace(dropboxAppSecretFlag)
	if appSecret == "" {
		var err error
		appSecret, err = prompt(reader, "Dropbox app secret: ")
		if err != nil {
			return err
		}
		appSecret = strings.TrimSpace(appSecret)
	}

	authorizeURL := dropbox.AuthorizeURL(appKey)
	fmt.Println()
	fmt.Println("Visit this URL, approve access, and copy the code shown:")
	fmt.Println(authorizeURL)
	if copyToClipboard(authorizeURL) {
		fmt.Println("(copied to clipboard)")
	}
	fmt.Println()

	code, err := prompt(reader, "Paste the code here: ")
	if err != nil {
		return err
	}
	code = strings.TrimSpace(code)

	refreshToken, err := dropbox.ExchangeCodeForRefreshToken(context.Background(), appKey, appSecret, code)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("Success. Add these to your app.env / prod.env:")
	fmt.Printf("DROPBOX_APP_KEY=%s\n", appKey)
	fmt.Printf("DROPBOX_APP_SECRET=%s\n", appSecret)
	fmt.Printf("DROPBOX_REFRESH_TOKEN=%s\n", refreshToken)
	return nil
}

// copyToClipboard best-effort copies text to the system clipboard (macOS only).
// Returns false silently if pbcopy isn't available, so callers can still print
// the value for manual copying.
func copyToClipboard(text string) bool {
	cmd := exec.Command("pbcopy")
	cmd.Stdin = bytes.NewBufferString(text)
	return cmd.Run() == nil
}
