package stringutil

import (
	"path/filepath"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Slugify converts a display name to a filesystem/URL-safe slug.
// e.g. "Billy in the Lowground" → "billy_in_the_lowground"
func Slugify(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, "-", "_")
	out := make([]rune, 0, len(s))
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			out = append(out, r)
		}
	}
	s = string(out)
	return strings.Trim(s, "_")
}

// EnsureTrailingSlash appends a "/" to p if it doesn't already end with one.
func EnsureTrailingSlash(p string) string {
	if p == "" || strings.HasSuffix(p, "/") {
		return p
	}
	return p + "/"
}

// FallbackNameFromKey derives a human-readable display name from an S3 key
// when no explicit display name is stored.
func FallbackNameFromKey(key, prefix string) string {
	base := strings.TrimPrefix(key, EnsureTrailingSlash(prefix))
	base = strings.TrimSuffix(base, filepath.Ext(base))
	base = strings.ReplaceAll(base, "_", " ")
	base = strings.ReplaceAll(base, "-", " ")
	c := cases.Title(language.AmericanEnglish)
	return c.String(base)
}
