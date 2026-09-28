package bedrock

import "testing"

func TestLegacyBiomeIdToStringIdMapKnowsEveryClientBiome(t *testing.T) {
	m := GetLegacyBiomeIdToStringIdMap()
	if got := len(m.legacyToString); got != len(BiomeDefinitionList().BiomeDefinitions) {
		t.Errorf("mapped %d biomes, the client knows %d", got, len(BiomeDefinitionList().BiomeDefinitions))
	}
	for id, want := range map[int]string{0: "ocean", 1: "plains", 8: "hell"} {
		if got, ok := m.LegacyToString(id); !ok || got != want {
			t.Errorf("LegacyToString(%d) = %q, want %q", id, got, want)
		}
	}
	if _, ok := m.LegacyToString(250); ok {
		t.Error("biome 250 shouldn't exist")
	}
}
