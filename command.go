package main

import (
	"fmt"
	"io"
	"os"

	"github.com/zoumas/pokedexcli/internal/cache"
	"github.com/zoumas/pokedexcli/internal/pokeapi"
)

type config struct {
	w        io.Writer
	cache    *cache.Cache
	next     string
	previous string
	args     []string
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
		"map": {
			callback:    commandMap,
			name:        "map",
			description: "Display the names of the next 20 location areas of the Pokemon world",
		},
		"mapb": {
			callback:    commandMapb,
			name:        "mapb",
			description: "The opposite of map. Displays the previous 20 location areas",
		},
		"explore": {
			callback:    commandExplore,
			name:        "explore",
			description: "See possible encounters for the given location area",
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

func commandMap(cfg *config) error {
	if cfg.next == "" {
		fmt.Fprintln(cfg.w,
			"An unseen force prevents you from going forward. It seems this is the end...")
		return nil
	}
	return listLocationsAreas(cfg, cfg.next)
}

func commandMapb(cfg *config) error {
	if cfg.previous == "" {
		fmt.Fprintln(cfg.w,
			"A gust of wind blows leaves around 🍃... There is nothing back there.")
		return nil
	}
	return listLocationsAreas(cfg, cfg.previous)
}

func listLocationsAreas(cfg *config, url string) error {
	l, err := pokeapi.GetLocationAreas(cfg.cache, url)
	if err != nil {
		return err
	}

	cfg.next = l.Next
	cfg.previous = l.Previous

	for _, r := range l.Results {
		fmt.Fprintln(cfg.w, r.Name)
	}

	return nil
}

func commandExplore(cfg *config) error {
	if len(cfg.args) == 0 {
		return fmt.Errorf("There is nothing to explore...")
	}
	name := cfg.args[0]
	fmt.Fprintf(cfg.w, "Exploring %s...\n", name)

	l, err := pokeapi.GetLocationArea(cfg.cache, name)
	if err != nil {
		return fmt.Errorf("Something went wrong while exploring: %v", err)
	}

	for _, e := range l.PokemonEncounters {
		fmt.Fprintln(cfg.w, "-", e.Pokemon.Name)
	}

	return nil
}
