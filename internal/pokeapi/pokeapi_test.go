package pokeapi_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/zoumas/pokedexcli/internal/pokeapi"
)

func Test_GetLocationArea_returnsErrorForNonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusBadGateway)
	}))
	defer server.Close()

	client := &http.Client{Timeout: 10 * time.Second}

	_, err := pokeapi.GetLocationArea(client, server.URL)
	if err == nil {
		t.Fatal("GetLocationArea() error = nil, want non-nil")
	}

	if !strings.Contains(err.Error(), "unexpected status code: 502 Bad Gateway") {
		t.Fatalf("GetLocationArea() error = %q, want status code details", err)
	}
}
