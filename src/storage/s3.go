package storage

import (
	"fmt"
	"io"

	"github.com/casapps/casreg/src/config"
)

type s3Store struct {
	cfg *config.Config
}

func (s *s3Store) Get(_, _, _ string) (io.ReadCloser, error) {
	return nil, fmt.Errorf("s3 storage: not yet implemented")
}

func (s *s3Store) Put(_, _ string, _ io.Reader, _ int64) (string, error) {
	return "", fmt.Errorf("s3 storage: not yet implemented")
}

func (s *s3Store) Delete(_, _, _ string) error {
	return fmt.Errorf("s3 storage: not yet implemented")
}

func (s *s3Store) Exists(_, _, _ string) (bool, error) {
	return false, fmt.Errorf("s3 storage: not yet implemented")
}

func (s *s3Store) List(_, _ string) ([]string, error) {
	return nil, fmt.Errorf("s3 storage: not yet implemented")
}
