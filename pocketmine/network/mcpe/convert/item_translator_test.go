package convert

import (
	"testing"

	"pocketmine-go/pocketmine/data/bedrock"
	"pocketmine-go/pocketmine/item"
)

// TestItemTypeNamesAllResolveToRealBedrockItems is the real correctness check for itemTypeNames:
// every single mapped string must exist in the actual vendored Bedrock item table (not just be
// syntactically plausible) - this is what catches a typo'd or guessed name immediately instead of
// silently sending the client a bogus item ID.
func TestItemTypeNamesAllResolveToRealBedrockItems(t *testing.T) {
	for typeID, name := range itemTypeNames {
		if _, ok := bedrock.ItemRuntimeIDFor(name); !ok {
			t.Errorf("itemTypeNames[%d] = %q, which does not exist in the vendored Bedrock item table", typeID, name)
		}
	}
}

func TestItemTypeNamesCoversAtLeastTheDocumentedCount(t *testing.T) {
	// 74 non-record simple items (Clownfish deliberately unmapped) + 20 records.
	if len(itemTypeNames) < 94 {
		t.Errorf("itemTypeNames has %d entries, want at least 94", len(itemTypeNames))
	}
}

func TestToNetworkIDForAKnownItem(t *testing.T) {
	tr := NewItemTranslator()
	networkID, meta, blockRuntimeID, ok := tr.ToNetworkID(item.VanillaApple())
	if !ok {
		t.Fatal("ToNetworkID(apple) ok = false, want true")
	}
	wantID, _ := bedrock.ItemRuntimeIDFor("minecraft:apple")
	if networkID != wantID {
		t.Errorf("ToNetworkID(apple) networkID = %d, want %d", networkID, wantID)
	}
	if meta != 0 {
		t.Errorf("ToNetworkID(apple) meta = %d, want 0", meta)
	}
	if blockRuntimeID != noBlockRuntimeID {
		t.Errorf("ToNetworkID(apple) blockRuntimeID = %d, want %d", blockRuntimeID, noBlockRuntimeID)
	}
}

func TestToNetworkIDForAnUnmappedItemReturnsFalse(t *testing.T) {
	tr := NewItemTranslator()
	if _, _, _, ok := tr.ToNetworkID(item.VanillaClownfish()); ok {
		t.Error("ToNetworkID(clownfish) ok = true, want false (no real Bedrock network item exists)")
	}
}
