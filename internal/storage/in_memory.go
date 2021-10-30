package storage

import (
	"bytes"
	"context"
	"io"
	"karma8"
	"sync"
)

type inMemory struct {
	pathToData map[string][]byte
	mutex      sync.RWMutex
}

func (m *inMemory) UploadFilePart(ctx context.Context, path string, body io.Reader) error {
	data, err := io.ReadAll(body)
	if err != nil {
		return err
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.pathToData[path] = data

	return nil
}

func (m *inMemory) ReadFilePart(ctx context.Context, path string) (io.ReadCloser, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return io.NopCloser(bytes.NewReader(m.pathToData[path])), nil
}

func (m *inMemory) DeleteFilePart(ctx context.Context, path string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	delete(m.pathToData, path)

	return nil
}

func NewInMemory() karma8.Storage {
	return &inMemory{
		pathToData: map[string][]byte{},
	}
}
