package main

import "os"

func main() {
	os.Exit(startREPL(os.Stdin, os.Stdout))
}
