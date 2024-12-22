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
		// Url  string `json:"url"`
	} `json:"results"`
	// Count int `json:"count"`
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
	// Location struct {
	// 	Name string `json:"name"`
	// 	URL  string `json:"url"`
	// } `json:"location"`
	PokemonEncounters []struct {
		Pokemon struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"pokemon"`
		// VersionDetails []struct {
		// 	Version struct {
		// 		Name string `json:"name"`
		// 		URL  string `json:"url"`
		// 	} `json:"version"`
		// 	EncounterDetails []struct {
		// 		Method struct {
		// 			Name string `json:"name"`
		// 			URL  string `json:"url"`
		// 		} `json:"method"`
		// 		ConditionValues []any `json:"condition_values"`
		// 		Chance          int   `json:"chance"`
		// 		MaxLevel        int   `json:"max_level"`
		// 		MinLevel        int   `json:"min_level"`
		// 	} `json:"encounter_details"`
		// 	MaxChance int `json:"max_chance"`
		// } `json:"version_details"`
	} `json:"pokemon_encounters"`
	// EncounterMethodRates []struct {
	// 	EncounterMethod struct {
	// 		Name string `json:"name"`
	// 		URL  string `json:"url"`
	// 	} `json:"encounter_method"`
	// 	VersionDetails []struct {
	// 		Version struct {
	// 			Name string `json:"name"`
	// 			URL  string `json:"url"`
	// 		} `json:"version"`
	// 		Rate int `json:"rate"`
	// 	} `json:"version_details"`
	// } `json:"encounter_method_rates"`
	// Names []struct {
	// 	Language struct {
	// 		Name string `json:"name"`
	// 		URL  string `json:"url"`
	// 	} `json:"language"`
	// 	Name string `json:"name"`
	// } `json:"names"`
	// GameIndex int    `json:"game_index"`
	// ID        int    `json:"id"`
	// Name      string `json:"name"`
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
