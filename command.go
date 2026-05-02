package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"slices"
	"time"
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

	cfg := &Config{
		client:   &http.Client{Timeout: 10 * time.Second},
		next:     "https://pokeapi.co/api/v2/location-area/",
		previous: nil,
	}

	commands["map"] = command{
		name:        "map",
		description: "Displays a list of location areas",
		callback: func(io.Writer) error {
			return commandMap(w, cfg, cfg.next)
		},
	}

	commands["mapb"] = command{
		name:        "mapb",
		description: "Displays a list of location areas in the previous page",
		callback: func(io.Writer) error {
			return commandMapb(w, cfg)
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

type Config struct {
	client   *http.Client
	next     string
	previous *string
}

func commandMap(w io.Writer, cfg *Config, url string) error {
	log.Printf("Fetching location areas from %q\n", url)

	area, err := getLocationArea(cfg.client, url)
	if err != nil {
		return err
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
