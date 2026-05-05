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

func Test_catchCommand_requiresPokemonName(t *testing.T) {
	commands := getCommands(bytes.NewBuffer(nil))

	err := commands["catch"].callback(bytes.NewBuffer(nil))
	if err == nil {
		t.Fatal("catch command callback() error = nil, want error")
	}

	if got, want := err.Error(), "catch command requires a Pokemon name"; got != want {
		t.Fatalf("catch command callback() error = %q, want %q", got, want)
	}
}

func Test_inspectCommand_requiresPokemonName(t *testing.T) {
	commands := getCommands(bytes.NewBuffer(nil))

	err := commands["inspect"].callback(bytes.NewBuffer(nil))
	if err == nil {
		t.Fatal("inspect command callback() error = nil, want error")
	}

	if got, want := err.Error(), "inspect command requires a Pokemon name"; got != want {
		t.Fatalf("inspect command callback() error = %q, want %q", got, want)
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

func Test_commandCatch_printsCaughtMessageAndUsesCache(t *testing.T) {
	const pokemonName = "pikachu"

	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++

		if got, want := r.URL.Path, "/api/v2/pokemon/"+pokemonName+"/"; got != want {
			t.Fatalf("request path = %q, want %q", got, want)
		}

		_, _ = w.Write([]byte(`{
			"id": 25,
			"name": "pikachu",
			"base_experience": 20,
			"height": 4,
			"weight": 60,
			"stats": [],
			"types": []
		}`))
	}))
	defer server.Close()

	client := serverRewriteClient(t, server.URL)
	cfg := &Config{
		client:       client,
		pokemonCache: pokecache.NewCache[pokeapi.Pokemon](5 * time.Minute),
		pokedex:      make(map[string]pokeapi.Pokemon),
		randIntn: func(int) int {
			return 10
		},
	}

	want := "Throwing a Pokeball at pikachu...\n" +
		"pikachu was caught!\n" +
		"You may now inspect it with the inspect command.\n"

	for i := range 2 {
		var out bytes.Buffer
		if err := commandCatch(&out, cfg, pokemonName); err != nil {
			t.Fatalf("commandCatch() run %d returned an error: %v", i+1, err)
		}

		if got := out.String(); got != want {
			t.Fatalf("commandCatch() run %d output = %q, want %q", i+1, got, want)
		}
	}

	if got, want := requests, 1; got != want {
		t.Fatalf("request count = %d, want %d", got, want)
	}

	if _, ok := cfg.pokedex[pokemonName]; !ok {
		t.Fatalf("pokedex missing %q after successful catch", pokemonName)
	}
}

func Test_commandCatch_printsEscapeMessageWhenRollMisses(t *testing.T) {
	const pokemonName = "mewtwo"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.Path, "/api/v2/pokemon/"+pokemonName+"/"; got != want {
			t.Fatalf("request path = %q, want %q", got, want)
		}

		_, _ = w.Write([]byte(`{
			"id": 150,
			"name": "mewtwo",
			"base_experience": 200,
			"height": 20,
			"weight": 1220,
			"stats": [],
			"types": []
		}`))
	}))
	defer server.Close()

	client := serverRewriteClient(t, server.URL)
	cfg := &Config{
		client:       client,
		pokemonCache: pokecache.NewCache[pokeapi.Pokemon](5 * time.Minute),
		pokedex:      make(map[string]pokeapi.Pokemon),
		randIntn: func(int) int {
			return 50
		},
	}

	var out bytes.Buffer
	if err := commandCatch(&out, cfg, pokemonName); err != nil {
		t.Fatalf("commandCatch() returned an error: %v", err)
	}

	want := "Throwing a Pokeball at mewtwo...\n" +
		"mewtwo escaped!\n"
	if got := out.String(); got != want {
		t.Fatalf("commandCatch() output = %q, want %q", got, want)
	}

	if _, ok := cfg.pokedex[pokemonName]; ok {
		t.Fatalf("pokedex unexpectedly contains %q after escape", pokemonName)
	}
}

func Test_catchChanceForBaseExperience_higherBaseExperienceIsHarder(t *testing.T) {
	tests := []struct {
		baseExperience int
		want           int
	}{
		{baseExperience: 20, want: 85},
		{baseExperience: 240, want: 30},
		{baseExperience: 324, want: 10},
		{baseExperience: 400, want: 10},
	}

	for _, tt := range tests {
		if got := catchChanceForBaseExperience(tt.baseExperience); got != tt.want {
			t.Fatalf("catchChanceForBaseExperience(%d) = %d, want %d", tt.baseExperience, got, tt.want)
		}
	}
}

func Test_commandInspect_printsNotCaughtMessage(t *testing.T) {
	cfg := &Config{
		pokedex: make(map[string]pokeapi.Pokemon),
	}

	var out bytes.Buffer
	if err := commandInspect(&out, cfg, "pidgey"); err != nil {
		t.Fatalf("commandInspect() returned an error: %v", err)
	}

	want := "You don't have pidgey in your Pokedex. Try catching it first!\n"
	if got := out.String(); got != want {
		t.Fatalf("commandInspect() output = %q, want %q", got, want)
	}
}

func Test_commandInspect_printsPokemonDetailsFromPokedex(t *testing.T) {
	cfg := &Config{
		pokedex: map[string]pokeapi.Pokemon{
			"pidgey": {
				Name:   "pidgey",
				Height: 3,
				Weight: 18,
				Stats: []struct {
					BaseStat int `json:"base_stat"`
					Stat     struct {
						Name string `json:"name"`
					} `json:"stat"`
				}{
					{
						BaseStat: 40,
						Stat: struct {
							Name string `json:"name"`
						}{Name: "hp"},
					},
					{
						BaseStat: 45,
						Stat: struct {
							Name string `json:"name"`
						}{Name: "attack"},
					},
				},
				Types: []struct {
					Type struct {
						Name string `json:"name"`
					} `json:"type"`
				}{
					{
						Type: struct {
							Name string `json:"name"`
						}{Name: "normal"},
					},
					{
						Type: struct {
							Name string `json:"name"`
						}{Name: "flying"},
					},
				},
			},
		},
	}

	var out bytes.Buffer
	if err := commandInspect(&out, cfg, "pidgey"); err != nil {
		t.Fatalf("commandInspect() returned an error: %v", err)
	}

	want := "Name: pidgey\n" +
		"Height: 3\n" +
		"Weight: 18\n" +
		"Stats:\n" +
		" - hp: 40\n" +
		" - attack: 45\n" +
		"Types:\n" +
		" - normal\n" +
		" - flying\n"
	if got := out.String(); got != want {
		t.Fatalf("commandInspect() output = %q, want %q", got, want)
	}
}

func Test_commandPokedex_printsCaughtPokemonInPokedexOrder(t *testing.T) {
	cfg := &Config{
		pokedex: map[string]pokeapi.Pokemon{
			"mew":   {ID: 151, Name: "mew"},
			"abra":  {ID: 63, Name: "abra"},
			"zubat": {ID: 41, Name: "zubat"},
		},
	}

	var out bytes.Buffer
	if err := commandPokedex(&out, cfg); err != nil {
		t.Fatalf("commandPokedex() returned an error: %v", err)
	}

	want := "Your Pokedex:\n" +
		" - zubat\n" +
		" - abra\n" +
		" - mew\n"
	if got := out.String(); got != want {
		t.Fatalf("commandPokedex() output = %q, want %q", got, want)
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
