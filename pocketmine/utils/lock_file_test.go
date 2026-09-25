package utils

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestCreateAndReleaseLockFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "server.lock")
	pid, err := CreateLockFile(path)
	if err != nil || pid != nil {
		t.Fatalf("first lock: pid=%v err=%v", pid, err)
	}
	content, _ := os.ReadFile(path)
	if string(content) != strconv.Itoa(os.Getpid()) {
		t.Errorf("lock file holds %q, want our PID", content)
	}
	if err := ReleaseLockFile(path); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("lock file wasn't deleted")
	}
}
