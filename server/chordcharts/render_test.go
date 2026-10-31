package chordcharts

import (
	"regexp"
	"strings"
	"testing"
)

func TestIsChordLine(t *testing.T) {
	cases := []struct {
		name string
		line string
		want bool
	}{
		{"chord line with bars", "I     |  IV    |  I     |", true},
		{"chord line with minor and flat", "vi    |  bVII  |", true},
		{"two chords in one bar", "I  V7  |  I     |", true},
		{"sevenths and qualities", "I7 ii7 IVmaj7 viidim #ivm7b5 Vsus4 bVI6", true},
		{"arabic numerals rejected", "1     |  5     |", false},
		{"mixed-case numeral rejected", "Iv    |", false},
		{"minor sus rejected", "iisus4 |", false},
		{"old m suffix rejected", "vim   |", false},
		{"lyric line", "Amazing grace, how sweet the sound,", false},
		{"section label", "Verse:", false},
		{"blank line", "", false},
		{"whitespace only line", "   ", false},
		{"mixed chord and lyric text", "I Amazing", false},
		{"time signature starting a measure", "I     |  2/4 V  |  I     |", true},
		{"time signature starting the line", "3/4 I  |  I     |", true},
		{"time signature mid-measure rejected", "I  2/4 V  |", false},
		{"two time signatures in one measure rejected", "|  2/4 3/4 V  |", false},
		{"bad time signature denominator rejected", "|  3/5 V  |", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := isChordLine(c.line); got != c.want {
				t.Errorf("isChordLine(%q) = %v, want %v", c.line, got, c.want)
			}
		})
	}
}

func TestParseChord(t *testing.T) {
	cases := []struct {
		token, degree, quality string
	}{
		{"I", "1", ""},
		{"V7", "5", "7"},
		{"ii", "2", "m"},
		{"ii7", "2", "m7"},
		{"iv6", "4", "m6"},
		{"IVmaj7", "4", "maj7"},
		{"bVII", "b7", ""},
		{"#iv", "#4", "m"},
		{"viidim", "7", "dim"},
		{"viim7b5", "7", "m7b5"},
		{"Vsus4", "5", "sus4"},
	}
	for _, c := range cases {
		t.Run(c.token, func(t *testing.T) {
			got, ok := parseChord(c.token)
			if !ok {
				t.Fatalf("parseChord(%q) rejected", c.token)
			}
			if got.Degree != c.degree || got.Quality != c.quality {
				t.Errorf("parseChord(%q) = %+v, want degree %q quality %q", c.token, got, c.degree, c.quality)
			}
		})
	}
}

func TestRenderBodyWrapsChordTokens(t *testing.T) {
	html := string(RenderBody("I     |  vi    |  V7  I  |\nAmazing grace"))
	if !strings.Contains(html, `data-numeral="I" data-degree="1" data-quality="" data-width="6">I     </span>|`) {
		t.Errorf("expected degree 1 span, got: %s", html)
	}
	if !strings.Contains(html, `data-numeral="vi" data-degree="6" data-quality="m"`) {
		t.Errorf("expected degree 6 minor span, got: %s", html)
	}
	if !strings.Contains(html, `data-numeral="V7" data-degree="5" data-quality="7" data-width="4">V7  </span>`) {
		t.Errorf("expected degree 5 dominant seventh span, got: %s", html)
	}
	if !strings.Contains(html, "Amazing grace") {
		t.Errorf("expected lyric line to pass through, got: %s", html)
	}
}

func TestRenderBodyEscapesHTML(t *testing.T) {
	html := string(RenderBody("Some <script>alert(1)</script> lyric"))
	if strings.Contains(html, "<script>") {
		t.Errorf("expected lyric line to be HTML-escaped, got: %s", html)
	}
}

func TestRenderBodyPreservesLineText(t *testing.T) {
	line := "I     |  IV    |  I  V7  |  bVI"
	html := string(RenderBody(line))
	text := regexp.MustCompile(`<[^>]+>`).ReplaceAllString(html, "")
	if text != line {
		t.Errorf("rendered text changed:\n got %q\nwant %q", text, line)
	}
}

func TestRenderBodyWrapsTimeSignatures(t *testing.T) {
	line := "I     |  2/4 V  |  I     |"
	html := string(RenderBody(line))
	if !strings.Contains(html, `<span class="timesig">2/4</span>`) {
		t.Errorf("expected timesig span, got: %s", html)
	}
	if !strings.Contains(html, `data-numeral="V"`) {
		t.Errorf("expected chord after time signature to render, got: %s", html)
	}
	text := regexp.MustCompile(`<[^>]+>`).ReplaceAllString(html, "")
	if text != line {
		t.Errorf("rendered text changed:\n got %q\nwant %q", text, line)
	}
}
