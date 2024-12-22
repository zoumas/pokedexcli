package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/zoumas/pokedexcli/internal/cache"
	"github.com/zoumas/pokedexcli/internal/pokeapi"
)

func main() {
	const prompt = "Pokedex > "
	repl(prompt, os.Stdin, os.Stdout)
}

// repl starts a read-eval-print loop.
// It takes an io.Reader and io.Writer for dependency injection.
// This way the repl can be tested. However it is not practical to do so.
func repl(prompt string, r io.Reader, w io.Writer) {
	scanner := bufio.NewScanner(r)
	cfg := &config{
		w:       w,
		next:    "https://pokeapi.co/api/v2/location-area/",
		cache:   cache.NewCache(30 * time.Second),
		pokemon: make(map[string]*pokeapi.Pokemon),
	}
	cmds := commands()

	for {
		fmt.Fprint(w, prompt)

		if !scanner.Scan() {
			break
		}

		text, err := scanner.Text(), scanner.Err()
		if err != nil {
			fmt.Fprintf(w, "reading input: %v", err)
			return
		}

		input := clean(text)
		if len(input) == 0 {
			continue
		}

		cmd, ok := cmds[input[0]]
		if !ok {
			fmt.Fprintf(w, "Unknown command: %q\n", input[0])
			continue
		}

		cfg.args = nil
		if len(input) > 1 {
			cfg.args = input[1:]
		}

		err = cmd.callback(cfg)
		if err != nil {
			fmt.Fprintln(w, err)
			continue
		}
	}
}

// clean splits the user's input into "words" based on whitespace.
// It also lowercases the input and trims any leading and trailing whitespace.
// clean returns nil when the length of the input is zero.
func clean(input string) []string {
	if len(input) == 0 {
		return nil
	}
	return strings.Fields(strings.ToLower(input))
}
