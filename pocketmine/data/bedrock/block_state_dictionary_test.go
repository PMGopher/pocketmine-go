package bedrock

import "testing"

func TestBlockStatesLoadsTheFullCanonicalList(t *testing.T) {
	states := BlockStates()
	// pmmp/BedrockData tag 6.7.0+bedrock-1.26.30 has exactly 16913 entries - a hard-coded
	// expectation is deliberate here: any change means the vendored asset changed, which should be
	// a conscious, visible event, not something a bug quietly slips past.
	if len(states) != 16913 {
		t.Fatalf("len(BlockStates()) = %d, want 16913", len(states))
	}
}

func TestRuntimeIDForKnownStatelessBlocks(t *testing.T) {
	cases := []struct {
		name string
		want int32
	}{
		{"minecraft:air", 13094},
		{"minecraft:stone", 2706},
		{"minecraft:dirt", 10392},
		{"minecraft:grass_block", 11608},
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
	if got != 13805 {
		t.Errorf("RuntimeIDFor(minecraft:bedrock, infiniburn_bit=0) = %d, want 13805", got)
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
