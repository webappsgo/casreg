package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"
)

// compressionWriter wraps http.ResponseWriter to provide compression
type compressionWriter struct {
	io.Writer
	http.ResponseWriter
	wroteHeader bool
}

func (cw *compressionWriter) WriteHeader(code int) {
	if !cw.wroteHeader {
		// Remove Content-Length header as it will change after compression
		cw.Header().Del("Content-Length")
		cw.ResponseWriter.WriteHeader(code)
		cw.wroteHeader = true
	}
}

func (cw *compressionWriter) Write(b []byte) (int, error) {
	if !cw.wroteHeader {
		cw.WriteHeader(http.StatusOK)
	}
	return cw.Writer.Write(b)
}

// Pool of gzip writers for reuse
var gzipWriterPool = sync.Pool{
	New: func() interface{} {
		return gzip.NewWriter(nil)
	},
}

// Compression middleware provides gzip compression for responses
func Compression(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if client accepts gzip
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		// Skip compression for certain content types
		contentType := w.Header().Get("Content-Type")
		if shouldSkipCompression(contentType) {
			next.ServeHTTP(w, r)
			return
		}

		// Skip compression for Docker registry blob requests (already compressed)
		if strings.HasPrefix(r.URL.Path, "/v2/") && strings.Contains(r.URL.Path, "/blobs/") {
			next.ServeHTTP(w, r)
			return
		}

		// Get gzip writer from pool
		gz := gzipWriterPool.Get().(*gzip.Writer)
		defer gzipWriterPool.Put(gz)

		gz.Reset(w)
		defer gz.Close()

		// Set compression headers
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Add("Vary", "Accept-Encoding")

		// Wrap response writer
		cw := &compressionWriter{
			Writer:         gz,
			ResponseWriter: w,
		}

		next.ServeHTTP(cw, r)
	})
}

// shouldSkipCompression determines if compression should be skipped based on content type
func shouldSkipCompression(contentType string) bool {
	// Skip compression for already compressed formats
	skipTypes := []string{
		"image/",          // Images (jpeg, png, gif, etc.)
		"video/",          // Videos
		"audio/",          // Audio
		"application/zip", // Zip files
		"application/gzip", // Gzip files
		"application/x-gzip",
		"application/x-compress",
		"application/x-compressed",
		"application/octet-stream", // Binary data (often already compressed)
	}

	for _, skipType := range skipTypes {
		if strings.HasPrefix(contentType, skipType) {
			return true
		}
	}

	return false
}
