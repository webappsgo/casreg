package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
)

// writeAndDigest copies r into w and returns the sha256 digest in Docker digest format.
func writeAndDigest(w io.Writer, r io.Reader) (string, error) {
	h := sha256.New()
	if _, err := io.Copy(io.MultiWriter(w, h), r); err != nil {
		return "", fmt.Errorf("digest copy: %w", err)
	}
	return "sha256:" + hex.EncodeToString(h.Sum(nil)), nil
}
