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
		url  string // `json:"url"`
	} `json:"results"`
	count int // `json:"count"`
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
