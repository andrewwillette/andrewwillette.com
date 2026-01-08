package images

import "testing"

func TestGenerateSongCoverArt(t *testing.T) {
	testKeys := []string{
		"audio/billy_in_the_lowground.wav",
		"audio/grey_eagle.wav",
	}
	GenerateSongCoverArt(testKeys)
}
