package main

import "strings"

func cleanInput(s string) []string {
	return strings.Fields(strings.ToLower(s))
}
