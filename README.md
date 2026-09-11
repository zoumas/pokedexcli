# pokedexcli

An interactive command-line Pokedex written in Go, built as part of boot.dev's guided
project [Build a Pokedex in Go](https://www.boot.dev/courses/build-pokedex-cli-golang).

The program is a REPL (Read, Eval, Print Loop) that reads commands from the user,
queries [PokeAPI](https://pokeapi.co/), and prints the results.

## Learning goals

### 1. Build a REPL

- Read user input from stdin and parse it into a command plus arguments.
- Dispatch commands through a registry of command name -> handler.
- Handle unknown input, whitespace, and casing gracefully.
- Implement the baseline commands: `help` and `exit`.

### 2. HTTP networking and data serialization

- Call a third-party JSON API (PokeAPI) with Go's `net/http` package.
- Model JSON responses as Go structs and deserialize them with `encoding/json`.
- Work with paginated endpoints (following `next` / `previous` links).
- Handle network and decoding errors without crashing the REPL.

### 3. A caching layer

- Implement an in-memory cache to avoid repeated network requests.
- Expire stale entries with a background reaping loop.
- Make the cache safe for concurrent use with mutexes and goroutines.
- Learn how time-based invalidation trades freshness for speed.

### 4. Putting it together

- Wire the REPL, HTTP client, and cache into a single working application.
- Implement the Pokedex commands: exploring locations, catching Pokemon,
  inspecting them, and listing what has been caught.
- Structure a Go project into packages with clear responsibilities.
- Write unit tests for the parsing and caching logic.

## Usage

```sh
go build -o pokedexcli && ./pokedexcli
```
