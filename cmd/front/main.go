package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kelseyhightower/envconfig"
	"golang.org/x/sync/errgroup"

	fileregistry "github.com/itimofeev/yas3/internal/provider/file-registry"
	serverRegistry "github.com/itimofeev/yas3/internal/provider/server-registry"
	"github.com/itimofeev/yas3/internal/server/front"
)

// FRONT_STORE_CLIENT_ADDR=https://localhost:9090,https://localhost:9091
type configuration struct {
	FrontAddr         string        `envconfig:"FRONT_ADDR" default:":8080"`
	FrontReadTimeout  time.Duration `envconfig:"FRONT_READ_DURATION" default:"10s"`
	FrontWriteTimeout time.Duration `envconfig:"FRONT_WRITE_DURATION" default:"10s"`
	StoreServerAddrs  []string      `envconfig:"FRONT_STORE_CLIENT_ADDR" default:"https://localhost:9090"`
	FilesDBPath       string        `envconfig:"FRONT_FILES_DB_PATH" default:"temp/store/badger"`
}

func main() {
	cfg := mustParseConfig()

	slog.Info("front-server is starting")
	if err := run(cfg); err != nil && !errors.Is(err, context.Canceled) {
		slog.Error("service is stopped with error", "err", err)
	}

	slog.Info("service is stopped")
}

func run(cfg configuration) error {
	ctx := signalContext()

	storeServersRegistry, err := serverRegistry.New(ctx, serverRegistry.Config{StoreServerAddrs: cfg.StoreServerAddrs})
	if err != nil {
		return err
	}

	fileRegistry, err := fileregistry.New(fileregistry.Config{
		DBPath: cfg.FilesDBPath,
	})
	if err != nil {
		return err
	}
	defer fileRegistry.Close()

	frontServer, err := front.New(front.Config{
		Addr:             cfg.FrontAddr,
		ReadTimeout:      cfg.FrontReadTimeout,
		WriteTimeout:     cfg.FrontWriteTimeout,
		MaxFileSizeBytes: 1024 * 1024,
		PartsCount:       2,
		ServersRegistry:  storeServersRegistry,
		FileRegistry:     fileRegistry,
	})
	if err != nil {
		return err
	}

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		return frontServer.Run(ctx)
	})
	eg.Go(func() error {
		return storeServersRegistry.Run(ctx)
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
