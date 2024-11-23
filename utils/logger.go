package utils

import (
	"log"
	"os"
)

// Logger struct provides a centralized logging mechanism
type Logger struct {
	infoLogger  *log.Logger
	errorLogger *log.Logger
}

// NewLogger initializes and returns a new Logger instance
func NewLogger(logFile string) (*Logger, error) {
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return nil, err
	}

	return &Logger{
		infoLogger:  log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		errorLogger: log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
	}, nil
}

// Info logs informational messages
func (l *Logger) Info(message string) {
	l.infoLogger.Println(message)
}

// Error logs error messages
func (l *Logger) Error(err error) {
	l.errorLogger.Println(err)
}
