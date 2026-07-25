package aws

import (
	"context"
	"path/filepath"
	"strings"
	"time"

	webCfg "github.com/andrewwillette/andrewwillettedotcom/config"
	"github.com/andrewwillette/andrewwillettedotcom/dropbox"
	"github.com/rs/zerolog/log"
)

const sheetMusicLinkRefreshInterval = 24 * time.Hour

// StartSheetMusicLinkRefreshJob runs the Link Refresh Job once immediately,
// then once every 24h thereafter. It is a no-op if Dropbox isn't configured
// yet (DROPBOX_REFRESH_TOKEN unset).
func StartSheetMusicLinkRefreshJob() {
	dbx := dropbox.NewClientFromConfig()
	if dbx == nil {
		log.Warn().Msg("Dropbox API not configured (DROPBOX_REFRESH_TOKEN unset); sheet music link refresh job disabled")
		return
	}

	go func() {
		RefreshSheetMusicLinks(dbx)
		ticker := time.NewTicker(sheetMusicLinkRefreshInterval)
		defer ticker.Stop()
		for range ticker.C {
			RefreshSheetMusicLinks(dbx)
		}
	}()
}

// RefreshSheetMusicLinks checks every Sheet Music Entry against Dropbox:
//   - entries with a known Dropbox File ID are checked directly; a Confirmed
//     Gone file causes the entry to be deleted, a moved/renamed file gets its
//     link refreshed.
//   - entries missing a Dropbox File ID (uploaded before this job existed) are
//     matched by name against files under webCfg.C.DropboxSheetMusicFolder; an
//     Ambiguous Match (or no match) is skipped and logged, never guessed.
func RefreshSheetMusicLinks(dbx *dropbox.Client) {
	ctx := context.Background()

	rows, err := listSheetJSONRaw()
	if err != nil {
		log.Error().Err(err).Msg("sheet music link refresh: failed to list entries")
		return
	}

	var folderFiles []dropbox.FileMetadata
	var folderListed bool
	changed := false

	for _, row := range rows {
		item := row.JSONItem

		if item.DropboxFileID != "" {
			if refreshKnownEntry(ctx, dbx, row.Key, item) {
				changed = true
			}
			continue
		}

		if !folderListed {
			folderFiles, err = dbx.ListFolder(ctx, webCfg.C.DropboxSheetMusicFolder)
			if err != nil {
				log.Error().Err(err).Str("folder", webCfg.C.DropboxSheetMusicFolder).
					Msg("sheet music link refresh: failed to list dropbox folder; skipping name-matching this run")
			}
			folderListed = true
		}

		if backfillEntry(ctx, dbx, item, folderFiles) {
			changed = true
		}
	}

	if changed {
		UpdateSheetMusicCache()
	}
}

// refreshKnownEntry handles an entry that already has a Dropbox File ID.
// Returns true if the entry was changed (updated or deleted).
func refreshKnownEntry(ctx context.Context, dbx *dropbox.Client, key string, item SheetMusicJSONObject) bool {
	meta, err := dbx.GetMetadata(ctx, item.DropboxFileID)
	if err != nil {
		log.Error().Err(err).Str("entry", item.DisplayName).Msg("sheet music link refresh: failed to check dropbox metadata")
		return false
	}
	if meta == nil {
		log.Warn().Str("entry", item.DisplayName).Msg("sheet music link refresh: backing dropbox file confirmed gone; deleting entry")
		if err := DeleteSheetMusicFromS3(key); err != nil {
			log.Error().Err(err).Str("entry", item.DisplayName).Msg("sheet music link refresh: failed to delete entry for gone file")
			return false
		}
		return true
	}

	freshURL, err := dbx.GetOrCreateSharedLink(ctx, item.DropboxFileID)
	if err != nil {
		log.Error().Err(err).Str("entry", item.DisplayName).Msg("sheet music link refresh: failed to fetch shared link")
		return false
	}
	freshURL = normalizeDropboxURL(freshURL)
	if freshURL == item.DropboxURL {
		return false
	}

	log.Info().Str("entry", item.DisplayName).Str("old_url", item.DropboxURL).Str("new_url", freshURL).
		Msg("sheet music link refresh: updating stale link")
	if err := PutSheetJSON(item.DisplayName, freshURL, item.DropboxFileID); err != nil {
		log.Error().Err(err).Str("entry", item.DisplayName).Msg("sheet music link refresh: failed to save refreshed link")
		return false
	}
	return true
}

// backfillEntry handles a legacy entry with no stored Dropbox File ID, matching
// it by name against files already listed from the configured Dropbox folder.
// Returns true if the entry was changed.
func backfillEntry(ctx context.Context, dbx *dropbox.Client, item SheetMusicJSONObject, folderFiles []dropbox.FileMetadata) bool {
	target := slugify(item.DisplayName)

	var match *dropbox.FileMetadata
	ambiguous := false
	for i := range folderFiles {
		f := folderFiles[i]
		name := strings.TrimSuffix(f.Name, filepath.Ext(f.Name))
		if slugify(name) != target {
			continue
		}
		if match != nil {
			ambiguous = true
			break
		}
		match = &folderFiles[i]
	}

	if ambiguous {
		log.Warn().Str("entry", item.DisplayName).Msg("sheet music link refresh: ambiguous name match against dropbox folder; skipping")
		return false
	}
	if match == nil {
		log.Warn().Str("entry", item.DisplayName).Msg("sheet music link refresh: no matching dropbox file found for entry lacking a file ID; skipping")
		return false
	}

	freshURL, err := dbx.GetOrCreateSharedLink(ctx, match.ID)
	if err != nil {
		log.Error().Err(err).Str("entry", item.DisplayName).Msg("sheet music link refresh: failed to fetch shared link for backfilled match")
		return false
	}
	freshURL = normalizeDropboxURL(freshURL)

	log.Info().Str("entry", item.DisplayName).Str("dropbox_file_id", match.ID).Msg("sheet music link refresh: backfilled dropbox file id")
	if err := PutSheetJSON(item.DisplayName, freshURL, match.ID); err != nil {
		log.Error().Err(err).Str("entry", item.DisplayName).Msg("sheet music link refresh: failed to save backfilled entry")
		return false
	}
	return true
}
