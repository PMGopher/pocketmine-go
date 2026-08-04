package plugin

import "testing"

func TestIsCompatibleExactMatch(t *testing.T) {
	if !IsCompatible("5.0.0", []string{"5.0.0"}) {
		t.Error("IsCompatible(5.0.0, [5.0.0]) = false, want true")
	}
}

func TestIsCompatiblePluginRequestsOlderMinorIsBackwardsCompatible(t *testing.T) {
	// Engine at 5.2.0 should still run a plugin that only asked for 5.0.0 (backwards-compatible).
	if !IsCompatible("5.2.0", []string{"5.0.0"}) {
		t.Error("IsCompatible(5.2.0, [5.0.0]) = false, want true")
	}
}

func TestIsCompatiblePluginRequestsNewerMinorIsIncompatible(t *testing.T) {
	// Plugin wants API features from 5.2.0, engine is only at 5.0.0.
	if IsCompatible("5.0.0", []string{"5.2.0"}) {
		t.Error("IsCompatible(5.0.0, [5.2.0]) = true, want false")
	}
}

func TestIsCompatibleDifferentMajorIsIncompatible(t *testing.T) {
	if IsCompatible("5.0.0", []string{"4.0.0"}) {
		t.Error("IsCompatible(5.0.0, [4.0.0]) = true, want false")
	}
}

func TestIsCompatibleAnyOfMultipleWantedVersionsMatching(t *testing.T) {
	if !IsCompatible("5.0.0", []string{"4.0.0", "5.0.0"}) {
		t.Error("IsCompatible(5.0.0, [4.0.0, 5.0.0]) = false, want true")
	}
}

func TestCheckAmbiguousVersionsFlagsSameMajorWithoutSuffix(t *testing.T) {
	got := CheckAmbiguousVersions([]string{"5.0.0", "5.1.0", "4.0.0"})
	if len(got) != 2 {
		t.Fatalf("CheckAmbiguousVersions returned %v, want 2 ambiguous major-5 entries", got)
	}
}

func TestCheckAmbiguousVersionsIgnoresSuffixedVersions(t *testing.T) {
	got := CheckAmbiguousVersions([]string{"5.0.0-alpha1", "5.1.0-alpha2"})
	if len(got) != 0 {
		t.Errorf("CheckAmbiguousVersions with only suffixed versions = %v, want empty (suffix is always unambiguous)", got)
	}
}
