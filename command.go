package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"slices"
	"time"

	"github.com/zoumas/pokedexcli/internal/pokeapi"
	"github.com/zoumas/pokedexcli/internal/pokecache"
)

type command struct {
	name        string
	description string
	callback    func(w io.Writer, args ...string) error
}

func getCommands(w io.Writer) map[string]command {
	commands := map[string]command{}

	commands["exit"] = command{
		name:        "exit",
		description: "Exit the Pokedex",
		callback: func(io.Writer, ...string) error {
			return commandExit(w, os.Exit)
		},
	}

	cfg := &Config{
		client:       &http.Client{Timeout: 10 * time.Second},
		mapCache:     pokecache.NewCache[pokeapi.LocationArea](5 * time.Minute),
		exploreCache: pokecache.NewCache[pokeapi.ExploredLocationArea](5 * time.Minute),
		next:         "https://pokeapi.co/api/v2/location-area/?offset=0&limit=20",
		previous:     nil,
	}

	commands["map"] = command{
		name:        "map",
		description: "Displays a list of location areas",
		callback: func(io.Writer, ...string) error {
			return commandMap(w, cfg, cfg.next)
		},
	}

	commands["mapb"] = command{
		name:        "mapb",
		description: "Displays a list of location areas in the previous page",
		callback: func(io.Writer, ...string) error {
			return commandMapb(w, cfg)
		},
	}

	commands["explore"] = command{
		name:        "explore",
		description: "Explores a location area and lists the Pokemon that can be found there",
		callback: func(w io.Writer, args ...string) error {
			if len(args) == 0 {
				return errors.New("explore command requires a location area name")
			}
			name := args[0]
			return commandExplore(w, cfg, name)
		},
	}

	commands["help"] = command{
		name:        "help",
		description: "Displays a help message",
		callback: func(io.Writer, ...string) error {
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

type Config struct {
	client       *http.Client
	mapCache     *pokecache.Cache[pokeapi.LocationArea]
	exploreCache *pokecache.Cache[pokeapi.ExploredLocationArea]
	next         string
	previous     *string
}

func commandMap(w io.Writer, cfg *Config, url string) error {
	log.Printf("Fetching location areas from %q\n", url)

	area, ok := cfg.mapCache.Get(url)
	if !ok {
		log.Printf("Cache miss for %q\n", url)
		var err error
		area, err = pokeapi.GetLocationArea(cfg.client, url)
		if err != nil {
			return err
		}
		cfg.mapCache.Add(url, area)
	} else {
		log.Printf("Cache hit for %q\n", url)
	}

	cfg.next = area.Next
	cfg.previous = area.Previous

	log.Println("next:", cfg.next)
	if cfg.previous != nil {
		log.Println("previous:", *cfg.previous)
	}

	for _, r := range area.Results {
		_, _ = fmt.Fprintln(w, r.Name)
	}

	return nil
}

func commandMapb(w io.Writer, cfg *Config) error {
	if cfg.previous == nil {
		_, _ = fmt.Fprintln(w, "There is nothing back there...")
		return nil
	}
	url := *cfg.previous
	return commandMap(w, cfg, url)
}

func commandExplore(w io.Writer, cfg *Config, name string) error {
	_, _ = fmt.Fprintf(w, "Exploring %s...\n", name)

	url := fmt.Sprintf("https://pokeapi.co/api/v2/location-area/%s/", name)
	pokemonArea, ok := cfg.exploreCache.Get(url)
	if !ok {
		log.Printf("Cache miss for %q\n", url)
		var err error
		pokemonArea, err = pokeapi.ExploreLocationArea(cfg.client, url)
		if err != nil {
			return err
		}
		cfg.exploreCache.Add(url, pokemonArea)
	} else {
		log.Printf("Cache hit for %q\n", url)
	}

	_, _ = fmt.Fprintln(w, "Found Pokemon:")
	for _, encounter := range pokemonArea.PokemonEncounters {
		name := encounter.Pokemon.Name
		_, _ = fmt.Fprintf(w, " - %s\n", name)
	}

	return nil
}
