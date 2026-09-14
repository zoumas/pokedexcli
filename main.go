package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/zoumas/pokedexcli/internal/pokeapi"
	"github.com/zoumas/pokedexcli/internal/pokecache"
)

// requestTimeout bounds every PokeAPI request, so a stalled connection cannot
// hang the REPL.
const requestTimeout = 10 * time.Second

func main() {
	os.Exit(run())
}

func run() int {
	logger, closeLog, err := initializeLogger(os.Getenv("POKEDEXCLI_LOG_FILE"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize logger: %s\n", err)
		return 1
	}
	defer func() {
		if err := closeLog(); err != nil {
			fmt.Fprintf(os.Stderr, "failed to close logger: %s\n", err)
		}
	}()

	const cacheInterval = 5 * time.Second

	// Cancelling ctx stops the cache reaper and, on an interrupt, unblocks run
	// so that the deferred log flush still happens.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	httpClient := &http.Client{
		Timeout: requestTimeout,
	}
	cache := pokecache.New(ctx, logger, cacheInterval)
	client := pokeapi.New(httpClient, cache)
	cfg := newConfig(os.Stdout, client)

	// startREPL blocks reading os.Stdin, so it cannot observe ctx itself. Run it
	// alongside the interrupt and take whichever finishes first.
	done := make(chan int, 1)
	go func() {
		done <- startREPL(os.Stdin, cfg)
	}()

	select {
	case exitCode := <-done:
		return exitCode
	case <-ctx.Done():
		fmt.Fprintln(os.Stdout)
		return exitCodeInterrupted
	}
}

// exitCodeInterrupted is the conventional shell exit code for a process
// terminated by SIGINT: 128 plus the signal number.
const exitCodeInterrupted = 130

type closeFunc func() error

func initializeLogger(name string) (*slog.Logger, closeFunc, error) {
	if name == "" {
		return slog.New(slog.NewTextHandler(io.Discard, nil)), func() error { return nil }, nil
	}
	file, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, func() error { return nil }, err
	}

	bufferedFile := bufio.NewWriterSize(file, 8192)

	handler := slog.NewJSONHandler(bufferedFile, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := slog.New(handler)

	return logger, func() error {
		return errors.Join(bufferedFile.Flush(), file.Close())
	}, nil
}
