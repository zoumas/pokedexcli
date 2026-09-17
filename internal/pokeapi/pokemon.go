package pokeapi

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// pokemonPath is the pokemon endpoint, relative to a Client's base URL.
const pokemonPath = "pokemon"

// Pokemon is a single Pokemon. Fields of the endpoint that this program does
// not use are omitted; encoding/json ignores them.
type Pokemon struct {
	Name           string        `json:"name"`
	BaseExperience int           `json:"base_experience"`
	Height         int           `json:"height"`
	Weight         int           `json:"weight"`
	Stats          []PokemonStat `json:"stats"`
	Types          []PokemonType `json:"types"`
}

// PokemonStat is one of a Pokemon's base stats.
type PokemonStat struct {
	BaseStat int           `json:"base_stat"`
	Stat     NamedResource `json:"stat"`
}

// PokemonType is one of a Pokemon's types.
type PokemonType struct {
	Type NamedResource `json:"type"`
}

// GetPokemon fetches the Pokemon with the given name or id.
func (c *Client) GetPokemon(name string) (*Pokemon, error) {
	fullURL, err := url.JoinPath(c.baseURL, pokemonPath, name)
	if err != nil {
		return nil, fmt.Errorf("building pokemon URL for %q: %w", name, err)
	}

	data, err := c.get(fullURL)
	if err != nil {
		return nil, fmt.Errorf("getting pokemon %q: %w", name, err)
	}

	var pokemon Pokemon
	if err := json.Unmarshal(data, &pokemon); err != nil {
		return nil, fmt.Errorf("unmarshalling pokemon %q: %w", name, err)
	}
	return &pokemon, nil
}
