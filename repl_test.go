package main

import (
	"slices"
	"testing"
)

func Test_cleanInput(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		s    string
		want []string
	}{
		{
			name: "empty string returns empty slice",
			s:    "",
			want: []string{},
		},
		{
			name: "whitespace only returns empty slice",
			s:    " \t\n\r ",
			want: []string{},
		},
		{
			name: "trim leading trailing and in between whitespace",
			s:    " 	hello  world 	",
			want: []string{"hello", "world"},
		},
		{
			name: "single word is returned as one lowercased token",
			s:    "Pikachu",
			want: []string{"pikachu"},
		},
		{
			name: "lower case input",
			s:    "Charmander Bulbasaur PIKACHU",
			want: []string{"charmander", "bulbasaur", "pikachu"},
		},
		{
			name: "handles mixed whitespace separators",
			s:    "Squirtle\tWartortle\nBlastoise",
			want: []string{"squirtle", "wartortle", "blastoise"},
		},
		{
			name: "preserves punctuation within tokens",
			s:    "Mr. Mime Farfetch'd Porygon-Z",
			want: []string{"mr.", "mime", "farfetch'd", "porygon-z"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cleanInput(tt.s)
			if !slices.Equal(got, tt.want) {
				t.Errorf("cleanInput() = %v, want %v", got, tt.want)
			}
		})
	}
}
