package block

import (
	"testing"

	"pocketmine-go/pocketmine/math"
)

func newTestRespawnAnchor(w World) *RespawnAnchor {
	r := NewRespawnAnchor(mustBlockIdentifier(1025), "Test Respawn Anchor", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	r.SetPosition(w, 1, 2, 3)
	return r
}

func TestRespawnAnchorChargesWithGlowstoneUpToMax(t *testing.T) {
	w := &fakeWorld{}
	r := newTestRespawnAnchor(w)

	glowstone := fakeItem{typeID: itemTypeIDsGlowstone}
	for i := 1; i <= respawnAnchorMaxCharges; i++ {
		if !r.OnInteract(glowstone, 0, math.Vector3{}, nil, nil) {
			t.Fatalf("charge %d: expected OnInteract to handle glowstone", i)
		}
		if r.Charges != i {
			t.Errorf("charge %d: Charges = %d, want %d", i, r.Charges, i)
		}
	}

	// Already at max: further glowstone charging should be rejected.
	if r.OnInteract(glowstone, 0, math.Vector3{}, nil, nil) {
		t.Error("expected OnInteract to reject charging past max")
	}
}

func TestRespawnAnchorGetLightLevelScalesWithCharges(t *testing.T) {
	w := &fakeWorld{}
	r := newTestRespawnAnchor(w)

	if r.GetLightLevel() != 0 {
		t.Errorf("uncharged: GetLightLevel() = %d, want 0", r.GetLightLevel())
	}
	r.Charges = 4
	if want := 4*4 - 1; r.GetLightLevel() != want {
		t.Errorf("4 charges: GetLightLevel() = %d, want %d", r.GetLightLevel(), want)
	}
}
