package utils

import (
	"slices"
	"testing"
)

func TestNatCompare(t *testing.T) {
	items := []string{"img12", "img10", "IMG2", "img1"}
	slices.SortFunc(items, NatCaseCompare)
	if want := []string{"img1", "IMG2", "img10", "img12"}; !slices.Equal(items, want) {
		t.Errorf("got %v, want %v", items, want)
	}
	if NatCompare("a2", "a10") >= 0 {
		t.Error("a2 should sort before a10")
	}
}
