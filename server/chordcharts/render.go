package chordcharts

import (
	"fmt"
	"html/template"
	"regexp"
	"strings"
)

var (
	// Roman-numeral chord token: optional accidental, numeral (uppercase =
	// major, lowercase = minor), optional quality suffix. Longer
	// alternatives must precede their prefixes (VII before VI before V, ...).
	chordTokenRe = regexp.MustCompile(`^([b#]?)(VII|VI|V|IV|III|II|I|vii|vi|v|iv|iii|ii|i)((?:maj7|m7b5|dim7|sus4|sus2|add9|aug|dim|6|7|9)?)$`)
	barTokenRe   = regexp.MustCompile(`^\|+$`)
	// Time signature, e.g. "3/4" or "6/8". In the header it sets the whole
	// chart's meter; inline it applies to the one measure it starts.
	timeSigRe = regexp.MustCompile(`^[1-9][0-9]?/(1|2|4|8|16)$`)
	fieldRe   = regexp.MustCompile(`\S+`)

	romanDegrees = map[string]int{"I": 1, "II": 2, "III": 3, "IV": 4, "V": 5, "VI": 6, "VII": 7}
)

// chord is a parsed Roman-numeral token, e.g. "bVII7" or "ii7".
type chord struct {
	Degree  string // accidental + scale degree as a digit, e.g. "b7"
	Quality string // letter-chord suffix, e.g. "", "m", "m7", "7", "dim"
}

// parseChord parses a Roman-numeral chord token. Lowercase numerals are
// minor, so "ii7" means a minor seventh (m7) on the 2nd degree. Lowercase
// numerals can't take aug/sus suffixes, which contradict a minor third.
func parseChord(token string) (chord, bool) {
	m := chordTokenRe.FindStringSubmatch(token)
	if m == nil {
		return chord{}, false
	}
	accidental, numeral, suffix := m[1], m[2], m[3]
	upper := strings.ToUpper(numeral)
	minor := numeral != upper
	quality := suffix
	if minor {
		switch suffix {
		case "aug", "sus2", "sus4":
			return chord{}, false
		case "dim", "dim7", "m7b5":
			// already imply a minor third
		default:
			quality = "m" + suffix
		}
	}
	return chord{Degree: fmt.Sprintf("%s%d", accidental, romanDegrees[upper]), Quality: quality}, true
}

// chordLineError reports why a line isn't a chord line, or nil if it is. A
// chord line is made only of bar-runs ("|", "||", ...), Roman-numeral chord
// tokens, and time signatures. A time signature applies to a single measure,
// so it must be the first token of that measure: at the start of the line or
// right after a bar.
func chordLineError(line string) error {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return fmt.Errorf("line is blank")
	}
	measureStart := true
	for _, f := range fields {
		switch {
		case barTokenRe.MatchString(f):
			measureStart = true
		case timeSigRe.MatchString(f):
			if !measureStart {
				return fmt.Errorf("time signature %q must be the first token in its measure", f)
			}
			measureStart = false
		default:
			if _, ok := parseChord(f); !ok {
				return fmt.Errorf("unrecognized token %q", f)
			}
			measureStart = false
		}
	}
	return nil
}

// isChordLine reports whether the line is a chord line (see chordLineError).
// Blank lines and lines with any non-matching field (lyrics, section labels)
// are "plain" lines, rendered verbatim.
func isChordLine(line string) bool {
	return chordLineError(line) == nil
}

// renderChordLine wraps each numeral token in a <span class="chord"> while
// preserving exact original whitespace/column alignment and bar characters.
// The span carries the original numeral (shown in "Numerals" mode) plus the
// degree and letter-chord quality the key picker uses to spell letter chords.
//
// Each chord span also swallows the spaces after it and records the total
// column width in data-width, so the key picker can pad a converted label
// (e.g. "IV" -> "C") back to that width and keep bars lined up.
func renderChordLine(line string) string {
	var sb strings.Builder
	last := 0
	for _, loc := range fieldRe.FindAllStringIndex(line, -1) {
		start, end := loc[0], loc[1]
		if start < last {
			continue // whitespace already consumed by the previous chord span
		}
		sb.WriteString(template.HTMLEscapeString(line[last:start]))
		token := line[start:end]
		if c, ok := parseChord(token); ok {
			cellEnd := end
			for cellEnd < len(line) && line[cellEnd] == ' ' {
				cellEnd++
			}
			cell := line[start:cellEnd]
			fmt.Fprintf(&sb, `<span class="chord" data-numeral="%s" data-degree="%s" data-quality="%s" data-width="%d">%s</span>`,
				template.HTMLEscapeString(token),
				template.HTMLEscapeString(c.Degree),
				template.HTMLEscapeString(c.Quality),
				len(cell),
				template.HTMLEscapeString(cell))
			end = cellEnd
		} else if timeSigRe.MatchString(token) {
			fmt.Fprintf(&sb, `<span class="timesig">%s</span>`, template.HTMLEscapeString(token))
		} else {
			sb.WriteString(template.HTMLEscapeString(token))
		}
		last = end
	}
	sb.WriteString(template.HTMLEscapeString(line[last:]))
	return sb.String()
}

// RenderBody renders a raw .chord file body into HTML for use inside a
// white-space: pre container: chord lines get <span class="chord"> tokens
// and <span class="timesig"> time signatures,
// everything else (lyrics, section labels, blank lines) passes through
// escaped as-is.
func RenderBody(body string) template.HTML {
	lines := strings.Split(body, "\n")
	out := make([]string, len(lines))
	for i, line := range lines {
		if isChordLine(line) {
			out[i] = renderChordLine(line)
		} else {
			out[i] = template.HTMLEscapeString(line)
		}
	}
	return template.HTML(strings.Join(out, "\n"))
}
