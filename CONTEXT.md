# andrewwillette.com

Personal website: blog, audio recordings, sheet music, and admin/traffic tracking, backed by S3/SQS and served via Echo.

## Language

**Sheet Music Entry**:
A single tune's listing on the `/sheet-music` page — a display name plus a link to the PDF, persisted as one JSON object in S3.
_Avoid_: sheet, record, item

**Dropbox File ID**:
The stable identifier (or path) of a Sheet Music Entry's backing file in Dropbox, stored alongside the display link so the entry can be relocated even after the file is renamed or moved.
_Avoid_: match key, dropbox path (when referring to the stored identifier specifically)

**Stale Link**:
A Sheet Music Entry whose stored Dropbox link no longer resolves to the correct file — detected by asking Dropbox about the entry's Dropbox File ID, not by fetching the link itself.
_Avoid_: broken link, dead link

**Link Refresh Job**:
The daily background process (alongside the existing SQS-driven cache refresher) that checks every Sheet Music Entry against Dropbox and repairs Stale Links automatically.
_Avoid_: link checker, dropbox sync

**Ambiguous Match**:
The outcome when a Sheet Music Entry has no Dropbox File ID yet and name-matching against Dropbox finds more than one plausible candidate file. The Link Refresh Job never guesses in this case — it skips the entry and logs a warning.
_Avoid_: conflict, uncertain match

**Confirmed Gone**:
The outcome when the Link Refresh Job looks up a Sheet Music Entry's known Dropbox File ID and Dropbox reports the file no longer exists. Distinct from an Ambiguous Match (which means "we don't know," not "it's gone") — only a Confirmed Gone entry is auto-deleted.
_Avoid_: deleted, not found
