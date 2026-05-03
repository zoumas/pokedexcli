package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
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
		client:   client,
		mapCache: pokecache.NewCache[pokeapi.LocationArea](5 * time.Minute),
		next:     startURL,
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

func Test_exploreCommand_requiresLocationAreaName(t *testing.T) {
	commands := getCommands(bytes.NewBuffer(nil))

	err := commands["explore"].callback(bytes.NewBuffer(nil))
	if err == nil {
		t.Fatal("explore command callback() error = nil, want error")
	}

	if got, want := err.Error(), "explore command requires a location area name"; got != want {
		t.Fatalf("explore command callback() error = %q, want %q", got, want)
	}
}

func Test_commandExplore_printsPokemonAndUsesCache(t *testing.T) {
	const areaName = "test-area"

	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++

		if got, want := r.URL.Path, "/api/v2/location-area/"+areaName+"/"; got != want {
			t.Fatalf("request path = %q, want %q", got, want)
		}

		_, _ = w.Write([]byte(`{
			"pokemon_encounters": [
				{"pokemon": {"name": "pikachu"}},
				{"pokemon": {"name": "bulbasaur"}}
			]
		}`))
	}))
	defer server.Close()

	client := serverRewriteClient(t, server.URL)
	cfg := &Config{
		client:       client,
		exploreCache: pokecache.NewCache[pokeapi.ExploredLocationArea](5 * time.Minute),
	}

	want := "Exploring test-area...\n" +
		"Found Pokemon:\n" +
		" - pikachu\n" +
		" - bulbasaur\n"

	for i := range 2 {
		var out bytes.Buffer
		if err := commandExplore(&out, cfg, areaName); err != nil {
			t.Fatalf("commandExplore() run %d returned an error: %v", i+1, err)
		}

		if got := out.String(); got != want {
			t.Fatalf("commandExplore() run %d output = %q, want %q", i+1, got, want)
		}
	}

	if got, want := requests, 1; got != want {
		t.Fatalf("request count = %d, want %d", got, want)
	}
}

func lines(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return strings.Join(values, "\n") + "\n"
}

func serverRewriteClient(t *testing.T, serverURL string) *http.Client {
	t.Helper()

	baseURL, err := url.Parse(serverURL)
	if err != nil {
		t.Fatalf("url.Parse(%q) returned an error: %v", serverURL, err)
	}

	transport := http.DefaultTransport
	return &http.Client{
		Timeout: 10 * time.Second,
		Transport: rewriteHostTransport{
			baseURL:   baseURL,
			transport: transport,
		},
	}
}

type rewriteHostTransport struct {
	baseURL   *url.URL
	transport http.RoundTripper
}

func (t rewriteHostTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	clone.URL.Scheme = t.baseURL.Scheme
	clone.URL.Host = t.baseURL.Host

	resp, err := t.transport.RoundTrip(clone)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
