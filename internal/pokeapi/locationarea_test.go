package pokeapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
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

			areas, err := GetLocationAreas(server.Client(), server.URL)
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

	areas, err := GetLocationAreas(server.Client(), server.URL)

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

	if _, err := GetLocationAreas(server.Client(), server.URL); err == nil {
		t.Errorf("GetLocationAreas(%s) error = nil, want a decode error", server.URL)
	}
}

func TestGetLocationAreasLive(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live PokeAPI call in short mode")
	}

	areas, err := GetLocationAreas(&http.Client{}, StartingLocationAreasURL)
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
