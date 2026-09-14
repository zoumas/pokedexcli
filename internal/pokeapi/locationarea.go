// Package pokeapi contains functions and types for interacting with the
// PokéAPI at https://pokeapi.co/.
package pokeapi

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// locationAreaPath is the location-area endpoint, relative to a Client's base URL.
const locationAreaPath = "location-area"

// LocationAreasURL returns the URL of the first page of location areas.
func (c *Client) LocationAreasURL() string {
	return strings.TrimSuffix(c.baseURL, "/") + "/" + locationAreaPath + "/?offset=0&limit=20"
}

// NamedResource is PokeAPI's NamedAPIResource: the name of a resource and the
// URL to fetch it from.
type NamedResource struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// LocationAreas is one page of the location-area endpoint. Next and Previous
// are nil on the last and first page respectively.
type LocationAreas struct {
	Count    int             `json:"count"`
	Next     *string         `json:"next"`
	Previous *string         `json:"previous"`
	Results  []NamedResource `json:"results"`
}

// GetLocationAreas fetches the page of location areas at url.
func (c *Client) GetLocationAreas(url string) (*LocationAreas, error) {
	data, err := c.get(url)
	if err != nil {
		return nil, fmt.Errorf("getting location areas: %w", err)
	}

	var locationAreas LocationAreas
	if err := json.Unmarshal(data, &locationAreas); err != nil {
		return nil, fmt.Errorf("unmarshalling location areas: %w", err)
	}
	return &locationAreas, nil
}

// LocationAreaDetail is a single location area, including which Pokemon are
// encountered there. Fields of the endpoint that this program does not use are
// omitted; encoding/json ignores them.
type LocationAreaDetail struct {
	Name              string             `json:"name"`
	PokemonEncounters []PokemonEncounter `json:"pokemon_encounters"`
}

// PokemonEncounter is one Pokemon that can be encountered in a location area.
type PokemonEncounter struct {
	Pokemon NamedResource `json:"pokemon"`
}

// GetLocationArea fetches the location area with the given name or id.
func (c *Client) GetLocationArea(name string) (*LocationAreaDetail, error) {
	url, err := url.JoinPath(c.baseURL, locationAreaPath, name)
	if err != nil {
		return nil, fmt.Errorf("building location area URL for %q: %w", name, err)
	}

	data, err := c.get(url)
	if err != nil {
		return nil, fmt.Errorf("getting location area %q: %w", name, err)
	}

	var detail LocationAreaDetail
	if err := json.Unmarshal(data, &detail); err != nil {
		return nil, fmt.Errorf("unmarshalling location area %q: %w", name, err)
	}
	return &detail, nil
}
