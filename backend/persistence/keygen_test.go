package persistence

import (
	"testing"
)

func TestShaGeneration(t *testing.T) {
	var shaMap = make(map[string]int)
	for i := 0; i <= 100; i++ {
		sha1 := newSHA1Hash(2)
		_, exists := shaMap[sha1]
		if exists {
			t.Log("sha already created")
			t.Fail()
		}
		shaMap[sha1] = 1
	}
}
