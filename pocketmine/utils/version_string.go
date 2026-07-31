package utils

import (
	"fmt"
	"regexp"
	"strconv"
)

var (
	versionStringPattern   = regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)(?:-(.*))?$`)
	baseVersionOnlyPattern = regexp.MustCompile(`^\d+\.\d+\.\d+(?:-(.*))?$`)
)

// VersionString is a port of pocketmine\utils\VersionString: manages and compares
// PocketMine-MP version strings.
type VersionString struct {
	baseVersion string
	isDevBuild  bool
	buildNumber int

	major, minor, patch int
	suffix              string
}

// NewVersionString mirrors the VersionString constructor, which fails (throws in PHP) if
// baseVersion doesn't contain at least 3 version digits.
func NewVersionString(baseVersion string, isDevBuild bool, buildNumber int) (*VersionString, error) {
	m := versionStringPattern.FindStringSubmatch(baseVersion)
	if m == nil {
		return nil, fmt.Errorf("invalid base version %q, should contain at least 3 version digits", baseVersion)
	}
	major, _ := strconv.Atoi(m[1])
	minor, _ := strconv.Atoi(m[2])
	patch, _ := strconv.Atoi(m[3])
	return &VersionString{
		baseVersion: baseVersion,
		isDevBuild:  isDevBuild,
		buildNumber: buildNumber,
		major:       major,
		minor:       minor,
		patch:       patch,
		suffix:      m[4],
	}, nil
}

func IsValidBaseVersion(baseVersion string) bool {
	return baseVersionOnlyPattern.MatchString(baseVersion)
}

func (v *VersionString) GetNumber() int {
	return v.major*1_000_000 + v.minor*1_000 + v.patch
}

func (v *VersionString) GetBaseVersion() string { return v.baseVersion }

func (v *VersionString) GetFullVersion(build bool) string {
	retval := v.baseVersion
	if v.isDevBuild {
		retval += "+dev"
		if build && v.buildNumber > 0 {
			retval += "." + strconv.Itoa(v.buildNumber)
		}
	}
	return retval
}

func (v *VersionString) GetMajor() int     { return v.major }
func (v *VersionString) GetMinor() int     { return v.minor }
func (v *VersionString) GetPatch() int     { return v.patch }
func (v *VersionString) GetSuffix() string { return v.suffix }
func (v *VersionString) GetBuild() int     { return v.buildNumber }
func (v *VersionString) IsDev() bool       { return v.isDevBuild }

// String replaces __toString().
func (v *VersionString) String() string {
	return v.GetFullVersion(false)
}

func signOfInt(n int) int {
	switch {
	case n > 0:
		return 1
	case n < 0:
		return -1
	default:
		return 0
	}
}

// Compare mirrors VersionString::compare(): returns -1/0/1 (or a raw difference if diff is
// true), ordering newer versions first, with dev builds and suffixed releases sorting as
// older than an equivalent release build.
func (v *VersionString) Compare(target *VersionString, diff bool) int {
	number := v.GetNumber()
	tNumber := target.GetNumber()
	if diff {
		return tNumber - number
	}

	if result := signOfInt(tNumber - number); result != 0 {
		return result
	}
	if target.IsDev() != v.IsDev() {
		if v.IsDev() {
			return 1 // Dev builds of the same version are always considered older than a release
		}
		return -1
	}
	if (target.GetSuffix() == "") != (v.suffix == "") {
		if v.suffix != "" {
			return 1 // alpha/beta/whatever releases are always considered older than a non-suffixed version
		}
		return -1
	}
	return signOfInt(target.GetBuild() - v.GetBuild())
}
