package fileservice

import (
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"io"
	"karma8"
)

type fileService struct {
	balancer        karma8.Balancer
	storageHolder   karma8.StorageHolder
	fileMetaStorage karma8.FileMetaStorage
	minChunkSize    int64
	hostSplitCount  int

	logger *zap.Logger
}

func minInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func (m *fileService) calculatePartsSize(total int64, splitCount int) []int64 {
	result := make([]int64, 0, splitCount)

	remain := total
	for remain > 0 {
		partSize := remain / int64(splitCount)
		if remain%int64(splitCount) != 0 {
			partSize += 1
		}

		if partSize <= m.minChunkSize {
			partSize = minInt64(remain, m.minChunkSize)
		}

		result = append(result, partSize)

		remain -= partSize
		splitCount -= 1
	}

	return result
}

func (m *fileService) calculateFileParts(
	hosts []string,
	fileMeta *karma8.FileMeta,
	partSizes []int64,
) []*karma8.FilePart {
	var fileParts []*karma8.FilePart
	for i, partSize := range partSizes {
		fileParts = append(fileParts, &karma8.FilePart{
			StorageURL:    hosts[i],
			Path:          fileMeta.Name,
			ContentLength: partSize,
		})
	}
	return fileParts
}

func (m *fileService) PutFile(ctx context.Context, file *karma8.File) error {
	partSizes := m.calculatePartsSize(file.Meta.ContentLength, m.hostSplitCount)

	hosts, err := m.balancer.GetHosts(ctx, len(partSizes))
	if err != nil {
		m.logger.Error("can't get hosts from balancer", zap.Error(err))
		return fmt.Errorf("get hosts error: %w", err)
	}

	fileParts := m.calculateFileParts(hosts, file.Meta, partSizes)
	file.Meta.Parts = fileParts

	if err := m.fileMetaStorage.PutProcessingFileMeta(ctx, file.Meta); err != nil {
		m.logger.Error("can't put processing file meta", zap.Error(err))
		return fmt.Errorf("can't put processing file meta: %w", err)
	}

	for _, filePart := range fileParts {
		body := io.LimitReader(file.Body, filePart.ContentLength)
		storage := m.storageHolder.GetStorage(filePart.StorageURL)
		if err := storage.UploadFilePart(ctx, filePart.Path, body); err != nil {
			m.logger.Error("can't upload file part", zap.Error(err))
			return fmt.Errorf("can't upload file part: %w", err)
		}
	}

	if err := m.fileMetaStorage.CompleteFileMeta(ctx, file.Meta.Name); err != nil {
		m.logger.Error("can't complete file meta", zap.Error(err))
		return fmt.Errorf("can't complete file meta: %w", err)
	}

	return nil
}

type multiStorageReader struct {
	fileMeta         *karma8.FileMeta
	storageHolder    karma8.StorageHolder
	currentPartIndex int
	currentBody      io.ReadCloser
	ctx              context.Context
}

func (m *multiStorageReader) moveToNextStorage() error {
	if m.currentPartIndex >= len(m.fileMeta.Parts) {
		return io.EOF
	}

	part := m.fileMeta.Parts[m.currentPartIndex]
	storage := m.storageHolder.GetStorage(part.StorageURL)

	body, err := storage.ReadFilePart(m.ctx, part.Path)
	if err != nil {
		m.currentBody = nil
		return err
	}

	m.currentBody = body

	m.currentPartIndex++

	return nil
}

func (m *multiStorageReader) Read(p []byte) (int, error) {
	if m.currentBody == nil {
		if err := m.moveToNextStorage(); err != nil {
			return 0, err
		}
	}

	n, err := m.currentBody.Read(p)
	if err != nil {
		if !errors.Is(err, io.EOF) {
			return n, err
		}

		if err := m.currentBody.Close(); err != nil {
			return n, err
		}

		if err := m.moveToNextStorage(); err != nil {
			return n, err
		}

		return n, nil
	}

	return n, nil
}

func (m *multiStorageReader) Close() error {
	if m.currentBody != nil {
		return m.currentBody.Close()
	}
	return nil
}

func (m *fileService) GetFile(ctx context.Context, filename string) (*karma8.File, error) {
	m.logger.Info("start get file request", zap.String("filename", filename))

	fileMeta, err := m.fileMetaStorage.GetFileMeta(ctx, filename)
	if err != nil {
		m.logger.Error("can't get file meta", zap.Error(err))
		return nil, fmt.Errorf("can't get file meta: %w", err)
	}

	return &karma8.File{
		Meta: fileMeta,
		Body: &multiStorageReader{
			fileMeta:      fileMeta,
			storageHolder: m.storageHolder,
			ctx:           ctx,
		},
	}, nil
}

func New(
	balancer karma8.Balancer,
	storageHolder karma8.StorageHolder,
	fileMetaStorage karma8.FileMetaStorage,
	minChunkSize int64,
	hostSplitCount int,
	logger *zap.Logger,
) karma8.FileService {
	return &fileService{
		balancer:        balancer,
		storageHolder:   storageHolder,
		fileMetaStorage: fileMetaStorage,
		minChunkSize:    minChunkSize,
		hostSplitCount:  hostSplitCount,
		logger:          logger,
	}
}
