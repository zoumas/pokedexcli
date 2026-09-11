package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/google/go-cmp/cmp"
)

func TestCleanInput(t *testing.T) {
	cases := []struct {
		name string
		text string
		want []string
	}{
		{
			name: "surrounding and interior whitespace",
			text: "	hello	world	",
			want: []string{"hello", "world"},
		},
		{
			name: "pokemon names in mixed case return all lower",
			text: "Charmander Bulbasaur PIKACHU",
			want: []string{"charmander", "bulbasaur", "pikachu"},
		},
		{
			name: "empty input returns empty slice",
			text: "",
			want: []string{},
		},
		{
			name: "whitespace only",
			text: "  	  	 ",
			want: []string{},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := cleanInput(c.text)
			if diff := cmp.Diff(c.want, got); diff != "" {
				t.Errorf("cleanInput(%q): diff (-want +got):\n%s", c.text, diff)
			}
		})
	}
}

func TestStartREPL(t *testing.T) {
	const prompt = "Pokedex > "

	cases := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "single command",
			input: "Pikachu\n",
			want:  prompt + "Your command was: pikachu\n" + prompt,
		},
		{
			name:  "only the first word is echoed",
			input: "Charmander Bulbasaur PIKACHU\n",
			want:  prompt + "Your command was: charmander\n" + prompt,
		},
		{
			name:  "successive commands",
			input: "first\nSECOND\n",
			want: prompt + "Your command was: first\n" +
				prompt + "Your command was: second\n" +
				prompt,
		},
		{
			name:  "blank lines reprompt without output",
			input: "\n   \n",
			want:  prompt + prompt + prompt,
		},
		{
			name:  "no input",
			input: "",
			want:  prompt,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var w bytes.Buffer

			exitCode := startREPL(strings.NewReader(c.input), &w)

			if exitCode != 0 {
				t.Errorf("startREPL(%q) exit code = %d, want 0", c.input, exitCode)
			}
			if diff := cmp.Diff(c.want, w.String()); diff != "" {
				t.Errorf("startREPL(%q) output diff (-want +got):\n%s", c.input, diff)
			}
		})
	}
}

func TestStartREPLReadError(t *testing.T) {
	wantErr := errors.New("boom")
	var w bytes.Buffer

	exitCode := startREPL(iotest.ErrReader(wantErr), &w)

	if exitCode != 1 {
		t.Errorf("startREPL(failing reader) exit code = %d, want 1", exitCode)
	}
	if got, want := w.String(), "Pokedex > "; got != want {
		t.Errorf("startREPL(failing reader) output = %q, want %q", got, want)
	}
}
