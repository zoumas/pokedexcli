package main

import (
	"fmt"
	"io"
	"os"
)

type config struct {
	w    io.Writer
	args []string
}

func newConfig(w io.Writer) *config {
	return &config{
		w: w,
	}
}

type command struct {
	callback    func(cfg *config) error
	name        string
	description string
}

func commands() map[string]command {
	return map[string]command{
		"exit": {
			callback:    commandExit,
			name:        "exit",
			description: "Exit the Pokedex",
		},
		"help": {
			callback:    commandHelp,
			name:        "help",
			description: "Displays a help message",
		},
	}
}

func commandExit(cfg *config) error {
	fmt.Fprintln(cfg.w, "Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(cfg *config) error {
	cmds := commands()

	if len(cfg.args) != 0 {
		cmd, ok := cmds[cfg.args[0]]
		if !ok {
			return fmt.Errorf("Unknown command: %q", cfg.args[0])
		}
		fmt.Fprintf(cfg.w, "%s: %s\n", cmd.name, cmd.description)
		return nil
	}

	fmt.Fprintln(cfg.w, "Welcome to the Pokedex!")
	fmt.Fprintln(cfg.w, "Usage:")
	fmt.Fprintln(cfg.w)

	for _, cmd := range cmds {
		fmt.Fprintf(cfg.w, "%s: %s\n", cmd.name, cmd.description)
	}
	return nil
}
