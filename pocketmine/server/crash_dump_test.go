package server

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestTickPanicWritesCrashDump(t *testing.T) {
	exited := false
	oldExit, oldSleep := crashExit, crashThrottleSleep
	crashExit = func() { exited = true }
	crashThrottleSleep = func(time.Duration) {}
	defer func() { crashExit, crashThrottleSleep = oldExit, oldSleep }()

	s := newTestServer(t)
	s.isRunning.Store(true)
	s.mu.Lock()
	func() {
		defer func() {
			if p := recover(); p != nil {
				t.Fatalf("the panic escaped: %v", p)
			}
		}()
		// A tick that panics, like a bug in a plugin's task or a block update.
		s.nextTick = time.Now()
		s.worldManager = nil
		s.tickOrCrash()
	}()
	s.mu.Unlock()

	if !exited {
		t.Fatal("the server didn't exit after the crash")
	}
	entries, err := os.ReadDir(filepath.Join(s.GetDataPath(), "crashdumps"))
	if err != nil || len(entries) != 1 {
		t.Fatalf("crashdumps = %v (%v)", entries, err)
	}
	data, err := os.ReadFile(filepath.Join(s.GetDataPath(), "crashdumps", entries[0].Name()))
	if err != nil {
		t.Fatal(err)
	}
	dump := string(data)
	for _, want := range []string{"PocketMine-MP Crash Dump", "Go version: go", "Error: runtime error: invalid memory address or nil pointer dereference", "Backtrace:", "(*Server).tick", "===BEGIN CRASH DUMP===", "===END CRASH DUMP==="} {
		if !strings.Contains(dump, want) {
			t.Errorf("crash dump is missing %q:\n%s", want, dump)
		}
	}
	if s.isRunning.Load() {
		t.Error("the server is still marked running")
	}
}
