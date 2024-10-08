package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/kelseyhightower/envconfig"
	"golang.org/x/sync/errgroup"

	"github.com/itimofeev/yas3/internal/server/store"
)

// STORE_SERVER_ADDR=:9090;STORE_BASE_PATH=temp/store/1
type configuration struct {
	StoreServerAddr     string `envconfig:"STORE_SERVER_ADDR" default:":9090"`
	StoreBasePath       string `envconfig:"STORE_BASE_PATH" default:"temp/store/1"`
	StoreTotalSizeBytes int    `envconfig:"STORE_TOTAL_SIZE_BYTES" default:"1073741824"` // 1Gb
}

func main() {
	cfg := mustParseConfig()

	slog.Info("store-server is starting")
	if err := run(cfg); err != nil && !errors.Is(err, context.Canceled) {
		slog.Error("store-server is stopped with error", "err", err)
	}

	slog.Info("store-server is stopped")
}

func run(cfg configuration) error {
	ctx := signalContext()

	storeServer, err := store.New(store.Config{
		Addr:                   cfg.StoreServerAddr,
		BasePath:               cfg.StoreBasePath,
		MaxAvailableSpaceBytes: cfg.StoreTotalSizeBytes,
	})
	if err != nil {
		return err
	}

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		return storeServer.Run(ctx)
	})

	return eg.Wait()
}

// signalContext returns a context that is canceled if either SIGTERM or SIGINT signal is received.
func signalContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		sig := <-c
		slog.Info("received signal", "sig", sig)
		cancel()
	}()

	return ctx
}

func mustParseConfig() configuration {
	var cfg configuration
	if err := envconfig.Process("", &cfg); err != nil {
		panic(fmt.Sprintf("failed to load configuration: %v", err))
	}
	return cfg
}
