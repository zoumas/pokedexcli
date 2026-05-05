package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
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

type Config struct {
	client       *http.Client
	mapCache     *pokecache.Cache[pokeapi.LocationArea]
	exploreCache *pokecache.Cache[pokeapi.ExploredLocationArea]
	pokemonCache *pokecache.Cache[pokeapi.Pokemon]
	pokedex      map[string]pokeapi.Pokemon
	randIntn     func(n int) int
	next         string
	previous     *string
}

func NewConfig() *Config {
	return &Config{
		client:       &http.Client{Timeout: 10 * time.Second},
		mapCache:     pokecache.NewCache[pokeapi.LocationArea](5 * time.Minute),
		exploreCache: pokecache.NewCache[pokeapi.ExploredLocationArea](5 * time.Minute),
		pokemonCache: pokecache.NewCache[pokeapi.Pokemon](5 * time.Minute),
		pokedex:      make(map[string]pokeapi.Pokemon),
		randIntn:     rand.New(rand.NewSource(time.Now().UnixNano())).Intn,
		next:         "https://pokeapi.co/api/v2/location-area/?offset=0&limit=20",
		previous:     nil,
	}
}

func getCommands(w io.Writer) map[string]command {
	commands := map[string]command{}
	cfg := NewConfig()

	commands["exit"] = command{
		name:        "exit",
		description: "Exit the Pokedex",
		callback: func(io.Writer, ...string) error {
			return commandExit(w, os.Exit)
		},
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

	commands["catch"] = command{
		name:        "catch",
		description: "Attempts to catch a Pokemon by name",
		callback: func(w io.Writer, args ...string) error {
			if len(args) == 0 {
				return errors.New("catch command requires a Pokemon name")
			}
			name := args[0]
			return commandCatch(w, cfg, name)
		},
	}

	commands["inspect"] = command{
		name:        "inspect",
		description: "Desplays detailed information about a Pokemon in your Pokedex by name",
		callback: func(w io.Writer, args ...string) error {
			if len(args) == 0 {
				return errors.New("inspect command requires a Pokemon name")
			}
			name := args[0]
			return commandInspect(w, cfg, name)
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

func commandCatch(w io.Writer, cfg *Config, name string) error {
	_, _ = fmt.Fprintf(w, "Throwing a Pokeball at %s...\n", name)

	url := fmt.Sprintf("https://pokeapi.co/api/v2/pokemon/%s/", name)

	pokemon, ok := cfg.pokemonCache.Get(url)
	if !ok {
		log.Printf("Cache miss for %q\n", url)
		var err error
		pokemon, err = pokeapi.GetPokemon(cfg.client, url)
		if err != nil {
			return fmt.Errorf("failed to catch %s: %w", name, err)
		}
		cfg.pokemonCache.Add(url, pokemon)
	} else {
		log.Printf("Cache hit for %q\n", url)
	}

	catchChance := catchChanceForBaseExperience(pokemon.BaseExperience)
	log.Printf("Catch chance for %s is %d%%\n", name, catchChance)
	randIntn := cfg.randIntn
	if randIntn == nil {
		randIntn = rand.New(rand.NewSource(time.Now().UnixNano())).Intn
	}
	roll := randIntn(100)
	caught := roll < catchChance
	log.Printf("Rolled a %d, caught: %t\n", roll, caught)
	if caught {
		_, _ = fmt.Fprintf(w, "%s was caught!\n", name)
		cfg.pokedex[name] = pokemon
		log.Printf("%v\n", cfg.pokedex)
	} else {
		_, _ = fmt.Fprintf(w, "%s escaped!\n", name)
	}

	return nil
}

func catchChanceForBaseExperience(baseExperience int) int {
	catchChance := 90 - baseExperience/4
	return max(10, catchChance)
}

func commandInspect(w io.Writer, cfg *Config, name string) error {
	pokemon, ok := cfg.pokedex[name]
	if !ok {
		_, _ = fmt.Fprintf(w, "You don't have %s in your Pokedex. Try catching it first!\n", name)
		return nil
	}

	_, _ = fmt.Fprintln(w, "Name:", pokemon.Name)
	_, _ = fmt.Fprintln(w, "Height:", pokemon.Height)
	_, _ = fmt.Fprintln(w, "Weight:", pokemon.Weight)
	_, _ = fmt.Fprintln(w, "Stats:")
	for _, stat := range pokemon.Stats {
		_, _ = fmt.Fprintf(w, " - %s: %d\n", stat.Stat.Name, stat.BaseStat)
	}
	_, _ = fmt.Fprintln(w, "Types:")
	for _, t := range pokemon.Types {
		_, _ = fmt.Fprintf(w, " - %s\n", t.Type.Name)
	}

	return nil
}
