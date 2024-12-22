package cache

import (
	"fmt"
	"testing"
	"time"
)

func TestCache(t *testing.T) {
	cases := []struct {
		key   string
		value []byte
	}{
		{
			key:   "https://example.com",
			value: []byte("testdata"),
		},
		{
			key:   "https://example.com/path",
			value: []byte("moretestdata"),
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			c := NewCache(1 * time.Second)
			c.Add(tc.key, tc.value)
			value, ok := c.Get(tc.key)
			if !ok {
				t.Fatal("expected to find key")
			}

			expected := string(tc.value)
			actual := string(value)
			if expected != actual {
				t.Fatalf("\nkey: %q\nexpected: %q\nactual: %q",
					tc.key, expected, actual)
			}
		})
	}

	t.Run("reap", func(t *testing.T) {
		const baseTime = 5 * time.Millisecond
		const waitTime = 2 * baseTime
		c := NewCache(baseTime)

		c.Add("https://example.com", []byte("testdata"))
		_, ok := c.Get("https://example.com")
		if !ok {
			t.Fatal("expected to find key")
		}
		time.Sleep(waitTime)
		_, ok = c.Get("https://example.com")
		if ok {
			t.Fatal("expected to not find key")
		}
	})
}
