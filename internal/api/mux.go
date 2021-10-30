package api

import (
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"karma8"
	"net/http"
)

func NewMux(fileService karma8.FileService, logger *zap.Logger) http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/file/{filename}", NewPutFileHandler(fileService, logger)).Methods(http.MethodPut)
	r.HandleFunc("/file/{filename}", NewGetFileHandler(fileService, logger)).Methods(http.MethodGet)
	return r
}
