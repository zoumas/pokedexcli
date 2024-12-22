package pokeapi

import (
	"encoding/json"
	"net/http"
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

func GetLocationAreas(url string) (*LocationAreas, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var l LocationAreas
	err = json.NewDecoder(resp.Body).Decode(&l)
	if err != nil {
		return nil, err
	}
	return &l, err
}
