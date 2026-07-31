package utils

import (
	"encoding/hex"
	"fmt"
	"math"
	"math/rand"
	"regexp"
	"runtime"
	"strings"
	"unicode/utf8"
)

// Utils is a port of the portable subset of pocketmine\utils\Utils.
//
// Deliberately NOT ported: getNiceClosureName/getNiceClassName/parseDocComment/
// validateCallableSignature/testValidInstance/getReferenceCount. These all lean on PHP's
// runtime reflection over a dynamically-typed, interpreted program (ReflectionFunction,
// debug_zval_dump, doc-comment parsing). Go is statically compiled and typed, so the problems
// they solve (checking a callable's signature at runtime, describing an arbitrary class by
// name, counting live references to a value) either don't arise (the compiler already
// enforces it) or have no equivalent hook to hang a port off of.
const (
	OSWindows = "win"
	OSIOS     = "ios"
	OSMacOS   = "mac"
	OSAndroid = "android"
	OSLinux   = "linux"
	OSBSD     = "bsd"
	OSUnknown = "other"
)

// GetOS returns one of the OS* constants.
//
// PHP's original parses php_uname() output to guess the OS; Go's runtime.GOOS already reports
// this directly for every platform PocketMine targets, including "android" and "ios" when
// cross-compiled for those targets, so no parsing is needed.
func GetOS() string {
	switch runtime.GOOS {
	case "windows":
		return OSWindows
	case "ios":
		return OSIOS
	case "darwin":
		return OSMacOS
	case "android":
		return OSAndroid
	case "linux":
		return OSLinux
	case "freebsd", "netbsd", "openbsd", "dragonfly":
		return OSBSD
	default:
		return OSUnknown
	}
}

// GetCoreCount returns the number of logical CPU cores.
//
// PHP's original shells out to OS-specific tools (/proc/cpuinfo, sysctl, %NUMBER_OF_PROCESSORS%);
// Go's runtime.NumCPU() already does this portably.
func GetCoreCount() int {
	return runtime.NumCPU()
}

// chunkSplit mirrors PHP's chunk_split(): splits s into chunks of chunkLen, appending end
// after every chunk, including the last (possibly shorter) one.
func chunkSplit(s string, chunkLen int, end string) string {
	var b strings.Builder
	for i := 0; i < len(s); i += chunkLen {
		e := i + chunkLen
		if e > len(s) {
			e = len(s)
		}
		b.WriteString(s[i:e])
		b.WriteString(end)
	}
	return b.String()
}

// Hexdump returns a prettified hexdump, 16 bytes per line, grouped 8-and-8 like the PHP
// original's nested chunk_split() calls.
func Hexdump(bin []byte) string {
	var out strings.Builder
	for offset := 0; offset < len(bin); offset += 16 {
		end := offset + 16
		if end > len(bin) {
			end = len(bin)
		}
		line := bin[offset:end]

		hexPart := hex.EncodeToString(line)
		hexPart += strings.Repeat(" ", 32-len(hexPart))
		groupedHex := chunkSplit(chunkSplit(hexPart, 2, " "), 24, " ")

		ascii := make([]byte, len(line))
		for i, b := range line {
			if b < 0x20 || b > 0x7E {
				ascii[i] = '.'
			} else {
				ascii[i] = b
			}
		}

		fmt.Fprintf(&out, "%04x  %s %s\n", offset, groupedHex, ascii)
	}
	return out.String()
}

var nonPrintablePattern = regexp.MustCompile(`[^\x20-\x7E]`)

// Printable returns a string with non-printable characters replaced, suitable for logging.
func Printable(s string) string {
	return nonPrintablePattern.ReplaceAllString(s, ".")
}

// JavaStringHash replicates java.lang.String#hashCode() over the raw bytes of s (matching the
// PHP original, which operates on PHP's byte-string representation rather than UTF-16 code
// units, so this only matches real Java hashCode() for ASCII strings).
//
// The PHP original masks with `& 0xFFFFFFFF` every iteration, which keeps the accumulator as a
// non-negative magnitude in PHP's 64-bit int rather than a wrapped signed 32-bit value; the
// return type here is uint32 to preserve that same magnitude instead of silently becoming
// negative the way a Go int32 would for hashes with the top bit set.
func JavaStringHash(s string) uint32 {
	var hash int32
	for i := 0; i < len(s); i++ {
		hash = 31*hash + int32(int8(s[i]))
	}
	return uint32(hash)
}

// CheckUTF8 mirrors Utils::checkUTF8().
func CheckUTF8(s string) error {
	if !utf8.ValidString(s) {
		return fmt.Errorf("text must be valid UTF-8")
	}
	return nil
}

// CheckFloatNotInfOrNaN mirrors Utils::checkFloatNotInfOrNaN().
func CheckFloatNotInfOrNaN(name string, f float64) error {
	if math.IsNaN(f) {
		return fmt.Errorf("%s cannot be NaN", name)
	}
	if math.IsInf(f, 0) {
		return fmt.Errorf("%s cannot be infinite", name)
	}
	return nil
}

// GetRandomFloat returns a random float between 0.0 and 1.0.
func GetRandomFloat() float64 {
	return rand.Float64()
}
