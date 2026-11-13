package chordcharts

import (
	"regexp"
	"strings"
)

// maxRootWidth is the widest letter root the key picker spells for a chord,
// e.g. "Bb" or "F#". Cells reserve room for it so the bars still line up
// after numerals are converted to letters ("I" -> "Bb").
const maxRootWidth = 2

// barRunRe matches a run of bars anywhere in a line, including one written
// up against a chord ("|i" or "IV|").
var barRunRe = regexp.MustCompile(`\|+`)

// FormatFile formats the body of a raw .chord file and leaves the header
// untouched. It returns raw unchanged if the file has no "---" separator.
func FormatFile(raw string) string {
	lines := strings.Split(raw, "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) == "---" {
			body := FormatBody(strings.Join(lines[i+1:], "\n"))
			return strings.Join(append(lines[:i+1], body), "\n")
		}
	}
	return raw
}

// FormatBody lines up the bars of every chord line in a chart body into
// shared columns, with one space between each bar and the chord after it.
// Bars written up against a chord ("|i") are split off it first. Section
// labels, blank lines, and any line that still isn't a valid chord line are
// left as they are.
//
// Each column is as wide as its widest cell, where a chord counts as the
// wider of its numeral and its longest letter spelling (a two-character
// root plus the quality, e.g. "Bbm7"). That keeps the bars aligned in every
// key, not just in Numerals.
func FormatBody(body string) string {
	lines := strings.Split(body, "\n")
	rows := make(map[int]chordRow)
	var cellWidths, barWidths []int
	for i, line := range lines {
		line = barRunRe.ReplaceAllString(line, " $0 ")
		if !isChordLine(line) {
			continue
		}
		r := splitChordRow(line)
		rows[i] = r
		for j, cell := range r.cells {
			cellWidths = growTo(cellWidths, j+1)
			cellWidths[j] = max(cellWidths[j], cellWidth(cell))
		}
		for j, bar := range r.bars {
			barWidths = growTo(barWidths, j+1)
			barWidths[j] = max(barWidths[j], len(bar))
		}
	}

	for i, r := range rows {
		var sb strings.Builder
		if cellWidths[0] > 0 {
			sb.WriteString(padRight(strings.Join(r.cells[0], " "), cellWidths[0]))
			sb.WriteString(" ")
		}
		for j, bar := range r.bars {
			sb.WriteString(padRight(bar, barWidths[j]))
			sb.WriteString(" ")
			sb.WriteString(padRight(strings.Join(r.cells[j+1], " "), cellWidths[j+1]))
			sb.WriteString(" ")
		}
		lines[i] = strings.TrimRight(sb.String(), " ")
	}
	return strings.Join(lines, "\n")
}

// chordRow is a chord line split at its bars: cells[0] holds the tokens
// before the first bar (often the first measure), and cells[j+1] the tokens
// after bars[j].
type chordRow struct {
	cells [][]string
	bars  []string
}

func splitChordRow(line string) chordRow {
	r := chordRow{cells: [][]string{nil}}
	for _, f := range strings.Fields(line) {
		if barTokenRe.MatchString(f) {
			r.bars = append(r.bars, f)
			r.cells = append(r.cells, nil)
			continue
		}
		last := len(r.cells) - 1
		r.cells[last] = append(r.cells[last], f)
	}
	return r
}

// cellWidth is the room a measure's tokens need, written with one space
// between them.
func cellWidth(tokens []string) int {
	width := 0
	for i, t := range tokens {
		if i > 0 {
			width++
		}
		w := len(t)
		if c, ok := parseChord(t); ok {
			w = max(w, maxRootWidth+len(c.Quality))
		}
		width += w
	}
	return width
}

func growTo(s []int, n int) []int {
	for len(s) < n {
		s = append(s, 0)
	}
	return s
}

func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}
