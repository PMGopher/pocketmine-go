package utils

import (
	"os"
	"os/exec"
	"os/user"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/google/uuid"
)

var (
	serverUniqueIDMu sync.Mutex
	serverUniqueID   *uuid.UUID
)

// GetMachineUniqueID is a port of Utils::getMachineUniqueId: a UUID made from machine-specific
// information (uname, CPU, temp dir, machine ID or MAC addresses, user). PHP adds its PHP_* build
// constants and loaded extensions; the Go runtime version and architecture take their place.
func GetMachineUniqueID(extra string) uuid.UUID {
	serverUniqueIDMu.Lock()
	defer serverUniqueIDMu.Unlock()
	if serverUniqueID != nil && extra == "" {
		return *serverUniqueID
	}

	hostname, _ := os.Hostname()
	machine := runtime.GOOS + " " + hostname + " " + runtime.GOARCH
	if cpuinfo, err := os.ReadFile("/proc/cpuinfo"); err == nil {
		for _, line := range strings.SplitAfter(string(cpuinfo), "\n") {
			if strings.Contains(line, "model name") || strings.Contains(line, "Processor") || strings.Contains(line, "Serial") {
				machine += line
			}
		}
	}
	machine += os.TempDir()
	machine += extra
	switch GetOS() {
	case OSWindows:
		if out, err := exec.Command("ipconfig", "/ALL").Output(); err == nil {
			var macs []string
			for _, m := range regexp.MustCompile(`Physical Address[. ]{1,}: ([0-9A-F\-]{17})`).FindAllStringSubmatch(string(out), -1) {
				if m[1] != "00-00-00-00-00-00" {
					macs = append(macs, m[1])
				}
			}
			machine += strings.Join(macs, " ") //Mac Addresses
		}
	case OSLinux:
		if id, err := os.ReadFile("/etc/machine-id"); err == nil {
			machine += string(id)
		}
	case OSMacOS:
		if out, err := exec.Command("sh", "-c", "system_profiler SPHardwareDataType | grep UUID").Output(); err == nil {
			machine += string(out)
		}
	}
	data := machine + strconv.Itoa(4096)
	data += strconv.FormatInt(int64(^uint64(0)>>1), 10)
	data += "8"
	if u, err := user.Current(); err == nil {
		data += u.Username
	}
	data += runtime.Version()

	//TODO: use of NIL as namespace is a hack; it works for now, but we should have a proper namespace UUID
	id := uuid.NewMD5(uuid.Nil, []byte(data))
	if extra == "" {
		serverUniqueID = &id
	}
	return id
}
