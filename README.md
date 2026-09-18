# pokedexcli

[![Go Reference](https://pkg.go.dev/badge/github.com/zoumas/pokedexcli.svg)](https://pkg.go.dev/github.com/zoumas/pokedexcli)
[![Go Version](https://img.shields.io/badge/go-1.27-00ADD8?logo=go)](https://go.dev/dl/)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

An interactive command-line Pokedex. `pokedexcli` is a REPL that queries
[PokeAPI](https://pokeapi.co/), lets you walk the world's location areas, catch
Pokemon, and inspect what you've caught — with an expiring in-memory cache so
repeated lookups never hit the network twice.

```text
Pokedex > explore pastoria-city-area
Exploring pastoria-city-area...
Found Pokemon:
 - tentacool
 - magikarp
 - gyarados
Pokedex > catch magikarp
Throwing a Pokeball at magikarp...
magikarp was caught!
Pokedex > inspect magikarp
Name: magikarp
Height: 9
Weight: 100
Stats:
  -hp: 20
  -attack: 10
  -defense: 55
Types:
  - water
Pokedex > pokedex
Your Pokedex:
  129. magikarp
```

## Features

- REPL with a command registry, tolerant of casing and stray whitespace.
- Paginated browsing of PokeAPI location areas (`map` / `mapb`).
- Catch mechanics weighted by a Pokemon's `base_experience`.
- Concurrency-safe in-memory cache with TTL expiry and a background reaper.
- Structured JSON logging via `log/slog`, off by default.
- Graceful `SIGINT` handling: logs are flushed before exit (exit code `130`).
- No third-party runtime dependencies.

## Installation

```sh
go install github.com/zoumas/pokedexcli@latest
```

Or build from source:

```sh
git clone https://github.com/zoumas/pokedexcli
cd pokedexcli
go build -o pokedexcli .
./pokedexcli
```

Requires Go 1.27 or newer.

## Usage

Start the REPL and type commands at the `Pokedex > ` prompt. `exit`, `Ctrl-C`,
or `Ctrl-D` ends the session.

| Command | Arguments | Description |
| --- | --- | --- |
| `help` | — | List the available commands. |
| `map` | — | Show the next 20 location areas. |
| `mapb` | — | Show the previous 20 location areas. |
| `explore` | `<location-area>` | List the Pokemon encountered in a location area. |
| `catch` | `<pokemon>` | Attempt to catch a Pokemon by name or id. |
| `inspect` | `<pokemon>` | Show height, weight, stats, and types of a caught Pokemon. |
| `pokedex` | — | List every Pokemon you've caught, ordered by Pokedex number. |
| `exit` | — | Close the Pokedex. |

Caught Pokemon live in memory only; each run starts with an empty Pokedex.

## Configuration

| Variable | Default | Description |
| --- | --- | --- |
| `POKEDEXCLI_LOG_FILE` | unset | Path to a JSON log file. When unset, all log records are discarded. |

```sh
POKEDEXCLI_LOG_FILE=debug.log ./pokedexcli
```

Records are written at `DEBUG` level and include cache hits, misses, reaps, and
command failures. The file is opened in append mode and flushed on exit.

Other knobs are compile-time constants, deliberately: the HTTP request timeout
(`10s`, `main.go`), the cache TTL (`5s`, `main.go`), and the catch difficulty
threshold (`50`, `command.go`).

## Architecture

```text
main.go       process wiring: logger, signal handling, HTTP client, exit codes
repl.go       read-eval-print loop and input parsing
command.go    config (REPL state) and the command registry with its handlers

internal/pokeapi     typed PokeAPI client (location areas, pokemon)
internal/pokecache   TTL cache, concurrency-safe, with a background reaper
```

Notes on the design:

- **State lives in `config`.** Handlers share one `*config` carrying the output
  writer, logger, registry, pagination cursors, caught Pokemon, and the RNG.
  Everything a handler touches is injected, so every handler is testable
  without touching the network, the clock, or global state.
- **The cache sits under the client, not beside it.** `pokeapi.Client.get`
  checks the cache before each request and stores only successful responses,
  keyed by full URL. Callers get caching for free and cannot forget it.
- **The reaper is tied to a context.** `pokecache.New` takes a `context.Context`
  and its goroutine stops when that context is cancelled, so the cache does not
  outlive the process that owns it.
- **The REPL blocks on stdin**, so it cannot select on a context. `main` runs it
  in a goroutine and races it against `signal.NotifyContext`, which is what makes
  an interrupt still flush the log file.
- **The base URL is a field**, not a constant reference, which is what lets the
  tests run the whole client against an `httptest.Server`.

## Development

```sh
go test ./...          # unit tests
go test -race ./...    # with the race detector
go vet ./...
gofmt -l .
```

Tests are table-driven with subtests, use `httptest.Server` for the API layer,
and `github.com/google/go-cmp` for diffs. `go-cmp` is the only dependency and it
is test-only.

## What this project is

`pokedexcli` was built as the guided project
[Build a Pokedex in Go](https://www.boot.dev/courses/build-pokedex-cli-golang)
on [boot.dev](https://www.boot.dev/), then taken past the course requirements:
structured logging, signal handling, context-scoped background work, injected
randomness, and an API client that is testable end to end.

[![Boot.dev Build a Pokedex in Go certificate](https://qvault-webapp-dynamic-assets.storage.googleapis.com/certificates/094f2f10-844e-4f90-9c7b-a11c7c2e3a70.jpeg?v=1789764895)](https://www.boot.dev/certificates/094f2f10-844e-4f90-9c7b-a11c7c2e3a70)

### What the project covers

**Building a REPL.** Reading stdin line by line, normalizing input into a
command plus arguments, and dispatching through a `map[string]cliCommand`
registry rather than a `switch`. Unknown commands and failing commands report
and continue; only `exit` and exhausted input stop the loop.

**HTTP and JSON.** Calling a third-party API with `net/http`, modelling only the
response fields the program actually uses, decoding with `encoding/json`, and
following `next`/`previous` pagination links. Errors are wrapped with `%w` and
context at each layer, so a failure reads as a chain rather than a bare string.

**Caching and concurrency.** A TTL cache guarded by a `sync.RWMutex`, with a
`time.Ticker` goroutine that reaps stale entries. This is where goroutines,
mutexes, and contexts stop being syntax exercises: the reaper needs a shutdown
signal, the log writer needs serializing because two goroutines write to it, and
the race detector tells you when you got it wrong.

**Designing for tests.** Accepting `io.Reader`/`io.Writer` instead of touching
`os.Stdin`/`os.Stdout`, passing a `*rand.Rand` instead of calling the package
functions, and making the API base URL a field so a test server can stand in for
PokeAPI. None of it is test-only machinery — it is just dependency injection,
and it happens to make the tests trivial.

**Project layout.** Splitting into `internal/` packages with clear
responsibilities, documenting every exported item, and keeping ownership rules
explicit — `Cache.Get` returns the stored slice rather than a copy, so the doc
comment says so.

## License

[MIT](LICENSE).

## Credits

Pokemon data from [PokeAPI](https://pokeapi.co/). Pokemon and Pokedex are
trademarks of Nintendo; this is an unofficial, non-commercial learning project.
