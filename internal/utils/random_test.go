package random

import (
	"strconv"
	"testing"
	"unicode"

	"github.com/stretchr/testify/assert"
)

func TestRandom(t *testing.T) {
	sizes := []int{1, 4, 10, 100}

	for _, size := range sizes {
		t.Run("Generate string of length "+strconv.Itoa(size), func(t *testing.T) {
			res := RandomString(size)

			assert.Len(t, res, size, "Generated string should have the expected length")

			for _, char := range res {
				assert.True(t, unicode.IsDigit(char) || unicode.IsLetter(char), "Invalid character '%c', found", char)
			}
		})
	}

	t.Run("It checks generated string are unique", func(t *testing.T) {
		c1 := RandomString(10)
		c2 := RandomString(10)

		assert.NotEqual(t, c1, c2, "Generated random strings are equal")
	})
}
