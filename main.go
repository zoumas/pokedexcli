package main

import "os"

func main() {
	cfg := newConfig(os.Stdout)
	os.Exit(startREPL(os.Stdin, cfg))
}
