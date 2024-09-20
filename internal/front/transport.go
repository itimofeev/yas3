package front

import (
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
