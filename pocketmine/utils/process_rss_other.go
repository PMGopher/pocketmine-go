//go:build !linux && !android && !windows

package utils

// ProcessRSS returns the process's resident set size in bytes. It isn't implemented on this
// platform (false).
func ProcessRSS() (uint64, bool) { return 0, false }
