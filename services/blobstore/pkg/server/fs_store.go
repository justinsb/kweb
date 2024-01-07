package server

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"

	"k8s.io/klog/v2"
)

type FilesystemStore struct {
	path string
}

func NewFilesystemStore(path string) *FilesystemStore {
	return &FilesystemStore{path: path}
}

type BlobInfo struct {
	Length int64
}

func (s *FilesystemStore) pathOnDisk(hash []byte) string {
	hashString := hex.EncodeToString(hash)
	p := filepath.Join(s.path, hashString[0:2], hashString[2:4], hashString)

	return p
}

func (s *FilesystemStore) Open(ctx context.Context, hash []byte) (io.ReadCloser, error) {
	p := s.pathOnDisk(hash)

	f, err := os.Open(p)
	if err != nil {
		return nil, fmt.Errorf("opening file %q: %w", p, err)
	}
	return f, nil
}

func (s *FilesystemStore) CreateBlob(ctx context.Context, hash []byte, r io.Reader) (*BlobInfo, error) {
	log := klog.FromContext(ctx)

	p := s.pathOnDisk(hash)
	uploadPath := p + ".tmp-" + strconv.Itoa(rand.Int())

	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return nil, fmt.Errorf("creating directories: %w", err)
	}

	// TODO: What if file already exists?

	f, err := os.Create(uploadPath)
	if err != nil {
		return nil, fmt.Errorf("creating file %q: %w", uploadPath, err)
	}

	shouldDelete := true
	defer func() {
		if shouldDelete {
			if err := os.Remove(uploadPath); err != nil {
				log.Error(err, "removing temp file", "path", uploadPath)
			}
		}
	}()

	shouldClose := true
	defer func() {
		if shouldClose {
			if err := f.Close(); err != nil {
				log.Error(err, "closing temp file", "path", uploadPath)
			}
		}
	}()

	n, err := io.Copy(f, r)
	if err != nil {
		return nil, fmt.Errorf("writing to temp file %q: %w", uploadPath, err)
	}

	if err := f.Close(); err != nil {
		return nil, fmt.Errorf("closing temp file %q: %w", uploadPath, err)
	}
	shouldClose = false

	if err := os.Rename(uploadPath, p); err != nil {
		return nil, fmt.Errorf("renaming temp file %q -> %q: %w", uploadPath, p, err)
	}
	shouldDelete = false

	klog.Infof("creating blob at %q", p)

	blobInfo := &BlobInfo{
		Length: n,
	}
	return blobInfo, nil
}
