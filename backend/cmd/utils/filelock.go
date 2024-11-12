package utils

import (
	"fmt"
	"os"
)

// FileLock simulates a mutex across program instances using a lock file.
type FileLock struct {
	filePath string
	file     *os.File
}

// NewFileLock creates a new FileLock for a given file path.
func NewFileLock(filePath string) *FileLock {
	return &FileLock{filePath: filePath}
}

// Lock attempts to create a lock file. If it exists, it indicates another instance is running.
func (fl *FileLock) Lock() error {
	file, err := os.OpenFile(fl.filePath, os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("another instance is already running")
		}
		return err
	}
	fl.file = file
	return nil
}

// Unlock removes the lock file to allow future instances to run.
func (fl *FileLock) Unlock() error {
	if fl.file != nil {
		fl.file.Close()
	}
	return os.Remove(fl.filePath)
}
