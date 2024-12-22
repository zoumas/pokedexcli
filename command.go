package main

import (
	"cmp"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"slices"

	"github.com/zoumas/pokedexcli/internal/cache"
	"github.com/zoumas/pokedexcli/internal/pokeapi"
)

type config struct {
	w        io.Writer
	cache    *cache.Cache
	pokemon  map[string]*pokeapi.Pokemon
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
		"catch": {
			callback:    commandCatch,
			name:        "catch",
			description: "Attempt to catch a wild Pokemon",
		},
		"inspect": {
			callback:    commandInspect,
			name:        "inspect",
			description: "See information about a caught Pokemon",
		},
		"pokedex": {
			callback:    commandPokedex,
			name:        "pokedex",
			description: "List the names of all caught Pokemon",
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

func commandCatch(cfg *config) error {
	if len(cfg.args) == 0 {
		return fmt.Errorf("There is nothing to catch...")
	}
	name := cfg.args[0]

	p, err := pokeapi.GetPokemon(name)
	if err != nil {
		return errors.New("An unseen force prevents you from throwing a Pokeball...")
	}

	fmt.Fprintf(cfg.w, "Throwing a Pokeball at %s...\n", name)
	cfg.pokemon[name] = p
	fmt.Fprintf(cfg.w, "%s has been caught!\n", name)

	return nil
}

func commandInspect(cfg *config) error {
	if len(cfg.args) == 0 {
		return fmt.Errorf("There is nothing to inspect...")
	}
	name := cfg.args[0]

	p, ok := cfg.pokemon[name]
	if !ok {
		return fmt.Errorf("You haven't caught %s yet!", name)
	}

	fmt.Fprintf(cfg.w, "Name: %s\n", p.Name)
	fmt.Fprintf(cfg.w, "Height: %d\n", p.Height)
	fmt.Fprintf(cfg.w, "Weight: %d\n", p.Weight)
	fmt.Fprintf(cfg.w, "Stats:\n")
	for _, s := range p.Stats {
		fmt.Fprintf(cfg.w, "\t- %s: %d\n", s.Stat.Name, s.BaseStat)
	}
	fmt.Fprintf(cfg.w, "Types:\n")
	for _, t := range p.Types {
		fmt.Fprintf(cfg.w, "\t- %s\n", t.Type.Name)
	}
	return nil
}

func commandPokedex(cfg *config) error {
	xs := slices.Collect(maps.Values(cfg.pokemon))
	slices.SortStableFunc(xs, func(a, b *pokeapi.Pokemon) int {
		return cmp.Compare(a.Order, b.Order)
	})
	for _, x := range xs {
		fmt.Fprintf(cfg.w, "%4d. %s\n", x.Order, x.Name)
	}
	return nil
}
