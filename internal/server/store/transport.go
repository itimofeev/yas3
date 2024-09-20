package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/pprof"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (s *Server) uploadFile(resp http.ResponseWriter, req *http.Request) {
	fileName := chi.URLParam(req, "fileName")
	file, err := os.OpenFile(s.cfg.BasePath+"/"+fileName, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		s.error(req, resp, err)
		return
	}
	defer file.Close()

	_, err = io.Copy(file, req.Body)
	if err != nil {
		s.error(req, resp, err)
		return
	}
	if err := file.Sync(); err != nil {
		s.error(req, resp, err)
		return
	}

	_, _ = resp.Write([]byte("ok"))
}

func (s *Server) getFile(resp http.ResponseWriter, req *http.Request) {
	fileName := chi.URLParam(req, "fileName")
	file, err := os.OpenFile(s.cfg.BasePath+"/"+fileName, os.O_RDONLY, 0o600)
	if err != nil {
		s.error(req, resp, err)
		return
	}
	defer file.Close()

	_, err = io.Copy(resp, file)
	if err != nil {
		slog.Warn("error while sending file", "err", err)
	}
}

func (s *Server) getAvailableSpace(resp http.ResponseWriter, req *http.Request) {
	var size int64
	err := filepath.Walk(s.cfg.BasePath, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})

	if err != nil {
		slog.Warn("error while sending file", "err", err)
	}

	_, _ = resp.Write([]byte(fmt.Sprintf(`{"total": %d, "used": %d}`, s.cfg.MaxAvailableSpaceBytes, size)))
}

func (s *Server) initRouter() http.Handler {
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
		r.Use(middleware.Logger)
		r.Route("/api/v1", func(api chi.Router) {
			api.Post("/uploadFile/{fileName}", s.uploadFile)
			api.Get("/getFile/{fileName}", s.getFile)
			api.Get("/getAvailableSpace", s.getAvailableSpace)
		})
	})

	return r
}

func (s *Server) error(_ *http.Request, w http.ResponseWriter, err error) {
	slog.Warn("got error while handling request", "err", err)
	switch {
	case errors.Is(err, context.Canceled):
		writeErrResponse(w, "timeout", http.StatusRequestTimeout)
	default:
		writeErrResponse(w, err.Error(), http.StatusInternalServerError)
	}
}

func writeErrResponse(w http.ResponseWriter, err string, status int) {
	w.Header().Set("Content-type", "application/json")

	w.WriteHeader(status)

	m := map[string]interface{}{
		"error": err,
		"code":  status,
	}
	_ = json.NewEncoder(w).Encode(m)
}
