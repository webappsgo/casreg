package storage

import (
	"fmt"
	"io"

	"github.com/casapps/casreg/src/config"
)

type nfsStore struct {
	cfg *config.Config
}

func (s *nfsStore) Get(_, _, _ string) (io.ReadCloser, error) {
	return nil, fmt.Errorf("nfs storage: not yet implemented")
}

func (s *nfsStore) Put(_, _ string, _ io.Reader, _ int64) (string, error) {
	return "", fmt.Errorf("nfs storage: not yet implemented")
}

func (s *nfsStore) Delete(_, _, _ string) error {
	return fmt.Errorf("nfs storage: not yet implemented")
}

func (s *nfsStore) Exists(_, _, _ string) (bool, error) {
	return false, fmt.Errorf("nfs storage: not yet implemented")
}

func (s *nfsStore) List(_, _ string) ([]string, error) {
	return nil, fmt.Errorf("nfs storage: not yet implemented")
}
