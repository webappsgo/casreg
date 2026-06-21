package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type localStore struct {
	base string
}

func (s *localStore) blobPath(registry, repository, digest string) string {
	return filepath.Join(s.base, registry, repository, digest)
}

func (s *localStore) Get(registry, repository, digest string) (io.ReadCloser, error) {
	f, err := os.Open(s.blobPath(registry, repository, digest))
	if err != nil {
		return nil, fmt.Errorf("storage get: %w", err)
	}
	return f, nil
}

func (s *localStore) Put(registry, repository string, r io.Reader, _ int64) (string, error) {
	dir := filepath.Join(s.base, registry, repository)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("storage put mkdir: %w", err)
	}
	tmp, err := os.CreateTemp(dir, ".upload-*")
	if err != nil {
		return "", fmt.Errorf("storage put temp: %w", err)
	}
	defer func() {
		tmp.Close()
		os.Remove(tmp.Name())
	}()
	digest, err := writeAndDigest(tmp, r)
	if err != nil {
		return "", fmt.Errorf("storage put write: %w", err)
	}
	dest := filepath.Join(dir, digest)
	if err := os.Rename(tmp.Name(), dest); err != nil {
		return "", fmt.Errorf("storage put rename: %w", err)
	}
	return digest, nil
}

func (s *localStore) Delete(registry, repository, digest string) error {
	if err := os.Remove(s.blobPath(registry, repository, digest)); err != nil {
		return fmt.Errorf("storage delete: %w", err)
	}
	return nil
}

func (s *localStore) Exists(registry, repository, digest string) (bool, error) {
	_, err := os.Stat(s.blobPath(registry, repository, digest))
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("storage exists: %w", err)
	}
	return true, nil
}

func (s *localStore) List(registry, repository string) ([]string, error) {
	dir := filepath.Join(s.base, registry, repository)
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("storage list: %w", err)
	}
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			out = append(out, e.Name())
		}
	}
	return out, nil
}
