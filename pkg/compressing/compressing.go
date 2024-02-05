// Package compressing implements gzip for middleware compression.
package compressing

import (
	"compress/gzip"
	"io"
	"net/http"
)

type gzipWriter struct {
	http.ResponseWriter
	Writer *gzip.Writer
}

// Write overrides the Write method to send a compressed response.
func (w *gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// NewGzipWriter creates a new instance for the gzip packager.
func NewGzipWriter(w http.ResponseWriter) *gzipWriter {
	return &gzipWriter{
		ResponseWriter: w,
		Writer:         gzip.NewWriter(w),
	}
}

type gzipReader struct {
	readCloser io.ReadCloser
	zipReader  *gzip.Reader
}

// NewGzipReader creates a new instance for unpacking data.
func NewGzipReader(r io.ReadCloser) (*gzipReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	return &gzipReader{
		readCloser: r,
		zipReader:  zr,
	}, nil
}

// Read overrides the Read method to unpack data.
func (r *gzipReader) Read(b []byte) (n int, err error) {
	return r.zipReader.Read(b)
}

// Close gzip.
func (r *gzipReader) Close() error {
	if err := r.readCloser.Close(); err != nil {
		return err
	}
	return r.zipReader.Close()
}
