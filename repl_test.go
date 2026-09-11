package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestCleanInput(t *testing.T) {
	cases := []struct {
		name string
		text string
		want []string
	}{
		{
			name: "surrounding and interior whitespace",
			text: "	hello	world	",
			want: []string{"hello", "world"},
		},
		{
			name: "pokemon names in mixed case return all lower",
			text: "Charmander Bulbasaur PIKACHU",
			want: []string{"charmander", "bulbasaur", "pikachu"},
		},
		{
			name: "empty input returns empty slice",
			text: "",
			want: []string{},
		},
		{
			name: "whitespace only",
			text: "  	  	 ",
			want: []string{},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := cleanInput(c.text)
			if diff := cmp.Diff(c.want, got); diff != "" {
				t.Errorf("cleanInput(%q): diff (-want +got):\n%s", c.text, diff)
			}
		})
	}
}
