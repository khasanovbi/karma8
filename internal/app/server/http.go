package server

import (
	"go.uber.org/zap"
	"karma8"
	"karma8/internal/api"
	"net/http"
)

func newHTTPServer(conf *HTTPConfig, fileService karma8.FileService, logger *zap.Logger) *http.Server {
	mux := api.NewMux(fileService, logger)
	return &http.Server{
		Addr:    conf.Addr,
		Handler: mux,
	}
}
