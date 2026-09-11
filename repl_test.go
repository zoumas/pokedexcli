package main

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/google/go-cmp/cmp"
)

const prompt = "Pokedex > "

// helpOutput builds the text commandHelp writes for registry, so that tests do
// not restate which commands exist.
func helpOutput(registry map[string]cliCommand) string {
	var b strings.Builder
	b.WriteString("Welcome to the Pokedex!\nUsage:\n\n")
	for _, c := range sortedCommands(registry) {
		fmt.Fprintf(&b, "%s: %s\n", c.name, c.description)
	}
	return b.String()
}

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
	const goodbye = "Closing the Pokedex... Goodbye!\n"
	helpText := helpOutput(newCommandRegistry())

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

			exitCode := startREPL(strings.NewReader(c.input), &w, newCommandRegistry())

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
	readErr := errors.New("boom")
	var w bytes.Buffer

	exitCode := startREPL(iotest.ErrReader(readErr), &w, newCommandRegistry())

	if exitCode != 1 {
		t.Errorf("startREPL(failing reader) exit code = %d, want 1", exitCode)
	}
	if got, want := w.String(), prompt; got != want {
		t.Errorf("startREPL(failing reader) output = %q, want %q", got, want)
	}
}

func TestStartREPLCommandError(t *testing.T) {
	const input = "boom\nboom\n"
	registry := map[string]cliCommand{
		"boom": {
			name:        "boom",
			description: "always fails",
			callback: func(commandConfig) error {
				return errors.New("command failed")
			},
		},
	}
	var w bytes.Buffer

	exitCode := startREPL(strings.NewReader(input), &w, registry)

	if exitCode != 0 {
		t.Errorf("startREPL(%q) exit code = %d, want 0", input, exitCode)
	}
	if diff := cmp.Diff(prompt+prompt+prompt, w.String()); diff != "" {
		t.Errorf("startREPL(%q) output diff (-want +got):\n%s", input, diff)
	}
}

func TestCommandHelpListsEveryRegisteredCommand(t *testing.T) {
	var w bytes.Buffer
	cfg := commandConfig{w: &w, registry: newCommandRegistry()}

	if err := commandHelp(cfg); err != nil {
		t.Fatalf("commandHelp() error = %v, want nil", err)
	}

	if diff := cmp.Diff(helpOutput(cfg.registry), w.String()); diff != "" {
		t.Errorf("commandHelp() output diff (-want +got):\n%s", diff)
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

func TestCommandWriteErrors(t *testing.T) {
	writeErr := errors.New("write failed")

	cases := []struct {
		name string
		cmd  commandFunc
	}{
		{name: "commandHelp", cmd: commandHelp},
		{name: "commandExit", cmd: commandExit},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cfg := commandConfig{w: errWriter{writeErr}, registry: newCommandRegistry()}

			err := c.cmd(cfg)

			if !errors.Is(err, writeErr) {
				t.Errorf("%s(failing writer) error = %v, want %v", c.name, err, writeErr)
			}
		})
	}
}

// errWriter fails every write with err.
type errWriter struct{ err error }

func (e errWriter) Write([]byte) (int, error) { return 0, e.err }
