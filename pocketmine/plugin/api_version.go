package plugin

import (
	"sort"

	"pocketmine-go/pocketmine/utils"
)

// parseAPIVersion parses a raw version string exactly like real PHP's `new VersionString($str)`
// does for API-compatibility purposes: utils.NewVersionString's own regex already extracts
// major/minor/patch/suffix from a bare "major.minor.patch[-suffix]" string, so isDevBuild/
// buildNumber (irrelevant to ApiVersion's comparisons) are always passed as false/0.
func parseAPIVersion(raw string) (*utils.VersionString, error) {
	return utils.NewVersionString(raw, false, 0)
}

// IsCompatible is a port of ApiVersion::isCompatible.
func IsCompatible(myVersionStr string, wantVersionsStr []string) bool {
	myVersion, err := parseAPIVersion(myVersionStr)
	if err != nil {
		return false
	}

	for _, versionStr := range wantVersionsStr {
		version, err := parseAPIVersion(versionStr)
		if err != nil {
			continue
		}

		if version.GetBaseVersion() != myVersion.GetBaseVersion() {
			if version.GetMajor() != myVersion.GetMajor() {
				continue
			}
			if version.GetMinor() > myVersion.GetMinor() {
				continue
			}
			if version.GetMinor() == myVersion.GetMinor() && version.GetPatch() > myVersion.GetPatch() {
				continue
			}
		}

		return true
	}

	return false
}

// CheckAmbiguousVersions is a port of ApiVersion::checkAmbiguousVersions.
func CheckAmbiguousVersions(versions []string) []string {
	indexed := map[int][]*utils.VersionString{}
	var majors []int
	for _, str := range versions {
		v, err := parseAPIVersion(str)
		if err != nil || v.GetSuffix() != "" { // suffix is always unambiguous
			continue
		}
		if _, ok := indexed[v.GetMajor()]; !ok {
			majors = append(majors, v.GetMajor())
		}
		indexed[v.GetMajor()] = append(indexed[v.GetMajor()], v)
	}

	var result []*utils.VersionString
	for _, major := range majors {
		list := indexed[major]
		if len(list) > 1 {
			result = append(result, list...)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Compare(result[j], false) < 0
	})

	out := make([]string, len(result))
	for i, v := range result {
		out[i] = v.GetBaseVersion()
	}
	return out
}
