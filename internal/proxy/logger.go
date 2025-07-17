package proxy

import (
	"fmt"
	"log"
	"os"
)

type StandardLogger struct {
	infoLogger  *log.Logger
	errorLogger *log.Logger
	debugLogger *log.Logger
}

func NewStandardLogger() Logger {
	return &StandardLogger{
		infoLogger:  log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		errorLogger: log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
		debugLogger: log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

func (l *StandardLogger) Info(msg string, args ...interface{}) {
	l.infoLogger.Printf(msg, args...)
}

func (l *StandardLogger) Error(msg string, args ...interface{}) {
	l.errorLogger.Printf(msg, args...)
}

func (l *StandardLogger) Debug(msg string, args ...interface{}) {
	l.debugLogger.Printf(msg, args...)
}

type SimpleLogger struct{}

func NewSimpleLogger() Logger {
	return &SimpleLogger{}
}

func (l *SimpleLogger) Info(msg string, args ...interface{}) {
	fmt.Printf(msg+"\n", args...)
}

func (l *SimpleLogger) Error(msg string, args ...interface{}) {
	fmt.Printf("ERROR: "+msg+"\n", args...)
}

func (l *SimpleLogger) Debug(msg string, args ...interface{}) {
	fmt.Printf("DEBUG: "+msg+"\n", args...)
}