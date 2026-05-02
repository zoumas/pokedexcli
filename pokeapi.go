package main

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

func getLocationArea(c *http.Client, url string) (area LocationArea, err error) {
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
