package server

import (
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"karma8"
	"karma8/internal/filemetastorage"
)

func newFileMetaStorage(pg *sqlx.DB, logger *zap.Logger) karma8.FileMetaStorage {
	return filemetastorage.NewPGStorage(pg, logger)
}
