package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"strings"
)

// startREPL reads commands from r, one per line, and writes the prompt and the
// result of each command to w. It runs until r is exhausted. It returns 0 on a
// clean end of input and 1 if reading r failed.
func startREPL(r io.Reader, w io.Writer) (exitCode int) {
	const prompt = "Pokedex > "
	scanner := bufio.NewScanner(r)

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
		_, _ = fmt.Fprintf(w, "Your command was: %s\n", command)
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
