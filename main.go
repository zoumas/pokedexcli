package main

import "os"

func main() {
	cfg := &config{
		w:        os.Stdout,
		registry: newCommandRegistry(),
	}
	os.Exit(startREPL(os.Stdin, cfg))
}
