package main

import (
	"bytes"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/zoumas/pokedexcli/internal/pokeapi"
	"github.com/zoumas/pokedexcli/internal/pokecache"
)

func Test_mapCommands_matchFirstThreeLivePages(t *testing.T) {
	client := &http.Client{Timeout: 10 * time.Second}
	startURL := "https://pokeapi.co/api/v2/location-area/"

	expectedPages := make([][]string, 0, 3)
	nextURL := startURL
	for range 3 {
		area, err := pokeapi.GetLocationArea(client, nextURL)
		if err != nil {
			t.Fatalf("GetLocationArea(%q) returned an error: %v", nextURL, err)
		}

		names := make([]string, 0, len(area.Results))
		for _, result := range area.Results {
			names = append(names, result.Name)
		}
		expectedPages = append(expectedPages, names)
		nextURL = area.Next
	}

	cfg := &Config{
		client:        client,
		locationCache: pokecache.NewCache[pokeapi.LocationArea](5 * time.Minute),
		next:          startURL,
	}

	for i, expectedNames := range expectedPages {
		var out bytes.Buffer
		if err := commandMap(&out, cfg, cfg.next); err != nil {
			t.Fatalf("commandMap() for page %d returned an error: %v", i+1, err)
		}

		if got, want := out.String(), lines(expectedNames); got != want {
			t.Fatalf("commandMap() page %d output = %q, want %q", i+1, got, want)
		}
	}

	for i := len(expectedPages) - 2; i >= 0; i-- {
		var out bytes.Buffer
		if err := commandMapb(&out, cfg); err != nil {
			t.Fatalf("commandMapb() returning to page %d returned an error: %v", i+1, err)
		}

		if got, want := out.String(), lines(expectedPages[i]); got != want {
			t.Fatalf("commandMapb() page %d output = %q, want %q", i+1, got, want)
		}
	}
}

func lines(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return strings.Join(values, "\n") + "\n"
}
