package middleware

import (
	"net/http"
	"time"

	"github.com/Julia-ivv/shortener-url/pkg/logger"
)

type (
	responseInfo struct {
		status int
		size   int
	}

	logResponseWriter struct {
		http.ResponseWriter
		responseInfo *responseInfo
	}
)

// Write overrides the Write method to add the response size to the logs.
func (res *logResponseWriter) Write(b []byte) (int, error) {
	size, err := res.ResponseWriter.Write(b)
	res.responseInfo.size += size
	return size, err
}

// WriteHeader overrides the Write method to add a status code to the logs.
func (res *logResponseWriter) WriteHeader(statusCode int) {
	res.ResponseWriter.WriteHeader(statusCode)
	res.responseInfo.status = statusCode
}

// HandlerWithLogging adds logging to the handler.
func HandlerWithLogging(h http.Handler) http.Handler {
	return http.HandlerFunc(
		func(res http.ResponseWriter, req *http.Request) {
			start := time.Now()
			responseInfo := &responseInfo{
				status: 0,
				size:   0,
			}
			logRespWriter := logResponseWriter{
				ResponseWriter: res,
				responseInfo:   responseInfo,
			}
			uri := req.RequestURI
			method := req.Method

			h.ServeHTTP(&logRespWriter, req)
			duration := time.Since(start)

			logger.ZapSugar.Infoln(
				"uri", uri,
				"method", method,
				"status", responseInfo.status,
				"size", responseInfo.size,
				"duration", duration,
			)
		})
}
