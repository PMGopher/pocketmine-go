package bedrock

import (
	"bytes"
	"testing"

	gtnbt "github.com/sandertv/gophertunnel/minecraft/nbt"
)

func TestBiomeDefinitionListDecodes(t *testing.T) {
	pk := BiomeDefinitionList()
	if len(pk.BiomeDefinitions) == 0 || len(pk.StringList) == 0 {
		t.Fatalf("BiomeDefinitionList has %d definitions and %d strings, want both non-empty", len(pk.BiomeDefinitions), len(pk.StringList))
	}
	names := map[string]bool{}
	for _, def := range pk.BiomeDefinitions {
		if int(def.NameIndex) >= len(pk.StringList) {
			t.Fatalf("biome name index %d out of range (%d strings)", def.NameIndex, len(pk.StringList))
		}
		names[pk.StringList[def.NameIndex]] = true
	}
	for _, want := range []string{"plains", "ocean", "desert", "forest"} {
		if !names[want] {
			t.Errorf("biome %q missing from the definitions", want)
		}
	}
}

func TestEntityIdentifiersAreNetworkNBT(t *testing.T) {
	var m struct {
		IDList []map[string]any `nbt:"idlist"`
	}
	dec := gtnbt.NewDecoderWithEncoding(bytes.NewReader(AvailableActorIdentifiers().SerialisedEntityIdentifiers), gtnbt.NetworkLittleEndian)
	if err := dec.Decode(&m); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, e := range m.IDList {
		if e["id"] == "minecraft:zombie" {
			found = true
		}
	}
	if len(m.IDList) < 100 || !found {
		t.Errorf("idlist has %d entries (zombie present: %v), want the full vanilla list", len(m.IDList), found)
	}
}
