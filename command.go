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

func getCommandRegistry() map[string]cliCommand {
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

type exitError string

func (e exitError) Error() string {
	return string(e)
}

const errExit exitError = "exit"

func commandExit(cfg commandConfig) error {
	_, _ = fmt.Fprintln(cfg.w, "Closing the Pokedex... Goodbye!")
	return errExit
}

func commandHelp(cfg commandConfig) error {
	commands := slices.Collect(maps.Values(cfg.registry))
	slices.SortFunc(commands, func(a, b cliCommand) int {
		return cmp.Compare(a.name, b.name)
	})

	_, _ = fmt.Fprintf(cfg.w, "Welcome to the Pokedex!\nUsage:\n\n")

	for _, c := range commands {
		_, _ = fmt.Fprintf(cfg.w, "%s: %s\n", c.name, c.description)
	}

	return nil
}
