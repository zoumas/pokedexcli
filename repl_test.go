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
	const (
		prompt   = "Pokedex > "
		helpText = "Welcome to the Pokedex!\nUsage:\n\n" +
			"exit: Exit the Pokedex\n" +
			"help: Displays a help menu\n"
		goodbye = "Closing the Pokedex... Goodbye!\n"
	)

	cases := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "help lists every registered command",
			input: "help\n",
			want:  prompt + helpText + prompt,
		},
		{
			name:  "command lookup is case insensitive",
			input: "HELP\n",
			want:  prompt + helpText + prompt,
		},
		{
			name:  "only the first word is treated as the command",
			input: "help me please\n",
			want:  prompt + helpText + prompt,
		},
		{
			name:  "unknown command is reported and the loop continues",
			input: "bogus\nhelp\n",
			want:  prompt + "Unknown command\n" + prompt + helpText + prompt,
		},
		{
			name:  "exit stops the loop and ignores later input",
			input: "exit\nhelp\n",
			want:  prompt + goodbye,
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

func TestCommandHelpListsEveryRegisteredCommand(t *testing.T) {
	var w bytes.Buffer

	cfg := commandConfig{w: &w, registry: getCommandRegistry()}

	if err := commandHelp(cfg); err != nil {
		t.Fatalf("commandHelp() error = %v, want nil", err)
	}

	got := w.String()
	for name, c := range cfg.registry {
		if !strings.Contains(got, name+": "+c.description+"\n") {
			t.Errorf("commandHelp() output is missing command %q, got:\n%s", name, got)
		}
	}
}

func TestCommandExitSignalsExit(t *testing.T) {
	var w bytes.Buffer

	err := commandExit(commandConfig{w: &w})

	if !errors.Is(err, errExit) {
		t.Errorf("commandExit() error = %v, want errExit", err)
	}
	if got, want := w.String(), "Closing the Pokedex... Goodbye!\n"; got != want {
		t.Errorf("commandExit() output = %q, want %q", got, want)
	}
}
