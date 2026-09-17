package main

import (
	"cmp"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"math/rand/v2"
	"slices"

	"github.com/zoumas/pokedexcli/internal/pokeapi"
)

type config struct {
	w                   io.Writer
	logger              *slog.Logger
	registry            map[string]cliCommand
	nextLocationURL     *string
	previousLocationURL *string
	client              *pokeapi.Client
	caughtPokemon       map[string]*pokeapi.Pokemon
	rng                 *rand.Rand
}

// catchThreshold tunes how catchable Pokemon are overall. A Pokemon with
// base_experience equal to it is caught half the time.
const catchThreshold = 50

// caught reports whether a Pokemon with the given base experience is caught.
// The chance is catchThreshold/(baseExperience+catchThreshold), so it falls as
// base experience rises and never reaches 0 or 1.
func caught(rng *rand.Rand, baseExperience int) bool {
	return rng.IntN(baseExperience+catchThreshold) < catchThreshold
}

func newConfig(w io.Writer, client *pokeapi.Client, logger *slog.Logger, rng *rand.Rand) *config {
	startingURL := client.LocationAreasURL()

	return &config{
		w:                   w,
		logger:              logger,
		registry:            newCommandRegistry(),
		client:              client,
		nextLocationURL:     &startingURL,
		previousLocationURL: nil,
		caughtPokemon:       make(map[string]*pokeapi.Pokemon),
		rng:                 rng,
	}
}

type commandFunc func(cfg *config, args []string) error

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
		"explore": {
			name:        "explore",
			description: "Get information about Pokemon encounters of a location area",
			callback:    commandExplore,
		},
		"catch": {
			name:        "catch",
			description: "Attempt to catch a Pokemon",
			callback:    commandCatch,
		},
		"inspect": {
			name:        "inspect",
			description: "Get information about a Pokemon",
			callback:    commandInspect,
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

func commandExit(cfg *config, args []string) error {
	if _, err := fmt.Fprintln(cfg.w, "Closing the Pokedex... Goodbye!"); err != nil {
		return err
	}
	return errExit
}

func commandHelp(cfg *config, args []string) error {
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

func commandMap(cfg *config, args []string) error {
	if cfg.nextLocationURL == nil {
		_, err := fmt.Fprintln(cfg.w, "you're on the last page")
		return err
	}
	return showLocationAreas(cfg, *cfg.nextLocationURL)
}

func commandMapb(cfg *config, args []string) error {
	if cfg.previousLocationURL == nil {
		_, err := fmt.Fprintln(cfg.w, "you're on the first page")
		return err
	}
	return showLocationAreas(cfg, *cfg.previousLocationURL)
}

// showLocationAreas fetches the page of location areas at url, records its
// pagination URLs in cfg, and writes each area name to cfg.w.
func showLocationAreas(cfg *config, url string) error {
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

// commandExplore writes the names of the Pokemon that can be encountered in the
// location area named by the first argument.
func commandExplore(cfg *config, args []string) error {
	if len(args) == 0 {
		return errors.New("usage: explore <location-area>")
	}
	name := args[0]

	if _, err := fmt.Fprintf(cfg.w, "Exploring %s...\n", name); err != nil {
		return err
	}

	area, err := cfg.client.GetLocationArea(name)
	if err != nil {
		return err
	}

	if len(area.PokemonEncounters) == 0 {
		_, err := fmt.Fprintln(cfg.w, "No Pokemon found.")
		return err
	}

	if _, err := fmt.Fprintln(cfg.w, "Found Pokemon:"); err != nil {
		return err
	}
	for _, e := range area.PokemonEncounters {
		if _, err := fmt.Fprintf(cfg.w, " - %s\n", e.Pokemon.Name); err != nil {
			return err
		}
	}
	return nil
}

// commandCatch attempts to catch the Pokemon named by the first argument,
// adding it to cfg.caughtPokemon on success.
func commandCatch(cfg *config, args []string) error {
	if len(args) == 0 {
		return errors.New("usage: catch <pokemon>")
	}
	name := args[0]

	if _, err := fmt.Fprintf(cfg.w, "Throwing a Pokeball at %s...\n", name); err != nil {
		return err
	}

	pokemon, err := cfg.client.GetPokemon(name)
	if err != nil {
		return err
	}

	if !caught(cfg.rng, pokemon.BaseExperience) {
		if _, err := fmt.Fprintf(cfg.w, "%s escaped!\n", name); err != nil {
			return err
		}
		return nil
	}

	if _, err := fmt.Fprintf(cfg.w, "%s was caught!\n", name); err != nil {
		return err
	}

	cfg.caughtPokemon[name] = pokemon

	return nil
}

// commandInspect writes the details of a previously caught Pokemon named by the
// first argument. It reads only cfg.caughtPokemon and makes no API call.
func commandInspect(cfg *config, args []string) error {
	if len(args) == 0 {
		return errors.New("usage: inspect <pokemon>")
	}
	name := args[0]

	p, ok := cfg.caughtPokemon[name]
	if !ok {
		_, err := fmt.Fprintln(cfg.w, "you have not caught that pokemon")
		return err
	}

	if _, err := fmt.Fprintf(
		cfg.w,
		"Name: %s\nHeight: %d\nWeight: %d\nStats:\n",
		p.Name,
		p.Height,
		p.Weight,
	); err != nil {
		return err
	}

	for _, s := range p.Stats {
		if _, err := fmt.Fprintf(cfg.w, "  -%s: %d\n", s.Stat.Name, s.BaseStat); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintln(cfg.w, "Types:"); err != nil {
		return err
	}
	for _, t := range p.Types {
		if _, err := fmt.Fprintf(cfg.w, "  - %s\n", t.Type.Name); err != nil {
			return err
		}
	}

	return nil
}
