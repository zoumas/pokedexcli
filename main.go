package main

import (
	"log"
	"os"
)

func main() {
	if err := repl(os.Stdin, os.Stdout); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
