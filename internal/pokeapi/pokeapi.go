// Package pokeapi provides functions to interact with the PokeAPI.
package pokeapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type LocationArea struct {
	// Count    int    `json:"count"`
	Next     string  `json:"next"`
	Previous *string `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
		// URL  string `json:"url"`
	} `json:"results"`
}

func GetLocationArea(c *http.Client, url string) (area LocationArea, err error) {
	resp, err := c.Get(url)
	if err != nil {
		return LocationArea{}, fmt.Errorf("failed to get location areas: %w", err)
	}
	defer func() {
		err = errors.Join(err, resp.Body.Close())
	}()

	if resp.StatusCode != http.StatusOK {
		return LocationArea{}, fmt.Errorf("failed to get location areas: unexpected status code: %s", resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(&area); err != nil {
		return LocationArea{}, fmt.Errorf("failed to decode location areas: %w", err)
	}

	return area, nil
}

type ExploredLocationArea struct {
	// ID       int `json:"id"`
	// Name string `json:"name"`
	PokemonEncounters []struct {
		Pokemon struct {
			Name string `json:"name"`
			// URL  string `json:"url"`
		} `json:"pokemon"`
	} `json:"pokemon_encounters"`
}

func ExploreLocationArea(c *http.Client, url string) (area ExploredLocationArea, err error) {
	resp, err := c.Get(url)
	if err != nil {
		return ExploredLocationArea{}, fmt.Errorf("failed to explore location area: %w", err)
	}
	defer func() {
		err = errors.Join(err, resp.Body.Close())
	}()

	if resp.StatusCode != http.StatusOK {
		return ExploredLocationArea{}, fmt.Errorf("failed to explore location area: unexpected status code: %s", resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(&area); err != nil {
		return ExploredLocationArea{}, fmt.Errorf("failed to decode explored location area: %w", err)
	}

	return area, nil
}

type Pokemon struct {
	BaseExperience int `json:"base_experience"`

	Height int `json:"height"`
	Weight int `json:"weight"`

	ID   int    `json:"id"`
	Name string `json:"name"`

	Stats []struct {
		BaseStat int `json:"base_stat"`
		// Effort   int `json:"effort"`
		Stat struct {
			Name string `json:"name"`
			// URL  string `json:"url"`
		} `json:"stat"`
	} `json:"stats"`

	Types []struct {
		// Slot int `json:"slot"`
		Type struct {
			Name string `json:"name"`
			// URL  string `json:"url"`
		} `json:"type"`
	} `json:"types"`
}

func GetPokemon(c *http.Client, url string) (pokemon Pokemon, err error) {
	resp, err := c.Get(url)
	if err != nil {
		return Pokemon{}, fmt.Errorf("failed to get pokemon: %w", err)
	}
	defer func() {
		err = errors.Join(err, resp.Body.Close())
	}()

	if resp.StatusCode != http.StatusOK {
		return Pokemon{}, fmt.Errorf("failed to get pokemon: unexpected status code: %s", resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(&pokemon); err != nil {
		return Pokemon{}, fmt.Errorf("failed to decode pokemon: %w", err)
	}

	return pokemon, nil
}
