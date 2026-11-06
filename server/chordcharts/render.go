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
	// Modulation marker in a section label, e.g. "B (-> IV):". The section's
	// numerals count from that degree of the chart's key instead.
	modulationRe       = regexp.MustCompile(`\(\s*->\s*([^)]*?)\s*\)`)
	modulationTargetRe = regexp.MustCompile(`^([b#]?)(VII|VI|V|IV|III|II|I)$`)

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

// isSectionLabel reports whether the line is a section label, e.g. "A part:".
func isSectionLabel(line string) bool {
	return strings.HasSuffix(strings.TrimSpace(line), ":")
}

// parseModulation returns the degree a section label modulates to, e.g. "4"
// for "B (-> IV):", and the [start, end) of the target numeral within the
// label. ok is false when the label has no "->" marker. err is set when it
// has one that doesn't name an uppercase numeral, since the new key is
// always a major key.
func parseModulation(label string) (degree string, loc []int, ok bool, err error) {
	m := modulationRe.FindStringSubmatchIndex(label)
	if m == nil {
		if strings.Contains(label, "->") {
			return "", nil, true, fmt.Errorf("key change must be written like \"(-> IV)\"")
		}
		return "", nil, false, nil
	}
	target := label[m[2]:m[3]]
	t := modulationTargetRe.FindStringSubmatch(target)
	if t == nil {
		return "", nil, true, fmt.Errorf("key change target %q must be an uppercase numeral with an optional b or #, such as IV or bVII", target)
	}
	return fmt.Sprintf("%s%d", t[1], romanDegrees[t[2]]), []int{m[2], m[3]}, true, nil
}

// renderSectionLabel escapes a section label, wrapping a key change target
// in <span class="modulation"> so the key picker can spell it as a letter.
func renderSectionLabel(line string) string {
	degree, loc, ok, err := parseModulation(line)
	if !ok || err != nil {
		return template.HTMLEscapeString(line)
	}
	return template.HTMLEscapeString(line[:loc[0]]) +
		fmt.Sprintf(`<span class="modulation" data-numeral="%s" data-degree="%s">%s</span>`,
			template.HTMLEscapeString(line[loc[0]:loc[1]]),
			template.HTMLEscapeString(degree),
			template.HTMLEscapeString(line[loc[0]:loc[1]])) +
		template.HTMLEscapeString(line[loc[1]:])
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
// tonic is the degree of the chart's key that the numerals count from, e.g.
// "4" in a section that modulates to IV, or "" for the chart's own key.
//
// Each chord span also swallows the spaces after it and records the total
// column width in data-width, so the key picker can pad a converted label
// (e.g. "IV" -> "C") back to that width and keep bars lined up.
func renderChordLine(line, tonic string) string {
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
			fmt.Fprintf(&sb, `<span class="chord" data-numeral="%s" data-degree="%s" data-quality="%s"%s data-width="%d">%s</span>`,
				template.HTMLEscapeString(token),
				template.HTMLEscapeString(c.Degree),
				template.HTMLEscapeString(c.Quality),
				tonicAttr(tonic),
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

func tonicAttr(tonic string) string {
	if tonic == "" {
		return ""
	}
	return fmt.Sprintf(` data-tonic="%s"`, template.HTMLEscapeString(tonic))
}

// RenderBody renders a raw .chord file body into HTML for use inside a
// white-space: pre container: chord lines get <span class="chord"> tokens
// and <span class="timesig"> time signatures,
// everything else (lyrics, section labels, blank lines) passes through
// escaped as-is. A section label with a key change, e.g. "B (-> IV):", makes
// the chords up to the next section label count from that degree.
func RenderBody(body string) template.HTML {
	lines := strings.Split(body, "\n")
	out := make([]string, len(lines))
	tonic := ""
	for i, line := range lines {
		if isChordLine(line) {
			out[i] = renderChordLine(line, tonic)
		} else if isSectionLabel(line) {
			tonic = ""
			if degree, _, ok, err := parseModulation(line); ok && err == nil {
				tonic = degree
			}
			out[i] = renderSectionLabel(line)
		} else {
			out[i] = template.HTMLEscapeString(line)
		}
	}
	return template.HTML(strings.Join(out, "\n"))
}
