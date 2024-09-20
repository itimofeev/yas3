package front

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"

	"github.com/itimofeev/yas3/internal/entity"
)

type storeClient interface {
	UploadFile(ctx context.Context, fileName string, content io.Reader) error
	GetFile(ctx context.Context, fileName string) (io.ReadCloser, error)
	GetAvailableSpace(ctx context.Context) (entity.AvailableSpace, error)
}

type Config struct {
	Addr             string        `validate:"required"`
	ReadTimeout      time.Duration `validate:"required"`
	WriteTimeout     time.Duration `validate:"required"`
	MaxFileSizeBytes int64         `validate:"required,gt=0"`
	PartsCount       int64         `validate:"required,gt=0"`
	StoreClient      storeClient   `validate:"required"`
}

type Server struct {
	srv         *http.Server
	cfg         Config
	storeClient storeClient
}

func New(cfg Config) (*Server, error) {
	err := validator.New().Struct(cfg)
	if err != nil {
		return nil, fmt.Errorf("config validation error: %w", err)
	}

	frontServer := &Server{
		cfg:         cfg,
		storeClient: cfg.StoreClient,
	}

	handler := frontServer.initServerHandler()
	frontServer.srv = &http.Server{
		Addr:         cfg.Addr,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		Handler:      handler,
	}

	return frontServer, nil
}

func (s *Server) Run(ctx context.Context) error {
	closedCh := make(chan struct{})

	go func() {
		<-ctx.Done()
		slog.Info("web server graceful shutdown is in progress")

		withTimeout, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		//nolint:contextcheck // intentionally used another context as main one is most probably already canceled
		if err := s.srv.Shutdown(withTimeout); err != nil {
			slog.Warn("err stopping http server", "err", err)
		}

		slog.Info("web server gracefully stopped")
		close(closedCh)
	}()

	slog.Info("starting http server on", "addr", s.srv.Addr)
	if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	<-closedCh

	return nil
}
