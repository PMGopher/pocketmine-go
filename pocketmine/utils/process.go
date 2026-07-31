package utils

import (
	"bytes"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
)

// AdvancedMemoryUsage returns [reserved, VmRSS, VmSize] in bytes, mirroring
// Process::getAdvancedMemoryUsage(). Go's runtime.MemStats reports the Go heap, which is used
// as the fallback on platforms without /proc, exactly like the PHP original falls back to
// memory_get_usage().
func AdvancedMemoryUsage() (reserved, vmRSS, vmSize uint64) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	reserved = m.HeapAlloc

	if GetOS() == OSLinux || GetOS() == OSAndroid {
		if status, err := os.ReadFile("/proc/self/status"); err == nil {
			if match := vmRSSPattern.FindSubmatch(status); match != nil {
				kb, _ := strconv.ParseUint(string(match[1]), 10, 64)
				vmRSS = kb * 1024
			}
			if match := vmSizePattern.FindSubmatch(status); match != nil {
				kb, _ := strconv.ParseUint(string(match[1]), 10, 64)
				vmSize = kb * 1024
			}
		}
	}

	if vmRSS == 0 {
		vmRSS = m.HeapAlloc
	}
	if vmSize == 0 {
		vmSize = m.Sys
	}
	return reserved, vmRSS, vmSize
}

var (
	vmRSSPattern    = regexp.MustCompile(`VmRSS:[ \t]+([0-9]+) kB`)
	vmSizePattern   = regexp.MustCompile(`VmSize:[ \t]+([0-9]+) kB`)
	threadsPattern  = regexp.MustCompile(`Threads:[ \t]+([0-9]+)`)
	mapsLinePattern = regexp.MustCompile(`([a-z0-9]+)-([a-z0-9]+) [rwxp-]{4} [a-z0-9]+ [^\[]*\[([a-zA-Z0-9]+)\]`)
)

func MemoryUsage() uint64 {
	_, rss, _ := AdvancedMemoryUsage()
	return rss
}

// RealMemoryUsage returns [heap, stack] bytes read from /proc/self/maps on Linux/Android; zero
// on other platforms, matching the PHP original's TODO for "more OS".
func RealMemoryUsage() (heap, stack uint64) {
	if GetOS() != OSLinux && GetOS() != OSAndroid {
		return 0, 0
	}
	data, err := os.ReadFile("/proc/self/maps")
	if err != nil {
		return 0, 0
	}
	for _, line := range bytes.Split(data, []byte("\n")) {
		m := mapsLinePattern.FindSubmatch(line)
		if m == nil {
			continue
		}
		start, _ := strconv.ParseUint(string(m[1]), 16, 64)
		end, _ := strconv.ParseUint(string(m[2]), 16, 64)
		switch {
		case bytes.HasPrefix(m[3], []byte("heap")):
			heap += end - start
		case bytes.HasPrefix(m[3], []byte("stack")):
			stack += end - start
		}
	}
	return heap, stack
}

// ThreadCount mirrors Process::getThreadCount(): reads the OS thread count on Linux/Android,
// and falls back to the Go scheduler's goroutine count elsewhere (a different unit, but the
// closest analogue Go exposes without OS-specific APIs).
func ThreadCount() int {
	if GetOS() == OSLinux || GetOS() == OSAndroid {
		if status, err := os.ReadFile("/proc/self/status"); err == nil {
			if m := threadsPattern.FindSubmatch(status); m != nil {
				n, _ := strconv.Atoi(string(m[1]))
				return n
			}
		}
	}
	return runtime.NumGoroutine()
}

// KillProcess mirrors Process::kill(): terminates the process with the given PID.
//
// os.Process.Kill() sends SIGKILL on Unix and calls TerminateProcess on Windows, so unlike the
// PHP original there's no need to branch on OS or shell out to taskkill/kill.
func KillProcess(pid int) error {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return proc.Kill()
}

// ExecuteProcess runs command via the shell and returns its stdout, stderr and exit code.
func ExecuteProcess(name string, args ...string) (stdout, stderr string, exitCode int) {
	cmd := exec.Command(name, args...)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	stdout, stderr = outBuf.String(), errBuf.String()
	if err == nil {
		return stdout, stderr, 0
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		return stdout, stderr, exitErr.ExitCode()
	}
	return stdout, stderr, -1
}

func Pid() int {
	return os.Getpid()
}

// Uid mirrors Process::uid(). os.Getuid() returns -1 on Windows without erroring (there's no
// concept of a POSIX uid there), matching the "doesn't work on this platform" case loosely.
func Uid() int {
	return os.Getuid()
}
