package compressing

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type gzipWriter struct {
	http.ResponseWriter
	Writer *gzip.Writer
}

func (w *gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func newGzipWriter(w http.ResponseWriter) *gzipWriter {
	return &gzipWriter{
		ResponseWriter: w,
		Writer:         gzip.NewWriter(w),
	}
}

type gzipReader struct {
	readCloser io.ReadCloser
	zipReader  *gzip.Reader
}

func newGzipReader(r io.ReadCloser) (*gzipReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	return &gzipReader{
		readCloser: r,
		zipReader:  zr,
	}, nil
}

func (r *gzipReader) Read(b []byte) (n int, err error) {
	return r.zipReader.Read(b)
}

func (r *gzipReader) Close() error {
	if err := r.readCloser.Close(); err != nil {
		return err
	}
	return r.zipReader.Close()
}

func HandlerWithGzipCompression(h http.HandlerFunc) http.HandlerFunc {
	contentTypeForCompression := [2]string{"application/json", "text/html"}
	return http.HandlerFunc(
		func(res http.ResponseWriter, req *http.Request) {
			newRes := res
			acceptEncoding := req.Header.Values("Accept-Encoding")
			supportGzip := false
			for _, v := range acceptEncoding {
				if strings.Contains(v, "gzip") {
					supportGzip = true
				}
			}
			contentType := req.Header.Get("Content-Type")
			needsCompressing := false
			for _, v := range contentTypeForCompression {
				if strings.Contains(contentType, v) {
					needsCompressing = true
					break
				}
			}

			if supportGzip && needsCompressing {
				zw := newGzipWriter(res)
				newRes = zw
				defer zw.Writer.Close()
				zw.Header().Set("Content-Encoding", "gzip")
			}

			contentEncoding := req.Header.Get("Content-Encoding")
			gzipEncoding := strings.Contains(contentEncoding, "gzip")
			if gzipEncoding {
				zr, err := newGzipReader(req.Body)
				if err != nil {
					http.Error(newRes, err.Error(), http.StatusInternalServerError)
					return
				}
				req.Body = zr
				defer zr.Close()
			}

			h.ServeHTTP(newRes, req)
		})
}
