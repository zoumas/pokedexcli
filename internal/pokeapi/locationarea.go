// Package pokeapi contains functions and types for interacting with the
// PokéAPI at https://pokeapi.co/.
package pokeapi

import (
	"encoding/json"
	"fmt"
	"net/http"
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

// GetLocationAreas fetches the page of location areas at url using client.
func GetLocationAreas(client *http.Client, url string) (*LocationAreas, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("getting location areas: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("getting location areas: unexpected status %s", resp.Status)
	}

	var locationAreas LocationAreas
	if err := json.NewDecoder(resp.Body).Decode(&locationAreas); err != nil {
		return nil, fmt.Errorf("decoding location areas: %w", err)
	}
	return &locationAreas, nil
}
