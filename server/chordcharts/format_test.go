package chordcharts

import (
	"testing"
)

func TestFormatBodyAlignsBars(t *testing.T) {
	body := "A Part:\n" +
		"I     |  I IV  |  I     |  I      |\n" +
		"V |   V |  IV    |  2/4 IV | I     | I     |\n" +
		"\n" +
		"some lyric that isn't a chord line"
	want := "A Part:\n" +
		"I  | I IV  | I  | I      |\n" +
		"V  | V     | IV | 2/4 IV | I  | I  |\n" +
		"\n" +
		"some lyric that isn't a chord line"
	if got := FormatBody(body); got != want {
		t.Errorf("FormatBody:\n got %q\nwant %q", got, want)
	}
}

func TestFormatBodyReservesRoomForLetterChords(t *testing.T) {
	// "ii7" can spell as "Bbm7", so its column is four wide.
	got := FormatBody("ii7 | V |\nI | I |")
	want := "ii7  | V  |\nI    | I  |"
	if got != want {
		t.Errorf("FormatBody:\n got %q\nwant %q", got, want)
	}
}

func TestFormatBodyReservesRoomForSlashBass(t *testing.T) {
	// "ii7/V" can spell as "Bbm7/Ab", so its column is seven wide.
	got := FormatBody("ii7/V | V |\nI | I |")
	want := "ii7/V   | V  |\nI       | I  |"
	if got != want {
		t.Errorf("FormatBody:\n got %q\nwant %q", got, want)
	}
}

func TestFormatBodyLeadingBar(t *testing.T) {
	got := FormatBody("|  I  |  V  |\n||  IV |  I  ||")
	want := "|  I  | V  |\n|| IV | I  ||"
	if got != want {
		t.Errorf("FormatBody:\n got %q\nwant %q", got, want)
	}
}

func TestFormatBodySplitsBarsFromChords(t *testing.T) {
	got := FormatBody("i   | i   | bVII | bVII |\ni  | i  |i  bVII  | v i|")
	want := "i   | i   | bVII     | bVII    |\ni   | i   | i bVII   | v i     |"
	if got != want {
		t.Errorf("FormatBody:\n got %q\nwant %q", got, want)
	}
}

func TestFormatBodyIsIdempotent(t *testing.T) {
	once := FormatBody("I | I IV | 2/4 V |\nbVII7 | vi | I |\n\nB (-> IV):\nI | V7 | I |")
	if twice := FormatBody(once); twice != once {
		t.Errorf("formatting twice changed the body:\n once %q\ntwice %q", once, twice)
	}
}

func TestFormatFileKeepsHeader(t *testing.T) {
	raw := "Title:   Spaced Out\nKey: D\n---\nI | V |\n"
	want := "Title:   Spaced Out\nKey: D\n---\nI  | V  |\n"
	if got := FormatFile(raw); got != want {
		t.Errorf("FormatFile:\n got %q\nwant %q", got, want)
	}
}
