package front

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (sf *Front) uploadFileHandler(resp http.ResponseWriter, req *http.Request) {
	fileID := chi.URLParam(req, "fileID")
	fileSize := req.URL.Query().Get("fileSize")
	resp.Write([]byte("fileID: " + fileID + ", fileSize: " + fileSize))
}

func (sf *Front) getFileHandler(resp http.ResponseWriter, req *http.Request) {
	fileID := chi.URLParam(req, "fileID")
	fileSize := req.URL.Query().Get("fileSize")
	resp.Write([]byte("fileID: " + fileID + ", fileSize: " + fileSize))
}

func (sf *Front) error(_ *http.Request, w http.ResponseWriter, err error) {
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
