package main

import (
	"cmp"
	"fmt"
	"io"
	"maps"
	"slices"

	"github.com/zoumas/pokedexcli/internal/pokeapi"
)

type config struct {
	w                   io.Writer
	registry            map[string]cliCommand
	nextLocationURL     *string
	previousLocationURL *string
	client              *pokeapi.Client
}

func newConfig(w io.Writer, client *pokeapi.Client) *config {
	startingURL := pokeapi.StartingLocationAreasURL

	return &config{
		w:                   w,
		registry:            newCommandRegistry(),
		client:              client,
		nextLocationURL:     &startingURL,
		previousLocationURL: nil,
	}
}

type commandFunc func(cfg *config) error

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
		"map": {
			name:        "map",
			description: "Displays the names of the next 20 location areas of the world",
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "Displays the names of the previous 20 location areas of the world",
			callback:    commandMapb,
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

func commandExit(cfg *config) error {
	if _, err := fmt.Fprintln(cfg.w, "Closing the Pokedex... Goodbye!"); err != nil {
		return err
	}
	return errExit
}

func commandHelp(cfg *config) error {
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

func commandMap(cfg *config) error {
	if cfg.nextLocationURL == nil {
		_, err := fmt.Fprintln(cfg.w, "you're on the last page")
		return err
	}
	return handleMap(cfg, *cfg.nextLocationURL)
}

func commandMapb(cfg *config) error {
	if cfg.previousLocationURL == nil {
		_, err := fmt.Fprintln(cfg.w, "you're on the first page")
		return err
	}
	return handleMap(cfg, *cfg.previousLocationURL)
}

func handleMap(cfg *config, url string) error {
	locationAreas, err := cfg.client.GetLocationAreas(url)
	if err != nil {
		return err
	}

	cfg.nextLocationURL = locationAreas.Next
	cfg.previousLocationURL = locationAreas.Previous

	for _, a := range locationAreas.Results {
		if _, err := fmt.Fprintln(cfg.w, a.Name); err != nil {
			return err
		}
	}
	return nil
}
