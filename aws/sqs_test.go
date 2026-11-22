package aws

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func matchedNames(t *testing.T, body string, candidates []contentType) []string {
	t.Helper()
	matched, err := matchContentTypes(body, candidates)
	require.NoError(t, err)
	var names []string
	for _, ct := range matched {
		names = append(names, ct.name)
	}
	return names
}

func TestMatchContentTypes(t *testing.T) {
	candidates := []contentType{
		{name: "audio", prefix: "audio/"},
		{name: "shows", prefix: "shows/"},
		{name: "unset", prefix: ""},
	}

	t.Run("parses the real S3 event shape", func(t *testing.T) {
		body := `{"Records":[{"eventName":"ObjectCreated:Put","s3":{"object":{"key":"shows/a+b.json"}}}]}`
		require.Equal(t, []string{"shows"}, matchedNames(t, body, candidates))
	})

	t.Run("handles every record once", func(t *testing.T) {
		body := `{"Records":[
			{"s3":{"object":{"key":"audio/a.wav"}}},
			{"s3":{"object":{"key":"audio/b.wav"}}},
			{"s3":{"object":{"key":"shows/c.json"}}}]}`
		require.Equal(t, []string{"audio", "shows"}, matchedNames(t, body, candidates))
	})

	t.Run("unset prefix matches nothing", func(t *testing.T) {
		body := `{"Records":[{"s3":{"object":{"key":"other/x"}}}]}`
		require.Empty(t, matchedNames(t, body, candidates))
	})

	t.Run("malformed body is an error", func(t *testing.T) {
		_, err := matchContentTypes("not json", candidates)
		require.Error(t, err)
	})
}
