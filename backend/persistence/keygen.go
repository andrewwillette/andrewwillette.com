package persistence

import (
	"crypto/sha1"
	"fmt"
	"math/rand"
)

// NewSHA1Hash generates a new SHA1 hash based on
// a random number of characters.
func newSHA1Hash(n ...int) string {
	randomCharactersNumber := 32
	if len(n) > 0 {
		randomCharactersNumber = n[0]
	}
	randString := randomString(randomCharactersNumber)
	hash := sha1.New()
	hash.Write([]byte(randString))
	bs := hash.Sum(nil)

	return fmt.Sprintf("%x", bs)
}

var characterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

// randomString generates a random string of n length
func randomString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = characterRunes[rand.Intn(len(characterRunes))]
	}
	return string(b)
}
