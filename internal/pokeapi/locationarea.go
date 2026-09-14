// Package pokeapi contains functions and types for interacting with the
// PokéAPI at https://pokeapi.co/.
package pokeapi

import (
	"encoding/json"
	"fmt"
)

// StartingLocationAreasURL is the first page of the location-area endpoint.
const StartingLocationAreasURL = "https://pokeapi.co/api/v2/location-area/?offset=0&limit=20"

// LocationAreas is one page of the location-area endpoint. Next and Previous
// are nil on the last and first page respectively.
type LocationAreas struct {
	Count    int     `json:"count"`
	Next     *string `json:"next"`
	Previous *string `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"results"`
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
