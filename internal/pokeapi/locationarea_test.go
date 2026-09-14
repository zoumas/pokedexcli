package pokeapi

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/zoumas/pokedexcli/internal/pokecache"
)

const firstPageBody = `{
  "count": 3,
  "next": "PLACEHOLDER/api/v2/location-area/?offset=2&limit=2",
  "previous": null,
  "results": [
    {"name": "canalave-city-area", "url": "https://pokeapi.co/api/v2/location-area/1/"},
    {"name": "eterna-city-area", "url": "https://pokeapi.co/api/v2/location-area/2/"}
  ]
}`

const lastPageBody = `{
  "count": 3,
  "next": null,
  "previous": "https://pokeapi.co/api/v2/location-area/?offset=0&limit=2",
  "results": [
    {"name": "pastoria-city-area", "url": "https://pokeapi.co/api/v2/location-area/3/"}
  ]
}`

// newTestClient returns a Client whose cache lives only for the duration of t.
func newTestClient(t *testing.T, httpClient *http.Client) *Client {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(t.Output(), &slog.HandlerOptions{Level: slog.LevelDebug}))
	return New(httpClient, pokecache.New(t.Context(), logger, 5*time.Second))
}

func TestGetLocationAreas(t *testing.T) {
	cases := []struct {
		name        string
		body        string
		wantNames   []string
		wantNextNil bool
		wantPrevNil bool
	}{
		{
			name:        "first page has a next and no previous",
			body:        firstPageBody,
			wantNames:   []string{"canalave-city-area", "eterna-city-area"},
			wantNextNil: false,
			wantPrevNil: true,
		},
		{
			name:        "last page has a previous and no next",
			body:        lastPageBody,
			wantNames:   []string{"pastoria-city-area"},
			wantNextNil: true,
			wantPrevNil: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte(c.body))
			}))
			defer server.Close()

			client := newTestClient(t, server.Client())

			areas, err := client.GetLocationAreas(server.URL)
			if err != nil {
				t.Fatalf("GetLocationAreas(%s) error = %v, want nil", server.URL, err)
			}

			gotNames := make([]string, 0, len(areas.Results))
			for _, a := range areas.Results {
				gotNames = append(gotNames, a.Name)
			}
			if diff := cmp.Diff(c.wantNames, gotNames); diff != "" {
				t.Errorf("GetLocationAreas(%s) names diff (-want +got):\n%s", server.URL, diff)
			}
			if got := areas.Next == nil; got != c.wantNextNil {
				t.Errorf("GetLocationAreas(%s) Next == nil is %v, want %v", server.URL, got, c.wantNextNil)
			}
			if got := areas.Previous == nil; got != c.wantPrevNil {
				t.Errorf("GetLocationAreas(%s) Previous == nil is %v, want %v", server.URL, got, c.wantPrevNil)
			}
		})
	}
}

func TestGetLocationAreasNonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"detail": "throttled"}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.Client())

	areas, err := client.GetLocationAreas(server.URL)

	if err == nil {
		t.Fatalf("GetLocationAreas(%s) error = nil, want an error for status 429", server.URL)
	}
	if areas != nil {
		t.Errorf("GetLocationAreas(%s) areas = %v, want nil", server.URL, areas)
	}
}

func TestGetLocationAreasBadJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("not json"))
	}))
	defer server.Close()

	client := newTestClient(t, server.Client())

	if _, err := client.GetLocationAreas(server.URL); err == nil {
		t.Errorf("GetLocationAreas(%s) error = nil, want a decode error", server.URL)
	}
}

func TestGetLocationAreasLive(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live PokeAPI call in short mode")
	}

	client := newTestClient(t, &http.Client{})

	areas, err := client.GetLocationAreas(StartingLocationAreasURL)
	if err != nil {
		t.Fatalf("GetLocationAreas(%s) error = %v, want nil", StartingLocationAreasURL, err)
	}
	if len(areas.Results) != 20 {
		t.Errorf("GetLocationAreas(%s) returned %d results, want 20", StartingLocationAreasURL, len(areas.Results))
	}
	if areas.Next == nil {
		t.Errorf("GetLocationAreas(%s) Next = nil, want a next page URL", StartingLocationAreasURL)
	}
}

func TestGetLocationAreasUsesCache(t *testing.T) {
	var requests int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests++
		_, _ = w.Write([]byte(firstPageBody))
	}))
	defer server.Close()

	client := newTestClient(t, server.Client())

	first, err := client.GetLocationAreas(server.URL)
	if err != nil {
		t.Fatalf("GetLocationAreas(%s) first call error = %v, want nil", server.URL, err)
	}
	second, err := client.GetLocationAreas(server.URL)
	if err != nil {
		t.Fatalf("GetLocationAreas(%s) second call error = %v, want nil", server.URL, err)
	}

	if requests != 1 {
		t.Errorf("GetLocationAreas(%s) made %d requests for 2 calls, want 1", server.URL, requests)
	}
	if diff := cmp.Diff(first, second); diff != "" {
		t.Errorf("GetLocationAreas(%s) cached result diff (-first +second):\n%s", server.URL, diff)
	}
}

func TestGetLocationAreasDoesNotCacheFailures(t *testing.T) {
	var requests int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests++
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	client := newTestClient(t, server.Client())

	for i := range 2 {
		if _, err := client.GetLocationAreas(server.URL); err == nil {
			t.Fatalf("GetLocationAreas(%s) call %d error = nil, want an error", server.URL, i+1)
		}
	}

	if requests != 2 {
		t.Errorf("GetLocationAreas(%s) made %d requests for 2 failing calls, want 2", server.URL, requests)
	}
}
