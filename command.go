package main

import (
	"fmt"
	"io"
	"os"
	"slices"
)

type command struct {
	name        string
	description string
	callback    func(io.Writer) error
}

func getCommands(w io.Writer) map[string]command {
	commands := map[string]command{}

	commands["exit"] = command{
		name:        "exit",
		description: "Exit the Pokedex",
		callback: func(io.Writer) error {
			return commandExit(w, os.Exit)
		},
	}

	commands["help"] = command{
		name:        "help",
		description: "Displays a help message",
		callback: func(io.Writer) error {
			return commandHelp(w, commands)
		},
	}

	return commands
}

func commandExit(w io.Writer, exit func(code int)) error {
	_, _ = fmt.Fprintln(w, "Closing the Pokedex... Goodbye!")
	exit(0)
	return nil
}

func commandHelp(w io.Writer, commands map[string]command) error {
	_, _ = fmt.Fprintln(w, "Welcome to the Pokedex!")
	_, _ = fmt.Fprintln(w, "Usage:")
	_, _ = fmt.Fprintln(w)

	names := make([]string, 0, len(commands))
	for name := range commands {
		names = append(names, name)
	}
	slices.Sort(names)

	for _, name := range names {
		cmd := commands[name]
		_, _ = fmt.Fprintf(w, "%s: %s\n", cmd.name, cmd.description)
	}

	return nil
}
