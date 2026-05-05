# pokedexcli

`pokedexcli` is an interactive command-line Pokedex built in Go. It uses the [PokeAPI](https://pokeapi.co/) to browse location areas, discover which Pokemon can appear there, attempt catches, and inspect the Pokemon you have added to your in-memory Pokedex.

The project started from the boot.dev guided course, but the repository now documents the finished application rather than the tutorial steps.

## Features

- interactive REPL with a simple command-driven workflow
- paginated browsing of PokeAPI location areas
- location exploration to list available Pokemon encounters
- catch attempts with success chance based on a Pokemon's base experience
- in-memory Pokedex for inspecting caught Pokemon during the current session
- generic in-memory caching layer to reduce repeated API calls
- table-driven tests covering command behavior, REPL flow, API integration helpers, and cache expiration

## Requirements

- Go 1.26.2 or a compatible Go toolchain
- network access to `https://pokeapi.co/`

## Running the CLI

From the repository root:

```bash
go run .
```

To build a local binary:

```bash
go build -o pokedexcli .
./pokedexcli
```

## Command reference

| Command | Arguments | Description |
| --- | --- | --- |
| `help` | none | Print the available commands. |
| `map` | none | Fetch the next page of location areas. |
| `mapb` | none | Fetch the previous page of location areas. |
| `explore` | `<location-area>` | List the Pokemon that can be encountered in a location area. |
| `catch` | `<pokemon>` | Attempt to catch a Pokemon and add it to your Pokedex if successful. |
| `inspect` | `<pokemon>` | Show height, weight, stats, and types for a caught Pokemon. |
| `pokedex` | none | List the Pokemon currently stored in your Pokedex. |
| `exit` | none | Exit the CLI. |

## Example session

```text
Pokedex > help
Welcome to the Pokedex!
Usage:

catch: Attempts to catch a Pokemon by name
exit: Exit the Pokedex
explore: Explores a location area and lists the Pokemon that can be found there
help: Displays a help message
inspect: Desplays detailed information about a Pokemon in your Pokedex by name
map: Displays a list of location areas
mapb: Displays a list of location areas in the previous page
pokedex: Lists all the Pokemon you have caught in your Pokedex
Pokedex > map
canalave-city-area
eterna-city-area
pastoria-city-area
...
Pokedex > explore canalave-city-area
Exploring canalave-city-area...
Found Pokemon:
 - tentacool
 - tentacruel
 ...
Pokedex > catch tentacool
Throwing a Pokeball at tentacool...
tentacool was caught!
You may now inspect it with the inspect command.
Pokedex > inspect tentacool
Name: tentacool
Height: 9
Weight: 455
Stats:
 - hp: 40
 - attack: 40
 ...
Types:
 - water
 - poison
```

## How the application works

`pokedexcli` keeps a single session configuration in memory for the lifetime of the process. That configuration owns:

- one reusable `http.Client`
- pagination state for location browsing
- separate caches for location pages, explored areas, and Pokemon lookups
- the current in-memory Pokedex

Catch outcomes are probabilistic. Pokemon with lower base experience are easier to catch, while stronger Pokemon become progressively harder, with a minimum catch chance floor.

The Pokedex is not persisted to disk, so restarting the CLI starts a fresh session.

## Project structure

| Path | Responsibility |
| --- | --- |
| `main.go` | Application entrypoint. |
| `repl.go` | Read-eval-print loop and command dispatch. |
| `command.go` | Command registration and command handlers. |
| `internal/pokeapi` | Typed PokeAPI client helpers and response models. |
| `internal/pokecache` | Generic expiring in-memory cache. |
| `*_test.go` | Unit and integration-style tests for CLI behavior and helpers. |

## Testing

Run the full test suite from the repository root:

```bash
go test ./...
```
