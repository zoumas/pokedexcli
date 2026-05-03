package pokecache

import (
	"fmt"
	"testing"
	"time"
)

func TestCacheAddGet(t *testing.T) {
	const interval = time.Second

	cases := []struct {
		key  string
		want []byte
	}{
		{
			key:  "https://example.com",
			want: []byte("testdata"),
		},
		{
			key:  "https://example.com/path",
			want: []byte("moretestdata"),
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("case_%d", i), func(t *testing.T) {
			cache := NewCache[[]byte](interval)
			cache.Add(tc.key, tc.want)

			got, ok := cache.Get(tc.key)
			if !ok {
				t.Fatalf("Get(%q) did not find cached value", tc.key)
			}

			if string(got) != string(tc.want) {
				t.Fatalf("Get(%q) = %q, want %q", tc.key, got, tc.want)
			}
		})
	}
}

func TestCacheGetReturnsFalseForMissingKey(t *testing.T) {
	cache := NewCache[[]byte](time.Second)

	got, ok := cache.Get("https://example.com/missing")
	if ok {
		t.Fatalf("Get() ok = true, want false")
	}

	if got != nil {
		t.Fatalf("Get() value = %v, want nil", got)
	}
}

func TestCacheReapLoopRemovesExpiredEntries(t *testing.T) {
	const interval = 10 * time.Millisecond

	cache := NewCache[[]byte](interval)
	cache.Add("https://example.com", []byte("testdata"))

	deadline := time.Now().Add(20 * interval)
	for time.Now().Before(deadline) {
		if _, ok := cache.Get("https://example.com"); !ok {
			return
		}

		time.Sleep(interval / 2)
	}

	t.Fatal("expected expired cache entry to be removed")
}
