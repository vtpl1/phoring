// Package utils exports all utility functions
package utils

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger wraps zerolog.Logger with an embedded lumberjack.Logger for log rotation.
type Logger struct {
	zerolog.Logger
	rotator *lumberjack.Logger // Keep a reference to close when done
}

// GetAppName returns Application name
func GetAppName() string {
	file, err := os.Executable()
	if err != nil {
		return "ojana"
	}
	file = filepath.Base(file)
	names := strings.Split(file, ".")
	if len(names) > 1 {
		return names[0]
	}
	return file
}

// GetAppRunningDir returns Application running directory
func GetAppRunningDir() string {
	file, err := os.Getwd()
	if err != nil {
		return "./"
	}
	return file
}

// GetSessionDir returns session directory
func GetSessionDir() string {
	return path.Join(GetAppRunningDir(), "session")
}

// GetLogDir returns log directory
func GetLogDir() string {
	return path.Join(GetAppRunningDir(), "logs", GetAppName())
}

// GetLibDir returns lib directory
func GetLibDir() string {
	return path.Join(GetAppRunningDir(), "lib")
}

// CreateDir creates directory if it does not exists
func CreateDir(path string) string {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.MkdirAll(path, 0o640)
		if err != nil {
			// nolint:forbidigo
			panic("Can not create file")
		}
	}
	return path
}

func GetLogFileName() string {
	return filepath.Join(GetLogDir(), fmt.Sprintf("%s.log", GetAppName()))
}

func GetLockFileName() string {
	return filepath.Join(os.TempDir(), fmt.Sprintf("%s.lock", GetAppName()))
}

func NewLogger() *Logger {

	rotator := &lumberjack.Logger{
		Filename:   GetLogFileName(),
		MaxSize:    20, // Max size in MB before rotation
		MaxBackups: 3,  // Max number of backup files
	}
	logger := zerolog.New(rotator).With().Timestamp().Logger()
	// Set the logging level based on environment variable
	if os.Getenv("LOG_LEVEL") == "DEBUG" {
		logger = logger.Level(zerolog.DebugLevel)
	} else {
		logger = logger.Level(zerolog.InfoLevel)
	}
	return &Logger{
		Logger:  logger,
		rotator: rotator,
	}
}

// Close releases resources held by the logger, especially the lumberjack rotator.
func (l *Logger) Close() error {
	return l.rotator.Close()
}

// GetLogger returns a logger
func GetLogger(_ string) Logger {
	if logger == nil {
		logger = NewLogger()

		return *logger
	}
	return *logger
}

func CloseLogger() {
	if logger != nil {
		logger.rotator.Close()
	}
}

// nolint:gochecknoglobals
var logger *Logger
