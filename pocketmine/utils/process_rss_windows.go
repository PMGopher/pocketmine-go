//go:build windows

package utils

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

// processMemoryCounters is psapi's PROCESS_MEMORY_COUNTERS.
type processMemoryCounters struct {
	cb                         uint32
	pageFaultCount             uint32
	peakWorkingSetSize         uintptr
	workingSetSize             uintptr
	quotaPeakPagedPoolUsage    uintptr
	quotaPagedPoolUsage        uintptr
	quotaPeakNonPagedPoolUsage uintptr
	quotaNonPagedPoolUsage     uintptr
	pagefileUsage              uintptr
	peakPagefileUsage          uintptr
}

var procGetProcessMemoryInfo = windows.NewLazySystemDLL("psapi.dll").NewProc("GetProcessMemoryInfo")

// ProcessRSS returns the process's resident set size in bytes: its working set, from psapi's
// GetProcessMemoryInfo.
func ProcessRSS() (uint64, bool) {
	var counters processMemoryCounters
	counters.cb = uint32(unsafe.Sizeof(counters))
	r, _, _ := procGetProcessMemoryInfo.Call(uintptr(windows.CurrentProcess()), uintptr(unsafe.Pointer(&counters)), uintptr(counters.cb))
	if r == 0 {
		return 0, false
	}
	return uint64(counters.workingSetSize), true
}
