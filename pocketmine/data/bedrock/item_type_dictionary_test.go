package bedrock

import "testing"

func TestItemTypesLoadsARealisticNumberOfEntries(t *testing.T) {
	types := ItemTypes()
	if len(types) < 1000 {
		t.Fatalf("ItemTypes() returned %d entries, want at least 1000 (this is the full vendored Bedrock item table)", len(types))
	}
}

func TestItemRuntimeIDForKnownItem(t *testing.T) {
	id, ok := ItemRuntimeIDFor("minecraft:diamond")
	if !ok {
		t.Fatal("ItemRuntimeIDFor(\"minecraft:diamond\") ok = false, want true")
	}
	name, ok := ItemNameForRuntimeID(id)
	if !ok || name != "minecraft:diamond" {
		t.Errorf("ItemNameForRuntimeID(%d) = %q, %v; want \"minecraft:diamond\", true", id, name, ok)
	}
}

func TestItemRuntimeIDForUnknownItemReturnsFalse(t *testing.T) {
	if _, ok := ItemRuntimeIDFor("minecraft:this_does_not_exist"); ok {
		t.Error("ItemRuntimeIDFor(unknown) ok = true, want false")
	}
}

func TestItemTypesIncludesComponentBasedItemsWithDecodedNBT(t *testing.T) {
	types := ItemTypes()
	found := false
	for _, e := range types {
		if e.Name == "minecraft:apple" {
			found = true
			if !e.ComponentBased {
				t.Error("minecraft:apple ComponentBased = false, want true")
			}
			if len(e.Data) == 0 {
				t.Error("minecraft:apple Data is empty, want decoded component NBT")
			}
		}
	}
	if !found {
		t.Fatal("minecraft:apple not found in ItemTypes()")
	}
}

func TestItemTypesEveryEntryHasAUniqueRuntimeID(t *testing.T) {
	seen := map[int32]string{}
	for _, e := range ItemTypes() {
		if other, ok := seen[e.RuntimeID]; ok {
			t.Errorf("runtime ID %d used by both %q and %q", e.RuntimeID, other, e.Name)
		}
		seen[e.RuntimeID] = e.Name
	}
}
