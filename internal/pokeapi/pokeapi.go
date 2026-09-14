package pokeapi

import (
	"fmt"
	"io"
	"net/http"

	"github.com/zoumas/pokedexcli/internal/pokecache"
)

// DefaultBaseURL is the root of the public PokéAPI v2.
const DefaultBaseURL = "https://pokeapi.co/api/v2/"

// Client fetches PokeAPI resources over HTTP, serving repeated requests for the
// same URL from its cache.
type Client struct {
	httpClient *http.Client
	cache      *pokecache.Cache
	baseURL    string
}

// New returns a Client that makes requests with httpClient against baseURL and
// stores response bodies in cache. Pass DefaultBaseURL for the public API; a
// test server's URL substitutes for it.
func New(httpClient *http.Client, cache *pokecache.Cache, baseURL string) *Client {
	return &Client{
		httpClient: httpClient,
		cache:      cache,
		baseURL:    baseURL,
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
