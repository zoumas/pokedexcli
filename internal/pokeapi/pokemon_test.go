package pokeapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
)

const pidgeyBody = `{
  "name": "pidgey",
  "base_experience": 50,
  "height": 3,
  "weight": 18,
  "stats": [
    {"base_stat": 40, "stat": {"name": "hp", "url": "https://example.test/stat/1/"}},
    {"base_stat": 45, "stat": {"name": "attack", "url": "https://example.test/stat/2/"}}
  ],
  "types": [
    {"type": {"name": "normal", "url": "https://example.test/type/1/"}},
    {"type": {"name": "flying", "url": "https://example.test/type/3/"}}
  ]
}`

func TestGetPokemon(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(pidgeyBody))
	}))
	defer server.Close()

	client := newTestClient(t, server.Client(), server.URL)

	got, err := client.GetPokemon("pidgey")
	if err != nil {
		t.Fatalf("GetPokemon(%q) error = %v, want nil", "pidgey", err)
	}

	want := &Pokemon{
		Name:           "pidgey",
		BaseExperience: 50,
		Height:         3,
		Weight:         18,
		Stats: []PokemonStat{
			{BaseStat: 40, Stat: NamedResource{Name: "hp", URL: "https://example.test/stat/1/"}},
			{BaseStat: 45, Stat: NamedResource{Name: "attack", URL: "https://example.test/stat/2/"}},
		},
		Types: []PokemonType{
			{Type: NamedResource{Name: "normal", URL: "https://example.test/type/1/"}},
			{Type: NamedResource{Name: "flying", URL: "https://example.test/type/3/"}},
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("GetPokemon(%q) diff (-want +got):\n%s", "pidgey", diff)
	}
}

func TestGetPokemonNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := newTestClient(t, server.Client(), server.URL)

	if _, err := client.GetPokemon("notarealmon"); err == nil {
		t.Errorf("GetPokemon(%q) error = nil, want an error for status 404", "notarealmon")
	}
}

func TestGetPokemonLive(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live PokeAPI call in short mode")
	}

	client := newTestClient(t, &http.Client{}, DefaultBaseURL)

	got, err := client.GetPokemon("pidgey")
	if err != nil {
		t.Fatalf("GetPokemon(%q) error = %v, want nil", "pidgey", err)
	}
	if got.Name != "pidgey" {
		t.Errorf("GetPokemon(%q) Name = %q, want %q", "pidgey", got.Name, "pidgey")
	}
	if got.BaseExperience <= 0 {
		t.Errorf("GetPokemon(%q) BaseExperience = %d, want a positive value", "pidgey", got.BaseExperience)
	}
}
