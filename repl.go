package main

import "strings"

// cleanInput splits the user's input into "words" based on whitespace.
// It also lowercases the input and trims any leading or trailing whitespace.
func cleanInput(text string) []string {
	return strings.Fields(strings.ToLower(text))
}
