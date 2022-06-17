package core

import (
	"log"
	"os"
)

type SessionLogger struct {
	innerLogger *log.Logger
}

func NewSessionLogger(prefix string) SessionLogger {
	sessionLogger := SessionLogger{
		innerLogger: log.New(os.Stdout, prefix, log.LstdFlags|log.Lshortfile|log.Lmicroseconds),
	}
	return sessionLogger
}

func (sessionLogger *SessionLogger) GetLogger() *log.Logger {
	if sessionLogger.innerLogger != nil {
		return sessionLogger.innerLogger
	} else {
		return InfoLogger
	}
}

func (sessionLogger *SessionLogger) SetLogger(logger *log.Logger) {
	sessionLogger.innerLogger = logger
}
