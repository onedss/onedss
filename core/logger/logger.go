package logger

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
)

var (
	TraceLogger   *log.Logger
	InfoLogger    *log.Logger
	WarningLogger *log.Logger
	ErrorLogger   *log.Logger
)

func init() {
	file, err := os.OpenFile("errors.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Failed to open error log file:", err)
	}
	TraceLogger = log.New(ioutil.Discard, "[TRACE] ", log.Ldate|log.Ltime|log.Lshortfile)

	InfoLogger = log.New(os.Stdout, "[INFO] ", log.Ldate|log.Ltime|log.Lshortfile)

	WarningLogger = log.New(os.Stdout, "[WARNING] ", log.Ldate|log.Ltime|log.Lshortfile)

	ErrorLogger = log.New(io.MultiWriter(file, os.Stderr), "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile)
}

func Trace(v ...interface{}) {
	TraceLogger.Output(2, fmt.Sprintln(v...))
}

func Tracef(format string, v ...interface{}) {
	TraceLogger.Output(2, fmt.Sprintf(format, v...))
}

func Info(v ...interface{}) {
	InfoLogger.Output(2, fmt.Sprintln(v...))
}

func Infof(format string, v ...interface{}) {
	InfoLogger.Output(2, fmt.Sprintf(format, v...))
}

func Warning(v ...interface{}) {
	WarningLogger.Output(2, fmt.Sprintln(v...))
}

func Warningf(format string, v ...interface{}) {
	WarningLogger.Output(2, fmt.Sprintf(format, v...))
}

func Error(v ...interface{}) {
	ErrorLogger.Output(2, fmt.Sprintln(v...))
}

func Errorf(format string, v ...interface{}) {
	ErrorLogger.Output(2, fmt.Sprintf(format, v...))
}
