package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
)

// startREPL reads commands from r, one per line, and writes the prompt and the
// output of each command to w. Commands are looked up in registry; unknown
// commands are reported and the loop continues, as does a command that fails
// for any reason other than errExit. It stops when r is exhausted or when a
// command returns errExit. It returns 0 on a clean end of input or a requested
// exit, and 1 if reading r failed.
func startREPL(r io.Reader, w io.Writer, registry map[string]cliCommand) (exitCode int) {
	const prompt = "Pokedex > "
	scanner := bufio.NewScanner(r)
	cfg := commandConfig{
		w:        w,
		registry: registry,
	}

	for {
		_, _ = fmt.Fprint(w, prompt)
		if !scanner.Scan() {
			break
		}

		text := scanner.Text()
		input := cleanInput(text)
		if len(input) == 0 {
			continue
		}

		command := input[0]
		c, ok := registry[command]
		if !ok {
			_, _ = fmt.Fprintln(w, "Unknown command")
			continue
		}

		if err := c.callback(cfg); err != nil {
			if errors.Is(err, errExit) {
				return 0
			}

			log.Printf("command error: %v", err)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("scan error: %v", err)
		return 1
	}
	return 0
}

// cleanInput splits the user's input into "words" based on whitespace.
// It also lowercases the input and trims any leading or trailing whitespace.
func cleanInput(text string) []string {
	return strings.Fields(strings.ToLower(text))
}
