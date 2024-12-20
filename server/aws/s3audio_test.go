package aws

import (
	"fmt"
	"testing"
)

func TestGetS3Songs(t *testing.T) {
	songs, err := getS3Songs()
	if err != nil {
		t.Errorf("listSongs() failed: %v", err)
	}
	fmt.Printf("%+v\n", songs)
}

func TestFormatAudioTitle(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"audio/this_is_a_test.mp3", "This Is A Test"},
	}
	fmt.Printf("testing %v\n", formatAudioTitle("audio/this_is_a_test.mp3"))
	for _, test := range tests {
		if got := formatAudioTitle(test.input); got != test.expected {
			t.Errorf("formatAudioTitle(%q) = %q, want %q", test.input, got, test.expected)
		}
	}
}
