package pokecache_test

import (
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/zoumas/pokedexcli/internal/pokecache"
)

func TestCache(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cache := pokecache.New(t.Context(), logger, 5*time.Second)

	k1 := "k1"
	v1 := []byte("value 1")

	cache.Add(k1, v1)

	k2 := "k2"

	gotV1, ok := cache.Get(k1)
	if !ok {
		t.Errorf("Get(%q): key not found", k1)
	}
	if got, want := string(gotV1), string(v1); got != want {
		t.Errorf("Get(%q) = %q, want %q", k1, got, want)
	}

	_, ok = cache.Get(k2)
	if ok {
		t.Errorf("Get(%q): unexpected key found", k2)
	}
}

func TestCache_reapLoop(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(t.Output(), &slog.HandlerOptions{Level: slog.LevelDebug}))

	interval := 10 * time.Millisecond
	cache := pokecache.New(t.Context(), logger, interval)

	k1 := "k1"
	v1 := []byte("value 1")

	cache.Add(k1, v1)

	gotV1, ok := cache.Get(k1)
	if !ok {
		t.Errorf("Get(%q): key not found", k1)
	}
	if got, want := string(gotV1), string(v1); got != want {
		t.Errorf("Get(%q) = %q, want %q", k1, got, want)
	}

	time.Sleep(2 * interval)

	_, ok = cache.Get(k1)
	if ok {
		t.Errorf("Get(%q): unexpected key found", k1)
	}
}
