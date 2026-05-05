package main

import (
	"bytes"
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

func Test_repl(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "help command prints stable usage output",
			input: "help\n",
			want: "Pokedex > Welcome to the Pokedex!\n" +
				"Usage:\n\n" +
				"catch: Attempts to catch a Pokemon by name\n" +
				"exit: Exit the Pokedex\n" +
				"explore: Explores a location area and lists the Pokemon that can be found there\n" +
				"help: Displays a help message\n" +
				"inspect: Desplays detailed information about a Pokemon in your Pokedex by name\n" +
				"map: Displays a list of location areas\n" +
				"mapb: Displays a list of location areas in the previous page\n" +
				"pokedex: Lists all the Pokemon you have caught in your Pokedex\n" +
				"Pokedex > ",
		},
		{
			name:  "unknown commands are rejected",
			input: "CHARMANDER is better than bulbasaur\n\nPikachu is kinda mean to ash\n",
			want:  "Pokedex > Unknown command\nPokedex > Pokedex > Unknown command\nPokedex > ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := bytes.NewBufferString(tt.input)
			out := bytes.NewBuffer(nil)

			if err := repl(in, out); err != nil {
				t.Fatalf("repl() returned error: %v", err)
			}

			if got := out.String(); got != tt.want {
				t.Fatalf("repl() output = %q, want %q", got, tt.want)
			}
		})
	}
}
