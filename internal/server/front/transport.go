package front

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/pprof"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
)

func (s *Server) uploadFileHandler(resp http.ResponseWriter, req *http.Request) {
	fileIDStr := chi.URLParam(req, "fileID")
	fileID, err := uuid.Parse(fileIDStr)
	if err != nil {
		s.error(req, resp, err)
		return
	}

	// fileSize parameter is used to be able to split files into parts without reading the whole content into memory
	// current code skips checks that real file size matches number in parameter, but it better to add this check in production implementation.
	fileSizeStr := req.URL.Query().Get("fileSize")
	fileSize, err := strconv.ParseInt(fileSizeStr, 10, 64)
	if err != nil {
		s.error(req, resp, err)
		return
	}
	if fileSize > s.cfg.MaxFileSizeBytes {
		s.error(req, resp, fmt.Errorf("too big file size, max size %d", s.cfg.MaxFileSizeBytes))
		return
	}

	// check if file already uploaded with this id
	if s.fileRegistry.IsFileExists(fileID.String()) {
		s.error(req, resp, fmt.Errorf("file already exists %s", fileID.String()))
		return
	}

	partSize := fileSize/s.cfg.PartsCount + 1

	// receives link to store servers where we can upload file parts
	storeServers, err := s.serversRegistry.GetServersForParts(s.cfg.PartsCount)
	if err != nil {
		s.error(req, resp, err)
		return
	}

	// upload file parts in cycle without any parallelism, because network to our store servers will be much faster than from client to rest server
	serverIDs := make([]string, 0, len(storeServers))
	for partNumber, storeForUpload := range storeServers {
		fileName := fileID.String() + "." + strconv.FormatInt(int64(partNumber), 10)
		partReader := io.LimitReader(req.Body, partSize)
		err := storeForUpload.UploadFile(req.Context(), fileName, partReader)

		if err != nil {
			s.error(req, resp, err)
			return
		}
		serverIDs = append(serverIDs, storeForUpload.GetID())
	}

	// save information about uploaded file parts to internal db,
	// current implementation is not very reliable, can discuss how to make it better
	if err := s.fileRegistry.SaveFileParts(fileID.String(), serverIDs); err != nil {
		s.error(req, resp, err)
		return
	}
}

func (s *Server) getFileHandler(resp http.ResponseWriter, req *http.Request) {
	fileIDStr := chi.URLParam(req, "fileID")
	fileID, err := uuid.Parse(fileIDStr)
	if err != nil {
		s.error(req, resp, err)
		return
	}

	serverIDs, err := s.fileRegistry.GetFileParts(fileID.String())
	if err != nil {
		s.error(req, resp, err)
		return
	}

	// get information about store servers on which file parts are stored
	storeClients, err := s.serversRegistry.GetStoreClients(serverIDs)
	if err != nil {
		s.error(req, resp, err)
		return
	}

	for partNumber, storeClient := range storeClients {
		fileName := fileID.String() + "." + strconv.FormatInt(int64(partNumber), 10)

		// copy part content from store server response directly to rest server response
		err := func(ctx context.Context) error {
			filePartReader, err := storeClient.GetFile(ctx, fileName)
			if err != nil {
				return err
			}
			defer filePartReader.Close()

			_, err = io.Copy(resp, filePartReader)
			return err
		}(req.Context())

		if err != nil {
			s.error(req, resp, err)
			return
		}
	}
}

func (s *Server) initServerHandler() *chi.Mux {
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
			api.Post("/uploadFile/{fileID}", s.uploadFileHandler)
			api.Get("/getFile/{fileID}", s.getFileHandler)
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
