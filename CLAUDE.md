# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Run Commands

```bash
# Build the binary
go build -o andrewwillettedotcom .
```

## CLI Commands

The application uses Cobra for CLI. Available subcommands:
- `serve` - Start the web server (uses `ENV=PROD` for SSL/TLS)
- `upload-audio -f <file>` or `-d <dir>` - Upload audio to S3
- `delete-audio` - Interactive deletion of audio from S3
- `upload-sheetmusic` - Upload sheet music to S3
- `delete-sheetmusic` - Interactive deletion of sheet music from S3

## Architecture

### Server (Echo Framework)
The web server uses Echo v4 with Go templates. Entry point is `server/server.go:StartServer()`.

Routes:
- `/` - Homepage
- `/music` - Audio recordings page (served from S3 with presigned URLs)
- `/sheet-music` - Sheet music page (PDFs from S3)
- `/blog`, `/blog/:blog`, `/blog/rss` - Blog system with RSS
- `/key-of-the-day` - Daily musical key feature
- `/resume` - Redirects to S3-hosted resume

Templates are in `server/templates/*.tmpl` with shared header/footer partials.

### AWS Integration
- **S3**: Stores audio files, sheet music PDFs, and images. Uses presigned URLs with 60-minute expiry.
- **SQS**: Polls for S3 events to trigger cache updates when audio/sheet music changes.
- Audio and sheet music are cached in memory with periodic refresh before presigned URLs expire.

### Blog System
Blogs are defined in `server/blog/blog.go` with markdown files in `server/blog/posts/`. Adding a new blog requires adding an entry to the `uninitializedBlogs` slice.

### Configuration
Uses Viper with `.env` files. Looks for `app.env` (dev) or `prod.env` (when `ENV=PROD`). Config can also be in `~/.config/andrewwillette.com/`.

Key config values: S3 bucket names/prefixes, SQS URL, logging settings, pprof toggle.
