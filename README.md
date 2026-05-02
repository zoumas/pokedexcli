# pokedexcli

This repository will follow the [boot.dev Build a Pokedex CLI in Golang](https://www.boot.dev/courses/build-pokedex-cli-golang) guided project.

## Learning goals

By working through the first lesson, the goal is to practice:

- setting up a local Go development environment and using the Go toolchain comfortably
- making HTTP requests in Go to fetch data from an external API
- decoding JSON responses into Go structs
- building a CLI tool that makes interacting with a back-end API easier
- understanding where caching can improve responsiveness as the project grows

## Idiomatic Go improvements to apply as you build

Use these as review criteria while working through the guided project:

- keep `main` small and move command, API, and state logic into focused packages or files
- prefer constructors and explicit dependencies over package-level mutable globals
- return `error` values with useful context instead of printing deep inside helpers
- use short, descriptive names that match Go conventions; avoid Java-style names and unnecessary prefixes
- define structs for API payloads close to the code that owns them
- pass `context.Context` to networked operations once requests become more than trivial
- keep command handlers narrow: parse input, call domain logic, format output
- use table-driven tests for parsing, command dispatch, and state transitions
- document exported identifiers only when they are part of a reusable package API
