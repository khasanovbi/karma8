package api

import (
	"errors"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"io"
	"karma8"
	"net/http"
	"strconv"
)

const (
	headerContentLength = "Content-Length"
)

var errUnknownContentLength = errors.New("unknown Content-Length")

func writePlainErr(w http.ResponseWriter, err error, status int, logger *zap.Logger) {
	w.WriteHeader(status)
	_, writeErr := w.Write([]byte(err.Error()))
	if writeErr != nil {
		logger.Error("can't write response", zap.Error(writeErr))
	}
}

func writePlainInternalErr(w http.ResponseWriter, err error, logger *zap.Logger) {
	writePlainErr(w, err, http.StatusInternalServerError, logger)
}

func NewPutFileHandler(service karma8.FileService, logger *zap.Logger) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		filename := mux.Vars(request)["filename"]

		if request.ContentLength == -1 {
			writePlainErr(writer, errUnknownContentLength, http.StatusBadRequest, logger)
			return
		}

		file := &karma8.File{
			Meta: &karma8.FileMeta{
				Name:          filename,
				ContentLength: request.ContentLength,
			},
			Body: request.Body,
		}

		if err := service.PutFile(request.Context(), file); err != nil {
			logger.Error("can't put file", zap.Error(err))
			writePlainInternalErr(writer, err, logger)
			return
		}
	}
}

func NewGetFileHandler(service karma8.FileService, logger *zap.Logger) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		filename := mux.Vars(request)["filename"]

		file, err := service.GetFile(request.Context(), filename)
		if err != nil {
			logger.Error("can't get file", zap.Error(err))
			writePlainInternalErr(writer, err, logger)
			return
		}

		defer file.Body.Close()

		writer.Header().Set(headerContentLength, strconv.FormatInt(file.Meta.ContentLength, 10))

		if _, err = io.Copy(writer, file.Body); err != nil {
			logger.Error("can't write file body", zap.Error(err))
			writePlainInternalErr(writer, err, logger)
			return
		}
	}
}
