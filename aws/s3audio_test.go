package aws

import (
	"fmt"
	"os"
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

func TestUploadAudioToS3(t *testing.T) {
	userHome := os.Getenv("HOME")
	file := fmt.Sprintf("%s/recordings/wasted_words_kick.wav", userHome)
	UploadAudioToS3(file)
}
