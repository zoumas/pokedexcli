package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/iotest"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/zoumas/pokedexcli/internal/pokeapi"
	"github.com/zoumas/pokedexcli/internal/pokecache"
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

// newTestConfig returns a config writing to w, with commands from registry and
// a logger that sends records to the test log.
func newTestConfig(t *testing.T, w io.Writer, registry map[string]cliCommand) *config {
	t.Helper()
	return &config{
		w:        w,
		logger:   slog.New(slog.NewTextHandler(t.Output(), &slog.HandlerOptions{Level: slog.LevelDebug})),
		registry: registry,
	}
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

			exitCode := startREPL(strings.NewReader(c.input), newTestConfig(t, &w, newCommandRegistry()))

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

	exitCode := startREPL(iotest.ErrReader(readErr), newTestConfig(t, &w, newCommandRegistry()))

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
			callback: func(*config, []string) error {
				return errors.New("command failed")
			},
		},
	}
	var w bytes.Buffer

	exitCode := startREPL(strings.NewReader(input), newTestConfig(t, &w, registry))

	// The failure is reported to the user and the loop keeps going.
	const failure = "boom: command failed\n"
	want := prompt + failure + prompt + failure + prompt

	if exitCode != 0 {
		t.Errorf("startREPL(%q) exit code = %d, want 0", input, exitCode)
	}
	if diff := cmp.Diff(want, w.String()); diff != "" {
		t.Errorf("startREPL(%q) output diff (-want +got):\n%s", input, diff)
	}
}

func TestCommandHelpListsEveryRegisteredCommand(t *testing.T) {
	var w bytes.Buffer
	cfg := newTestConfig(t, &w, newCommandRegistry())

	if err := commandHelp(cfg, []string{}); err != nil {
		t.Fatalf("commandHelp() error = %v, want nil", err)
	}

	if diff := cmp.Diff(helpOutput(cfg.registry), w.String()); diff != "" {
		t.Errorf("commandHelp() output diff (-want +got):\n%s", diff)
	}
}

func TestCommandExitSignalsExit(t *testing.T) {
	var w bytes.Buffer

	err := commandExit(newTestConfig(t, &w, nil), []string{})

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
			cfg := newTestConfig(t, errWriter{writeErr}, newCommandRegistry())

			err := c.cmd(cfg, []string{})

			if !errors.Is(err, writeErr) {
				t.Errorf("%s(failing writer) error = %v, want %v", c.name, err, writeErr)
			}
		})
	}
}

// errWriter fails every write with err.
type errWriter struct{ err error }

func (e errWriter) Write([]byte) (int, error) { return 0, e.err }

func TestCommandExploreRequiresAName(t *testing.T) {
	var w bytes.Buffer

	err := commandExplore(newTestConfig(t, &w, nil), nil)

	if err == nil {
		t.Fatalf("commandExplore(no args) error = nil, want a usage error")
	}
	if got, want := w.String(), ""; got != want {
		t.Errorf("commandExplore(no args) output = %q, want %q", got, want)
	}
}

func TestCommandExplore(t *testing.T) {
	const body = `{
	  "name": "pastoria-city-area",
	  "pokemon_encounters": [
	    {"pokemon": {"name": "tentacool", "url": "https://example.test/pokemon/72/"}},
	    {"pokemon": {"name": "magikarp", "url": "https://example.test/pokemon/129/"}}
	  ]
	}`

	cases := []struct {
		name string
		area string
		body string
		want string
	}{
		{
			name: "lists every encountered pokemon",
			area: "pastoria-city-area",
			body: body,
			want: "Exploring pastoria-city-area...\nFound Pokemon:\n - tentacool\n - magikarp\n",
		},
		{
			name: "reports an area with no encounters",
			area: "empty-area",
			body: `{"name": "empty-area", "pokemon_encounters": []}`,
			want: "Exploring empty-area...\nNo Pokemon found.\n",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte(c.body))
			}))
			defer server.Close()

			var w bytes.Buffer
			cfg := newTestConfig(t, &w, nil)
			cfg.client = pokeapi.New(server.Client(), pokecache.New(t.Context(), cfg.logger, time.Minute), server.URL)

			if err := commandExplore(cfg, []string{c.area}); err != nil {
				t.Fatalf("commandExplore(%q) error = %v, want nil", c.area, err)
			}
			if diff := cmp.Diff(c.want, w.String()); diff != "" {
				t.Errorf("commandExplore(%q) output diff (-want +got):\n%s", c.area, diff)
			}
		})
	}
}
