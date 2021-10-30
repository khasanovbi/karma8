package server

import (
	"go.uber.org/zap"
	"karma8"
	"karma8/internal/fileservice"
	"karma8/internal/storageholder"
)

func newFileService(
	balancer karma8.Balancer,
	fileMetaStorage karma8.FileMetaStorage,
	minChunkSize int64,
	hostSplitCount int,
	logger *zap.Logger,
) karma8.FileService {
	storageHolder := storageholder.New()
	return fileservice.New(balancer, storageHolder, fileMetaStorage, minChunkSize, hostSplitCount, logger)
}
