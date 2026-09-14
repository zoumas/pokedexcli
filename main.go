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
	"sync"
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
	client := pokeapi.New(httpClient, cache, pokeapi.DefaultBaseURL)
	cfg := newConfig(os.Stdout, client, logger)

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
		_, _ = fmt.Fprintln(os.Stdout)
		return exitCodeInterrupted
	}
}

// exitCodeInterrupted is the conventional shell exit code for a process
// terminated by SIGINT: 128 plus the signal number.
const exitCodeInterrupted = 130

type closeFunc func() error

// syncWriter serializes writes and the final flush and close. The REPL runs in
// its own goroutine and can still be logging when an interrupt makes run return
// and flush, and a bufio.Writer is not safe for concurrent use.
type syncWriter struct {
	mu sync.Mutex
	bw *bufio.Writer
	f  *os.File
}

func (s *syncWriter) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.bw.Write(p)
}

// Close flushes any buffered records and closes the underlying file.
func (s *syncWriter) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return errors.Join(s.bw.Flush(), s.f.Close())
}

// initializeLogger returns a logger writing JSON records to the file named by
// name, along with a function that flushes and closes it. An empty name
// discards all records.
func initializeLogger(name string) (*slog.Logger, closeFunc, error) {
	noop := func() error { return nil }

	if name == "" {
		return slog.New(slog.NewTextHandler(io.Discard, nil)), noop, nil
	}
	file, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, noop, err
	}

	w := &syncWriter{bw: bufio.NewWriter(file), f: file}

	handler := slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	return slog.New(handler), w.Close, nil
}
