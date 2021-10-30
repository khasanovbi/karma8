package karma8

import (
	"context"
	"io"
)

// StorageHolder required to keep clients to storage. It would be useful to preserve keep-alive requests to storage.
type StorageHolder interface {
	GetStorage(host string) Storage
}

type Storage interface {
	UploadFilePart(ctx context.Context, path string, body io.Reader) error
	ReadFilePart(ctx context.Context, path string) (io.ReadCloser, error)
	DeleteFilePart(ctx context.Context, path string) error
}

type FilePart struct {
	StorageURL    string
	Path          string
	ContentLength int64
}

type FileMetaStorage interface {
	// PutProcessingFileMeta saves data before upload, required to clean up storage in case of failures during upload.
	PutProcessingFileMeta(ctx context.Context, meta *FileMeta) error
	CompleteFileMeta(ctx context.Context, filename string) error
	GetFileMeta(ctx context.Context, filename string) (*FileMeta, error)
}

type Balancer interface {
	GetHosts(ctx context.Context, count int) ([]string, error)
}

type FileMeta struct {
	Name          string
	Parts         []*FilePart
	ContentLength int64
}

type File struct {
	Meta *FileMeta
	Body io.ReadCloser
}

type FileService interface {
	PutFile(ctx context.Context, file *File) error
	GetFile(ctx context.Context, filename string) (*File, error)
}
