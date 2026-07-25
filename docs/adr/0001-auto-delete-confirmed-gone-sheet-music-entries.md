# Link Refresh Job auto-deletes Sheet Music Entries that are Confirmed Gone

The Link Refresh Job runs unattended and can repair Stale Links using Dropbox's API. We decided that when it looks up a Sheet Music Entry's Dropbox File ID and Dropbox reports the file no longer exists (Confirmed Gone, not an Ambiguous Match), the job deletes the entry from S3 immediately, with no human confirmation step.

This was a deliberate choice over the safer alternative (leave the entry and just log/alert), made after weighing that surfacing every automated decision for review would undercut the point of the job running unattended at all. The trade-off: if a file is ever moved somewhere the job can't see it as the *same* file (rather than genuinely deleted), the corresponding tune silently disappears from the public site with no confirmation prompt — recoverable only by noticing the log line and re-uploading.
