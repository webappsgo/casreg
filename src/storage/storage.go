package storage

import (
	"io"

	"github.com/casapps/casreg/src/config"
)

// Storage is the backend-agnostic interface for blob and manifest persistence.
type Storage interface {
	// Get retrieves a blob by digest, returning a ReadCloser.
	Get(registry, repository, digest string) (io.ReadCloser, error)

	// Put stores a blob and returns its content-addressed digest.
	Put(registry, repository string, r io.Reader, size int64) (digest string, err error)

	// Delete removes a blob by digest.
	Delete(registry, repository, digest string) error

	// Exists reports whether a blob with the given digest is present.
	Exists(registry, repository, digest string) (bool, error)

	// List returns all digests stored under the given repository.
	List(registry, repository string) ([]string, error)
}

// NewLocalStorage returns a local-filesystem storage backend.
func NewLocalStorage(cfg *config.Config) (Storage, error) {
	return &localStore{base: cfg.Storage.Path}, nil
}

// NewS3Storage returns an S3-compatible storage backend.
func NewS3Storage(cfg *config.Config) (Storage, error) {
	return &s3Store{cfg: cfg}, nil
}

// NewNFSStorage returns an NFS-backed storage backend.
func NewNFSStorage(cfg *config.Config) (Storage, error) {
	return &nfsStore{cfg: cfg}, nil
}
