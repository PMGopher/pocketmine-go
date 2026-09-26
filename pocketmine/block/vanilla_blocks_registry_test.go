package block

import "testing"

// TestVanillaBlocksRegistry checks that every vanilla block registers (VanillaBlocksInputs::setup)
// with a unique type ID matching its BlockTypeIds constant.
func TestVanillaBlocksRegistry(t *testing.T) {
	names := GetVanillaBlockNames()
	if len(names) != 799 {
		t.Errorf("%d vanilla blocks registered, want 799 (every BlockTypeIds constant except CHERRY_SAPLING and POWDER_SNOW_CAULDRON, which PocketMine-MP 5.44.4 doesn't register either)", len(names))
	}
	seen := map[int]string{}
	for _, name := range names {
		blk := VanillaBlock(name)
		if other, ok := seen[blk.GetTypeId()]; ok {
			t.Errorf("%s and %s share type ID %d", name, other, blk.GetTypeId())
		}
		seen[blk.GetTypeId()] = name
		if blk.GetTypeId() >= FIRST_UNUSED_BLOCK_ID {
			t.Errorf("%s has no BlockTypeIds constant", name)
		}
	}
}

func TestRuntimeBlockStateRegistry(t *testing.T) {
	r := GetRuntimeBlockStateRegistry()
	states := r.GetAllKnownStates()
	t.Logf("%d block states", len(states))
	for _, state := range states {
		if got := r.FromStateId(state.GetStateId()); got.GetStateId() != state.GetStateId() {
			t.Fatalf("FromStateId(%d) returned state %d", state.GetStateId(), got.GetStateId())
		}
	}
	if !r.HasStateId(VanillaBlock("stone").GetStateId()) {
		t.Error("stone isn't registered")
	}
}
