# andrewwillette.com
My personal website hosted at `http://andrewwillette.com`.

## Build

```bash
go build -o andrewwillettedotcom .
```

## CLI Commands

The site's content (audio, sheet music, shows, chord charts) is managed by running these
commands against the live server, rather than by redeploying. All commands are run as
`./andrewwillettedotcom <command>`.

### `serve`
Starts the web server. Set `ENV=PROD` to enable SSL/TLS (autocert) and production config
(`prod.env`); otherwise it runs locally against `nonprod.env`.

### `upload-audio`
Uploads an audio file to S3 so it shows up on `/music`.
- `-f, --file` — path to a single audio file
- `-d, --dir` — directory containing audio file(s) to upload
- `-r, --result-dir` — move the uploaded file here (timestamp-suffixed) after a successful upload

### `delete-audio`
Interactively pick (via fzf) and delete an audio file from S3.

### `upload-sheet-music`
Uploads a sheet-music entry (a Dropbox link, not the PDF itself) to S3 so it shows up on
`/sheet-music`.
- `-n, --name` — display name (e.g. `"Jerusalem Ridge"`)
- `-u, --url` — Dropbox URL to the PDF; if omitted, interactively lists the configured
  Dropbox folder and lets you pick a file via fzf
- `-y, --overwrite` — overwrite an existing entry with the same slug without prompting

### `delete-sheet-music`
Interactively pick (via fzf) and delete a sheet-music entry from S3.

### `upload-show`
Uploads a show (title, date, time, description) to S3 so it shows up on `/shows`.
- `-t, --title` — show title
- `-D, --date` — show date (`YYYY-MM-DD`)
- `-T, --time` — show time (e.g. `8:00pm-10:00pm`)
- `-d, --description` — show description

### `delete-show`
Interactively pick (via fzf) and delete a show entry from S3.

### `generate-images`
Generates per-song cover art by overlaying song titles on the base album image, for every
audio file currently in S3.
- `-u, --upload` — also upload the generated images to S3
- `-k, --keep` — keep the generated images locally after upload

### `dropbox-auth`
One-time interactive setup that walks through Dropbox OAuth to obtain a refresh token
(`DROPBOX_REFRESH_TOKEN`), needed before `upload-sheet-music` can list your Dropbox folder.
- `--app-key` / `--app-secret` — Dropbox app credentials (from the App Console); prompted for
  interactively if omitted

### `sync-chord-charts`
Validates the local `.chord` files and syncs them to S3 so they show up on `/chord-charts`,
without a redeploy — same pattern as `upload-audio`/`upload-sheet-music`/`upload-show`: it
writes straight to S3 using your local AWS credentials, no server involvement needed. Each
`.chord` file (default location: `server/chordcharts/charts/`) has a small header followed by
a `---` separator and the chart body, e.g.:

```
Title: Soldier's Joy
Artist: Traditional (Benton Flippen)
Key: D
---
A part:
1     |  1     |  1     |  5     |
```

Every file is validated locally first (title/key/filename format, and every body line must be
a valid chord line, section label, or blank) — if anything fails, nothing is uploaded. Valid
files are then synced to S3 to match the local directory exactly: charts are added/updated
based on a content checksum (unchanged charts are skipped), and any chart no longer present
locally is deleted from S3. A running server picks up the change automatically — same as it
does for audio/sheet-music/shows — via the SQS poller that watches for S3 bucket events.
- `--dir` — local directory of `.chord` files. Defaults to the `CHORD_CHARTS_DIR` config value
  (a leading `~/` is expanded), then to `server/chordcharts/charts` relative to the current
  directory. Set `CHORD_CHARTS_DIR` to run the command from anywhere.

## Verifying Changes

```bash
./scripts/local-integration-test.sh
```

Builds the app, starts the server, and checks that all endpoints return expected status codes.

## Deploy

```bash
./scripts/deploy-prod-podman.sh
```
