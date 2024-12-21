package main

import (
	"slices"
	"testing"
)

func TestClean(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "remove leading trailing and in-between whitespace",
			input:    "  hello  world  ",
			expected: []string{"hello", "world"},
		},
		{
			name:     "lowercase all words",
			input:    "  Charmander  Bulbasaur  PIKACHU  ",
			expected: []string{"charmander", "bulbasaur", "pikachu"},
		},
		{
			name:     "empty input returns nil",
			input:    "",
			expected: nil,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual := clean(c.input)

			if !slices.Equal(c.expected, actual) {
				t.Errorf("\ninput: %q\nexpected: %+v\nactual: %+v",
					c.input, c.expected, actual)
			}
		},
		)
	}
}

func BenchmarkClean(b *testing.B) {
	b.Run("empty input", func(b *testing.B) {
		for range b.N {
			clean("")
		}
	})

	b.Run("whitespace and lowercase", func(b *testing.B) {
		for range b.N {
			clean("  Charmander  Bulbasaur  PIKACHU  ")
		}
	})
}
