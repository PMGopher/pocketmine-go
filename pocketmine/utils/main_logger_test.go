package utils

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestMainLoggerWritesAndFlushes(t *testing.T) {
	dir := t.TempDir()
	logFile := filepath.Join(dir, "server.log")

	logger, err := NewMainLogger(logFile, false, "Main", time.UTC, true, "")
	if err != nil {
		t.Fatalf("NewMainLogger() error = %v", err)
	}

	logger.Info("hello world")
	logger.SyncFlushBuffer()
	logger.Shutdown()

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !strings.Contains(string(data), "hello world") {
		t.Fatalf("log file = %q, want it to contain %q", data, "hello world")
	}
	if !strings.Contains(string(data), "INFO") {
		t.Fatalf("log file = %q, want it to contain INFO prefix", data)
	}
}

func TestMainLoggerNoFileIsNoop(t *testing.T) {
	logger, err := NewMainLogger("", false, "Main", time.UTC, true, "")
	if err != nil {
		t.Fatalf("NewMainLogger() error = %v", err)
	}
	logger.Info("no disk write, no panic please")
	logger.SyncFlushBuffer()
	logger.Shutdown()
}
