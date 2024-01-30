// Package logger creates a logger instance.
package logger

import (
	"go.uber.org/zap"
)

// ZapSugar provides access to the logger.
var ZapSugar *zap.SugaredLogger

// NewLogger creates a logger instance.
func NewLogger() *zap.SugaredLogger {
	log, errLog := zap.NewDevelopment()
	if errLog != nil {
		panic(errLog)
	}
	defer log.Sync()

	zapSugar := log.Sugar()

	return zapSugar
}
