package main

import (
	"cmp"
	"fmt"
	"io"
	"maps"
	"slices"
)

type commandConfig struct {
	w        io.Writer
	registry map[string]cliCommand
}

type commandFunc func(cfg commandConfig) error

type cliCommand struct {
	name        string
	description string
	callback    commandFunc
}

// newCommandRegistry returns the commands the REPL accepts, keyed by name.
func newCommandRegistry() map[string]cliCommand {
	return map[string]cliCommand{
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
		"help": {
			name:        "help",
			description: "Displays a help menu",
			callback:    commandHelp,
		},
	}
}

// sortedCommands returns the commands in registry ordered by name, so that
// output built from a registry is stable across runs.
func sortedCommands(registry map[string]cliCommand) []cliCommand {
	commands := slices.Collect(maps.Values(registry))
	slices.SortFunc(commands, func(a, b cliCommand) int {
		return cmp.Compare(a.name, b.name)
	})
	return commands
}

// exitError is returned by a command that asks the REPL to stop.
type exitError string

func (e exitError) Error() string {
	return string(e)
}

// errExit tells the REPL to stop reading commands and return successfully.
const errExit exitError = "exit"

func commandExit(cfg commandConfig) error {
	if _, err := fmt.Fprintln(cfg.w, "Closing the Pokedex... Goodbye!"); err != nil {
		return err
	}
	return errExit
}

func commandHelp(cfg commandConfig) error {
	if _, err := fmt.Fprint(cfg.w, "Welcome to the Pokedex!\nUsage:\n\n"); err != nil {
		return err
	}

	for _, c := range sortedCommands(cfg.registry) {
		if _, err := fmt.Fprintf(cfg.w, "%s: %s\n", c.name, c.description); err != nil {
			return err
		}
	}

	return nil
}
