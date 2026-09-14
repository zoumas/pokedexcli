package pokeapi

import (
	"fmt"
	"io"
	"net/http"

	"github.com/zoumas/pokedexcli/internal/pokecache"
)

// Client fetches PokeAPI resources over HTTP, serving repeated requests for the
// same URL from its cache.
type Client struct {
	httpClient *http.Client
	cache      *pokecache.Cache
}

// New returns a Client that makes requests with httpClient and stores response
// bodies in cache.
func New(httpClient *http.Client, cache *pokecache.Cache) *Client {
	return &Client{
		httpClient: httpClient,
		cache:      cache,
	}
}

// get returns the body of a GET request to url, serving it from the cache when
// present and caching it otherwise. Only successful responses are cached.
func (c *Client) get(url string) ([]byte, error) {
	if data, ok := c.cache.Get(url); ok {
		return data, nil
	}

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %s", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	c.cache.Add(url, data)
	return data, nil
}
