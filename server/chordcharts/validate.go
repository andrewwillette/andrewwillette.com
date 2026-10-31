package chordcharts

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	keyRe  = regexp.MustCompile(`^[A-G](#|b)?m?$`)
	slugRe = regexp.MustCompile(`^[a-z0-9_]+$`)
)

// ParsedChordFile is the local, authored representation of a .chord file: a
// small header block (Title/Artist/Key/Time) followed by a "---" separator and
// the chart body, e.g.:
//
//	Title: Soldier's Joy
//	Artist: Traditional (Benton Flippen)
//	Key: D
//	---
//	A part:
//	1     |  1     |  1     |  5     |
type ParsedChordFile struct {
	FileName   string
	Title      string
	Artist     string
	DefaultKey string
	// TimeSignature is the chart-wide meter, e.g. "3/4". Empty means 4/4.
	TimeSignature string
	Body          string
}

// URLVal derives the URL slug for this chart from its filename.
func (p ParsedChordFile) URLVal() string {
	return strings.TrimSuffix(p.FileName, filepath.Ext(p.FileName))
}

// ParseChordFile parses the header+body .chord file format described above.
func ParseChordFile(fileName string, raw string) (ParsedChordFile, error) {
	parsed := ParsedChordFile{FileName: fileName}

	lines := strings.Split(raw, "\n")
	sepIdx := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == "---" {
			sepIdx = i
			break
		}
	}
	if sepIdx == -1 {
		return parsed, fmt.Errorf("missing '---' header separator")
	}

	for _, line := range lines[:sepIdx] {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		key, val, ok := strings.Cut(line, ":")
		if !ok {
			return parsed, fmt.Errorf("malformed header line %q (want \"Key: value\")", line)
		}
		val = strings.TrimSpace(val)
		switch strings.ToLower(strings.TrimSpace(key)) {
		case "title":
			parsed.Title = val
		case "artist":
			parsed.Artist = val
		case "key":
			parsed.DefaultKey = val
		case "time":
			parsed.TimeSignature = val
		default:
			return parsed, fmt.Errorf("unknown header field %q", key)
		}
	}

	parsed.Body = strings.Trim(strings.Join(lines[sepIdx+1:], "\n"), "\n")
	return parsed, nil
}

// Validate checks a parsed chord file for problems that would make it unsafe
// or broken to publish, returning every problem found (not just the first)
// so a sync can report them all at once.
func Validate(p ParsedChordFile) []string {
	var errs []string

	if strings.TrimSpace(p.Title) == "" {
		errs = append(errs, "title is required")
	}
	if !strings.HasSuffix(strings.ToLower(p.FileName), ".chord") {
		errs = append(errs, "file name must end in .chord")
	}

	slug := p.URLVal()
	if slug == "" || !slugRe.MatchString(slug) {
		errs = append(errs, fmt.Sprintf("file name %q must yield a URL slug of lowercase letters, digits, and underscores only", p.FileName))
	}

	if !keyRe.MatchString(strings.TrimSpace(p.DefaultKey)) {
		errs = append(errs, fmt.Sprintf("key %q is not a recognized musical key (expected like D, Bb, F#, or Am)", p.DefaultKey))
	}

	if p.TimeSignature != "" && !timeSigRe.MatchString(p.TimeSignature) {
		errs = append(errs, fmt.Sprintf("time %q is not a recognized time signature (expected like 4/4, 3/4, or 6/8)", p.TimeSignature))
	}

	if strings.TrimSpace(p.Body) == "" {
		errs = append(errs, "body is empty")
	}
	for i, line := range strings.Split(p.Body, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasSuffix(trimmed, ":") {
			continue // blank line or section label, e.g. "A part:"
		}
		if err := chordLineError(line); err != nil {
			errs = append(errs, fmt.Sprintf("line %d is not a valid chord line, section label, or blank line (%s): %q", i+1, err, line))
		}
	}

	return errs
}
