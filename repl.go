package main

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

func repl(r io.Reader, w io.Writer) error {
	const PROMPT = "Pokedex > "

	commands := getCommands(w)

	scanner := bufio.NewScanner(r)
	for {
		_, _ = fmt.Fprint(w, PROMPT)
		if !scanner.Scan() {
			break
		}

		text := scanner.Text()

		cleaned := cleanInput(text)
		if len(cleaned) == 0 {
			continue
		}

		firstWord := cleaned[0]

		cmd, ok := commands[firstWord]
		if !ok {
			_, _ = fmt.Fprintln(w, "Unknown command")
			continue
		}

		if err := cmd.callback(w, cleaned[1:]...); err != nil {
			_, _ = fmt.Fprintf(w, "Error executing command: %v\n", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading input: %w", err)
	}
	return nil
}

func cleanInput(s string) []string {
	return strings.Fields(strings.ToLower(s))
}
