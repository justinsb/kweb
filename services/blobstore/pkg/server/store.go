package server

import (
	"context"
	"io"
)

type Store interface {
	// CreateBlob will create a blob from the content, verifying the sha256 hash matches.
	CreateBlob(ctx context.Context, sha256 []byte, content io.Reader) (*BlobInfo, error)
	// OpenBlob returns a reader for the blob with specified sha256 hash.
	Open(ctx context.Context, sha256 []byte) (io.ReadCloser, error)
}

type BlobInfo struct {
	Length int64
}
