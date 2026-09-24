package bedrock

import "testing"

func TestBlockStatesLoadsTheFullCanonicalList(t *testing.T) {
	states := BlockStates()
	// the vendored Bedrock 1.26.50 list has exactly 22091 entries - a hard-coded
	// expectation is deliberate here: any change means the vendored asset changed, which should be
	// a conscious, visible event, not something a bug quietly slips past.
	if len(states) != 22091 {
		t.Fatalf("len(BlockStates()) = %d, want 22091", len(states))
	}
}

func TestRuntimeIDForKnownStatelessBlocks(t *testing.T) {
	cases := []struct {
		name string
		want int32
	}{
		{"minecraft:air", 17025},
		{"minecraft:stone", 3317},
		{"minecraft:dirt", 13456},
		{"minecraft:grass_block", 14944},
	}
	for _, c := range cases {
		got, ok := RuntimeIDFor(c.name, map[string]any{})
		if !ok {
			t.Errorf("RuntimeIDFor(%q) not found", c.name)
			continue
		}
		if got != c.want {
			t.Errorf("RuntimeIDFor(%q) = %d, want %d", c.name, got, c.want)
		}
	}
}

func TestRuntimeIDForBlockWithProperties(t *testing.T) {
	got, ok := RuntimeIDFor("minecraft:bedrock", map[string]any{"infiniburn_bit": uint8(0)})
	if !ok {
		t.Fatal("RuntimeIDFor(minecraft:bedrock, infiniburn_bit=0) not found")
	}
	if got != 17901 {
		t.Errorf("RuntimeIDFor(minecraft:bedrock, infiniburn_bit=0) = %d, want 17901", got)
	}
}

func TestRuntimeIDForReturnsFalseForUnknownState(t *testing.T) {
	if _, ok := RuntimeIDFor("minecraft:stone", map[string]any{"nonexistent_property": "x"}); ok {
		t.Error("expected RuntimeIDFor to report not-found for a nonexistent property set")
	}
	if _, ok := RuntimeIDFor("minecraft:this_block_does_not_exist", map[string]any{}); ok {
		t.Error("expected RuntimeIDFor to report not-found for an unknown block name")
	}
}

func TestBlockStatesRuntimeIDMatchesSliceIndex(t *testing.T) {
	states := BlockStates()
	got, ok := RuntimeIDFor("minecraft:grass_block", map[string]any{})
	if !ok {
		t.Fatal("RuntimeIDFor(minecraft:grass_block) not found")
	}
	if states[got].Name != "minecraft:grass_block" {
		t.Errorf("BlockStates()[%d].Name = %q, want minecraft:grass_block", got, states[got].Name)
	}
}
