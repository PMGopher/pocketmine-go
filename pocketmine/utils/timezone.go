package utils

import "time"

// Timezone is a port of pocketmine\utils\Timezone.
//
// The PHP original manually detects the OS timezone (Windows registry, /etc/timezone,
// /etc/sysconfig/clock, macOS's /etc/localtime symlink) and falls back to IP geolocation,
// because PHP's date extension needs an explicit ini timezone string and does not reliably
// auto-detect one across platforms. Go's time package already resolves the local timezone
// from the OS on every platform PocketMine targets (time.Local), so none of that detection
// logic is needed. Init is kept as a no-op purely so startup code that historically called
// Timezone::init() still has something to call.
func InitTimezone() {}

// GetTimezone returns the name of the currently configured local timezone.
func GetTimezone() string {
	return time.Local.String()
}
