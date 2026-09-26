//go:build linux || android

package utils

import (
	"os"
	"strconv"
	"strings"
)

// ProcessRSS returns the process's resident set size (memory actually in RAM) in bytes, read from
// /proc/self/statm.
func ProcessRSS() (uint64, bool) {
	data, err := os.ReadFile("/proc/self/statm")
	if err != nil {
		return 0, false
	}
	fields := strings.Fields(string(data))
	if len(fields) < 2 {
		return 0, false
	}
	pages, err := strconv.ParseUint(fields[1], 10, 64)
	if err != nil {
		return 0, false
	}
	return pages * uint64(os.Getpagesize()), true
}
