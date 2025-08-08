// utils/logger.go
package utils

import (
	"log"
	"os"
)

var Logger *log.Logger

func init() {
	// Simple logger to stdout, you can configure it further (e.g., file output, levels)
	Logger = log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)
}

// Info logs an info message.
func Info(v ...interface{}) {
	Logger.Println("INFO:", v)
}

// Infof logs an info message with formatting.
func Infof(format string, v ...interface{}) {
	Logger.Printf("INFO: "+format, v...)
}

// Error logs an error message.
func Error(v ...interface{}) {
	Logger.Println("ERROR:", v)
}

// Errorf logs an error message with formatting.
func Errorf(format string, v ...interface{}) {
	Logger.Printf("ERROR: "+format, v...)
}

// Debug logs a debug message (conditional logging example).
func Debug(v ...interface{}) {
	// You could check an environment variable like DEBUG=true to enable
	if os.Getenv("DEBUG") == "true" {
		Logger.Println("DEBUG:", v)
	}
}
