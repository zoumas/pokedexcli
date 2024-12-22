package pokeapi

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/zoumas/pokedexcli/internal/cache"
)

type LocationAreas struct {
	Previous string `json:"previous"`
	Next     string `json:"next"`
	Results  []struct {
		Name string `json:"name"`
	} `json:"results"`
}

func GetLocationAreas(c *cache.Cache, url string) (*LocationAreas, error) {
	value, ok := c.Get(url)
	if !ok {
		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		value, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		c.Add(url, value)
	}

	var l LocationAreas
	err := json.Unmarshal(value, &l)
	if err != nil {
		return nil, err
	}
	return &l, err
}

type LocationArea struct {
	PokemonEncounters []struct {
		Pokemon struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"pokemon"`
	} `json:"pokemon_encounters"`
}

func GetLocationArea(c *cache.Cache, name string) (*LocationArea, error) {
	url := "https://pokeapi.co/api/v2/location-area/" + name

	value, ok := c.Get(url)
	if !ok {
		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		value, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		c.Add(url, value)
	}

	var l LocationArea
	err := json.Unmarshal(value, &l)
	if err != nil {
		return nil, err
	}
	return &l, nil
}

type Pokemon struct {
	Name    string `json:"name"`
	Species struct {
		Name string `json:"name"`
	} `json:"species"`
	Types []struct {
		Type struct {
			Name string `json:"name"`
		} `json:"type"`
	} `json:"types"`
	Stats []struct {
		Stat struct {
			Name string `json:"name"`
		} `json:"stat"`
		BaseStat int `json:"base_stat"`
	} `json:"stats"`
	Height         int `json:"height"`
	Order          int `json:"order"`
	BaseExperience int `json:"base_experience"`
	Weight         int `json:"weight"`
}

func GetPokemon(name string) (*Pokemon, error) {
	url := "https://pokeapi.co/api/v2/pokemon/" + name
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var p Pokemon
	err = json.NewDecoder(resp.Body).Decode(&p)
	if err != nil {
		return nil, err
	}
	return &p, nil
}
