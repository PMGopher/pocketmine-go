//go:build windows

package utils

import (
	"os"

	"golang.org/x/sys/windows"
)

// lockOffset is where the locked byte range starts: past any PID, so it stays readable.
const lockOffset = 1 << 30

func tryLockFile(f *os.File) bool {
	ol := &windows.Overlapped{Offset: lockOffset}
	return windows.LockFileEx(windows.Handle(f.Fd()), windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY, 0, 1, 0, ol) == nil
}

func unlockFile(f *os.File) {
	ol := &windows.Overlapped{Offset: lockOffset}
	_ = windows.UnlockFileEx(windows.Handle(f.Fd()), 0, 1, 0, ol)
}
