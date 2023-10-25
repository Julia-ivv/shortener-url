package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
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

var ZapSugar *zap.SugaredLogger

func (res *logResponseWriter) Write(b []byte) (int, error) {
	size, err := res.ResponseWriter.Write(b)
	res.responseInfo.size += size
	return size, err
}

func (res *logResponseWriter) WriteHeader(statusCode int) {
	res.ResponseWriter.WriteHeader(statusCode)
	res.responseInfo.status = statusCode
}

func HandlerWithLogging(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(
		func(res http.ResponseWriter, req *http.Request) {
			start := time.Now()
			responseInfo := &responseInfo{
				status: 0,
				size:   0,
			}
			logResponseWriter := logResponseWriter{
				ResponseWriter: res,
				responseInfo:   responseInfo,
			}
			uri := req.RequestURI
			method := req.Method

			h(&logResponseWriter, req)
			duration := time.Since(start)

			ZapSugar.Infoln(
				"uri", uri,
				"method", method,
				"status", responseInfo.status,
				"size", responseInfo.size,
				"duration", duration,
			)
		})
}

func NewLogger() *zap.SugaredLogger {
	log, errLog := zap.NewDevelopment()
	if errLog != nil {
		panic(errLog)
	}
	defer log.Sync()

	zapSugar := log.Sugar()

	return zapSugar
}
