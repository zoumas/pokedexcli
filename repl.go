package main

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

func repl(r io.Reader, w io.Writer) error {
	const PROMPT = "Pokedex > "

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
		_, _ = fmt.Fprintln(w, "Your command was:", firstWord)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading input: %w", err)
	}
	return nil
}

func cleanInput(s string) []string {
	return strings.Fields(strings.ToLower(s))
}
