package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/andrewwillette/andrewwillettedotcom/dropbox"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: fetchpdf <dropbox-path-or-shared-url> <out-file>")
		os.Exit(1)
	}
	dbx := dropbox.NewClientFromConfig()
	if dbx == nil {
		fmt.Fprintln(os.Stderr, "dropbox not configured")
		os.Exit(1)
	}

	ctx := context.Background()
	target := os.Args[1]
	if strings.Contains(target, "dropbox.com/") {
		fileID, err := dbx.ResolveSharedLinkFileID(ctx, target)
		if err != nil {
			fmt.Fprintln(os.Stderr, "resolving shared link failed:", err)
			os.Exit(1)
		}
		target = fileID
	}

	content, err := dbx.Download(ctx, target)
	if err != nil {
		fmt.Fprintln(os.Stderr, "download failed:", err)
		os.Exit(1)
	}
	if err := os.WriteFile(os.Args[2], content, 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "write failed:", err)
		os.Exit(1)
	}
	fmt.Printf("wrote %d bytes to %s\n", len(content), os.Args[2])
}
