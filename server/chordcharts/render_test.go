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
		{"slash chords", "vi/V | I/III | IVmaj7/bVII | ii7/#IV |", true},
		{"lowercase slash bass rejected", "vi/v |", false},
		{"slash with quality bass rejected", "I/V7 |", false},
		{"empty slash bass rejected", "vi/ |", false},
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
		token, degree, quality, bass string
	}{
		{"I", "1", "", ""},
		{"V7", "5", "7", ""},
		{"ii", "2", "m", ""},
		{"ii7", "2", "m7", ""},
		{"iv6", "4", "m6", ""},
		{"IVmaj7", "4", "maj7", ""},
		{"bVII", "b7", "", ""},
		{"#iv", "#4", "m", ""},
		{"viidim", "7", "dim", ""},
		{"viim7b5", "7", "m7b5", ""},
		{"Vsus4", "5", "sus4", ""},
		{"vi/V", "6", "m", "5"},
		{"I/III", "1", "", "3"},
		{"ii7/bVII", "2", "m7", "b7"},
	}
	for _, c := range cases {
		t.Run(c.token, func(t *testing.T) {
			got, ok := parseChord(c.token)
			if !ok {
				t.Fatalf("parseChord(%q) rejected", c.token)
			}
			if got.Degree != c.degree || got.Quality != c.quality || got.Bass != c.bass {
				t.Errorf("parseChord(%q) = %+v, want degree %q quality %q bass %q", c.token, got, c.degree, c.quality, c.bass)
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

func TestRenderBodyAddsSlashBass(t *testing.T) {
	html := string(RenderBody("vi/V  |  I  |"))
	if !strings.Contains(html, `data-numeral="vi/V" data-degree="6" data-quality="m" data-bass="5" data-width="6">vi/V  </span>|`) {
		t.Errorf("expected slash chord span with bass, got: %s", html)
	}
	if strings.Count(html, "data-bass") != 1 {
		t.Errorf("expected data-bass only on the slash chord, got: %s", html)
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

func TestParseModulation(t *testing.T) {
	cases := []struct {
		label, degree string
		ok, wantErr   bool
	}{
		{"B (-> IV):", "4", true, false},
		{"B part (->bVII):", "b7", true, false},
		{"Bridge ( -> #IV ):", "#4", true, false},
		{"A part:", "", false, false},
		{"B (-> iv):", "", true, true},
		{"B (-> IV7):", "", true, true},
		{"B (-> D):", "", true, true},
		{"B -> IV:", "", true, true},
	}
	for _, c := range cases {
		t.Run(c.label, func(t *testing.T) {
			degree, _, ok, err := parseModulation(c.label)
			if ok != c.ok || (err != nil) != c.wantErr || degree != c.degree {
				t.Errorf("parseModulation(%q) = %q, %v, %v; want %q, %v, err %v", c.label, degree, ok, err, c.degree, c.ok, c.wantErr)
			}
		})
	}
}

func TestRenderBodyKeyChangeAppliesToItsSection(t *testing.T) {
	body := "A:\nI  |\n\nB (-> IV):\nI  |  V  |\n\nC:\nI  |"
	html := string(RenderBody(body))
	if !strings.Contains(html, `B (-&gt; <span class="modulation" data-numeral="IV" data-degree="4">IV</span>):`) {
		t.Errorf("expected modulation span in label, got: %s", html)
	}
	lines := strings.Split(html, "\n")
	if strings.Contains(lines[1], "data-tonic") {
		t.Errorf("expected no tonic before the key change, got: %s", lines[1])
	}
	if strings.Count(lines[4], `data-tonic="4"`) != 2 {
		t.Errorf("expected both chords after the key change to have tonic 4, got: %s", lines[4])
	}
	if strings.Contains(lines[7], "data-tonic") {
		t.Errorf("expected the next section label to return to the home key, got: %s", lines[7])
	}
	text := regexp.MustCompile(`<[^>]+>`).ReplaceAllString(html, "")
	if text != strings.ReplaceAll(body, ">", "&gt;") {
		t.Errorf("rendered text changed:\n got %q\nwant %q", text, body)
	}
}
