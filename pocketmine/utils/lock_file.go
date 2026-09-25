package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"sync"
)

var (
	lockFileHandlesMu sync.Mutex
	lockFileHandles   = map[string]*os.File{}
)

// CreateLockFile is a port of Filesystem::createLockFile: locks the file so no other process can
// lock it, and writes this process's PID to it. It returns nil on success, or the PID of the
// process holding the lock (-1 if it can't be read). An error means the file couldn't be opened.
//
// PHP downgrades its exclusive lock to a shared one so other processes can read the PID; the
// lock here is advisory (flock) or covers a byte range past the PID (Windows), so the PID stays
// readable while the exclusive lock is held.
func CreateLockFile(lockFilePath string) (*int, error) {
	f, err := os.OpenFile(lockFilePath, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, fmt.Errorf("Failed to open lock file: %w", err)
	}
	if !tryLockFile(f) {
		defer f.Close()
		content, _ := io.ReadAll(f)
		pid := -1
		if regexp.MustCompile(`^\d+$`).Match(content) {
			pid, _ = strconv.Atoi(string(content))
		}
		return &pid, nil
	}
	_ = f.Truncate(0)
	_, _ = f.WriteAt([]byte(strconv.Itoa(os.Getpid())), 0)
	_ = f.Sync()

	abs, _ := filepath.Abs(lockFilePath)
	lockFileHandlesMu.Lock()
	lockFileHandles[abs] = f //keep the file open to preserve the lock
	lockFileHandlesMu.Unlock()
	return nil, nil
}

// ReleaseLockFile is a port of Filesystem::releaseLockFile: releases a lock taken by
// CreateLockFile and deletes the lock file.
func ReleaseLockFile(lockFilePath string) error {
	abs, err := filepath.Abs(lockFilePath)
	if err != nil {
		return fmt.Errorf("Invalid lock file path")
	}
	lockFileHandlesMu.Lock()
	f, ok := lockFileHandles[abs]
	delete(lockFileHandles, abs)
	lockFileHandlesMu.Unlock()
	if ok {
		unlockFile(f)
		_ = f.Close()
		_ = os.Remove(abs)
	}
	return nil
}
