package rtsp

import (
	"github.com/onedss/onedss/core/logger"
	"log"
)

type SessionLogger struct {
	innerLogger *log.Logger
}

func (sessionLogger *SessionLogger) GetLogger() *log.Logger {
	if sessionLogger != nil {
		return sessionLogger.innerLogger
	} else {
		return logger.InfoLogger
	}
}
