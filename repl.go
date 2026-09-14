package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
)

// startREPL reads commands from r, one per line, and writes the prompt and the
// output of each command to cfg.w. Commands are looked up in cfg.registry;
// unknown commands are reported and the loop continues, as does a command that
// fails for any reason other than errExit. A failure is both logged and
// reported to cfg.w, so the user always sees it. It stops when r is exhausted or when a
// command returns errExit. It returns 0 on a clean end of input or a requested
// exit, and 1 if reading r failed.
func startREPL(r io.Reader, cfg *config) (exitCode int) {
	const prompt = "Pokedex > "
	scanner := bufio.NewScanner(r)

	for {
		_, _ = fmt.Fprint(cfg.w, prompt)
		if !scanner.Scan() {
			break
		}

		text := scanner.Text()
		input := cleanInput(text)
		if len(input) == 0 {
			continue
		}

		command := input[0]
		c, ok := cfg.registry[command]
		if !ok {
			_, _ = fmt.Fprintln(cfg.w, "Unknown command")
			continue
		}

		args := input[1:]

		if err := c.callback(cfg, args); err != nil {
			if errors.Is(err, errExit) {
				return 0
			}

			cfg.logger.Error("command failed", slog.String("command", command), slog.Any("error", err))
			if _, err := fmt.Fprintf(cfg.w, "%s: %v\n", command, err); err != nil {
				cfg.logger.Error("writing command error failed", slog.Any("error", err))
			}
		}
	}

	if err := scanner.Err(); err != nil {
		cfg.logger.Error("reading input failed", slog.Any("error", err))
		return 1
	}
	return 0
}

// cleanInput splits the user's input into "words" based on whitespace.
// It also lowercases the input and trims any leading or trailing whitespace.
func cleanInput(text string) []string {
	return strings.Fields(strings.ToLower(text))
}
