package chordcharts

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const validChordFile = `Title: Soldier's Joy
Artist: Traditional (Benton Flippen)
Key: D
---
A part:
I     |  I     |  I     |  V7    |
I     |  I     |  V7    |  I     |
`

func TestParseChordFile(t *testing.T) {
	parsed, err := ParseChordFile("soldiers_joy.chord", validChordFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.Title != "Soldier's Joy" {
		t.Errorf("Title = %q", parsed.Title)
	}
	if parsed.Artist != "Traditional (Benton Flippen)" {
		t.Errorf("Artist = %q", parsed.Artist)
	}
	if parsed.DefaultKey != "D" {
		t.Errorf("DefaultKey = %q", parsed.DefaultKey)
	}
	if parsed.URLVal() != "soldiers_joy" {
		t.Errorf("URLVal() = %q", parsed.URLVal())
	}
	if !strings.Contains(parsed.Body, "A part:") {
		t.Errorf("Body missing content: %q", parsed.Body)
	}
	if strings.Contains(parsed.Body, "---") {
		t.Errorf("Body should not contain the header separator: %q", parsed.Body)
	}
}

func TestParseChordFileMissingSeparator(t *testing.T) {
	_, err := ParseChordFile("bad.chord", "Title: X\nKey: D\nI | I |\n")
	if err == nil {
		t.Fatal("expected error for missing '---' separator")
	}
}

func TestParseChordFileMalformedHeaderLine(t *testing.T) {
	_, err := ParseChordFile("bad.chord", "Title X\n---\nI | I |\n")
	if err == nil {
		t.Fatal("expected error for malformed header line")
	}
}

func TestValidateAcceptsWellFormedChart(t *testing.T) {
	parsed, err := ParseChordFile("soldiers_joy.chord", validChordFile)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if errs := Validate(parsed); len(errs) != 0 {
		t.Errorf("expected no validation errors, got: %v", errs)
	}
}

func TestValidateRejectsProblems(t *testing.T) {
	cases := []struct {
		name string
		p    ParsedChordFile
	}{
		{"missing title", ParsedChordFile{FileName: "x.chord", DefaultKey: "D", Body: "I | I |"}},
		{"bad extension", ParsedChordFile{FileName: "x.txt", Title: "X", DefaultKey: "D", Body: "I | I |"}},
		{"bad slug chars", ParsedChordFile{FileName: "My Song.chord", Title: "X", DefaultKey: "D", Body: "I | I |"}},
		{"bad key", ParsedChordFile{FileName: "x.chord", Title: "X", DefaultKey: "H", Body: "I | I |"}},
		{"empty body", ParsedChordFile{FileName: "x.chord", Title: "X", DefaultKey: "D", Body: ""}},
		{"invalid chord line", ParsedChordFile{FileName: "x.chord", Title: "X", DefaultKey: "D", Body: "this is not a chord line at all"}},
		{"old arabic numerals", ParsedChordFile{FileName: "x.chord", Title: "X", DefaultKey: "D", Body: "1 | 5 |"}},
		{"bad header time signature", ParsedChordFile{FileName: "x.chord", Title: "X", DefaultKey: "D", TimeSignature: "waltz", Body: "I | I |"}},
		{"lowercase key change", ParsedChordFile{FileName: "x.chord", Title: "X", DefaultKey: "D", Body: "B (-> iv):\nI | I |"}},
		{"malformed key change", ParsedChordFile{FileName: "x.chord", Title: "X", DefaultKey: "D", Body: "B -> IV:\nI | I |"}},
		{"mid-measure time signature", ParsedChordFile{FileName: "x.chord", Title: "X", DefaultKey: "D", Body: "I 2/4 V | I |"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if errs := Validate(c.p); len(errs) == 0 {
				t.Errorf("expected validation errors for case %q, got none", c.name)
			}
		})
	}
}

func TestParseChordFileTimeSignature(t *testing.T) {
	raw := "Title: X\nKey: G\nTime: 3/4\n---\nI  |  2/4 V  |  I  |\n"
	parsed, err := ParseChordFile("x.chord", raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.TimeSignature != "3/4" {
		t.Errorf("TimeSignature = %q", parsed.TimeSignature)
	}
	if errs := Validate(parsed); len(errs) != 0 {
		t.Errorf("expected no validation errors, got: %v", errs)
	}
}

func TestValidateAllowsSectionLabelsAndBlankLines(t *testing.T) {
	p := ParsedChordFile{
		FileName:   "x.chord",
		Title:      "X",
		DefaultKey: "Bb",
		Body:       "A part:\nI     |  vi    |\n\nB part (-> IV):\nIV    |  V7    |",
	}
	if errs := Validate(p); len(errs) != 0 {
		t.Errorf("expected no validation errors, got: %v", errs)
	}
}

func TestChartsDirectoryFilesAreValid(t *testing.T) {
	paths, err := filepath.Glob("charts/*.chord")
	if err != nil || len(paths) == 0 {
		t.Fatalf("no chord files found: %v", err)
	}
	for _, path := range paths {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		parsed, err := ParseChordFile(filepath.Base(path), string(raw))
		if err != nil {
			t.Errorf("%s: %v", path, err)
			continue
		}
		if errs := Validate(parsed); len(errs) != 0 {
			t.Errorf("%s: %v", path, errs)
		}
	}
}
