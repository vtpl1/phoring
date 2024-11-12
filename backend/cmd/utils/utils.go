// Package utils exports all utility functions
package utils

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

// GetAppName returns Application name
func GetAppName() string {
	file, err := os.Executable()
	if err != nil {
		return "ojana"
	}
	return filepath.Base(file)
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
		err := os.Mkdir(path, 0o640)
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

func newLogger() zerolog.Logger {

	lumberjackLogger := &lumberjack.Logger{
		Filename:   GetLogFileName(),
		MaxSize:    1,
		MaxBackups: 3,
	}
	return zerolog.New(lumberjackLogger).With().Timestamp().Logger()
}

// GetLogger returns a logger
func GetLogger(_ string) zerolog.Logger {
	if logger == nil {
		logger1 := newLogger()
		logger = &logger1
		return *logger
	}
	return *logger
}

// nolint:gochecknoglobals
var logger *zerolog.Logger
