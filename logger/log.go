package logger

import (
	"encoding/json"
	"fmt"
	"os"
)

// LogInfo represents the type for structured loggin
type LogInfo struct {
	LogLevel string `json:"logLevel"`
	Message  string `json:"message"`
}

func write(info *LogInfo) {
	b, _ := json.Marshal(info)
	fmt.Printf(string(b))
}

// Debug log the message with debug log level
func Debug(message string, args ...interface{}) {
	info := &LogInfo{
		LogLevel: "debug",
		Message:  fmt.Sprintf(message, args...),
	}
	write(info)
}

// Info log the message with info log level
func Info(message string, args ...interface{}) {
	info := &LogInfo{
		LogLevel: "info",
		Message:  fmt.Sprintf(message, args...),
	}
	write(info)
}

// Error log the message with error log level
func Error(message string, args ...interface{}) {
	info := &LogInfo{
		LogLevel: "error",
		Message:  fmt.Sprintf(message, args...),
	}
	write(info)
}

// Fatal log the message as critical error and exits
func Fatal(message string, args ...interface{}) {
	info := &LogInfo{
		LogLevel: "fatal",
		Message:  fmt.Sprintf(message, args...),
	}
	write(info)
	os.Exit(1)
}
