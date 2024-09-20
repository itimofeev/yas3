package front

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Config struct {
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type Server struct {
	srv *http.Server
}

func New(cfg Config) (*Server, error) {
	frontServer := &Server{}

	handler := frontServer.initServerHandler()
	frontServer.srv = &http.Server{
		Addr:         cfg.Addr,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		Handler:      handler,
	}

	return frontServer, nil
}

func (sf *Server) initServerHandler() *chi.Mux {
	r := chi.NewRouter()
	r.Use(
		middleware.Recoverer,
	)

	r.Group(func(router chi.Router) {
		router.Group(func(telemetry chi.Router) {
			telemetry.HandleFunc("/pprof", func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, r.RequestURI+"/", http.StatusMovedPermanently)
			})
			telemetry.HandleFunc("/pprof/*", pprof.Index)
			telemetry.HandleFunc("/pprof/cmdline", pprof.Cmdline)
			telemetry.HandleFunc("/pprof/profile", pprof.Profile)
			telemetry.HandleFunc("/pprof/symbol", pprof.Symbol)
			telemetry.HandleFunc("/pprof/trace", pprof.Trace)

			telemetry.Handle("/pprof/goroutine", pprof.Handler("goroutine"))
			telemetry.Handle("/pprof/threadcreate", pprof.Handler("threadcreate"))
			telemetry.Handle("/pprof/mutex", pprof.Handler("mutex"))
			telemetry.Handle("/pprof/heap", pprof.Handler("heap"))
			telemetry.Handle("/pprof/block", pprof.Handler("block"))
			telemetry.Handle("/pprof/allocs", pprof.Handler("allocs"))
		})
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.RequestID)
		r.Route("/api/v1", func(api chi.Router) {
			api.Post("/uploadFile/{fileID}", sf.uploadFileHandler)
			api.Post("/getFile/{fileID}", sf.getFileHandler)
		})
	})

	return r
}

func (sf *Server) Run(ctx context.Context) error {
	closedCh := make(chan struct{})

	go func() {
		<-ctx.Done()
		slog.Info("web server graceful shutdown is in progress")

		withTimeout, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		//nolint:contextcheck // intentionally used another context as main one is most probably already canceled
		if err := sf.srv.Shutdown(withTimeout); err != nil {
			slog.Warn("err stopping http server", "err", err)
		}

		slog.Info("web server gracefully stopped")
		close(closedCh)
	}()

	slog.Info("starting http server on", "addr", sf.srv.Addr)
	if err := sf.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	<-closedCh

	return nil
}
