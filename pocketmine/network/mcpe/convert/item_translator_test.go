package convert

import (
	"testing"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/data/bedrock"
	"pocketmine-go/pocketmine/item"
)

func TestToNetworkIDForAKnownItem(t *testing.T) {
	tr := NewItemTranslator()
	networkID, meta, blockRuntimeID, err := tr.ToNetworkID(item.VanillaApple())
	if err != nil {
		t.Fatal(err)
	}
	wantID, _ := bedrock.ItemRuntimeIDFor("minecraft:apple")
	if networkID != wantID || meta != 0 || blockRuntimeID != noBlockRuntimeID {
		t.Errorf("ToNetworkID(apple) = %d, %d, %d, want %d, 0, %d", networkID, meta, blockRuntimeID, wantID, noBlockRuntimeID)
	}
}

// TestNetworkRoundTrip sends every vanilla item and every block item through ToNetworkID and back
// through FromNetworkID.
func TestNetworkRoundTrip(t *testing.T) {
	tr := NewItemTranslator()
	check := func(what string, it item.Item) {
		if it.GetTypeId() == item.OMINOUS_BANNER {
			//its only difference from a banner is a Type tag in the saved item data, which the network
			//item ID doesn't carry (same as PocketMine-MP)
			return
		}
		networkID, meta, blockRuntimeID, err := tr.ToNetworkID(it)
		if err != nil {
			t.Errorf("%s: ToNetworkID: %v", what, err)
			return
		}
		back, err := tr.FromNetworkID(networkID, meta, blockRuntimeID)
		if err != nil {
			t.Errorf("%s: FromNetworkID: %v", what, err)
			return
		}
		if back.GetStateId() != it.GetStateId() {
			t.Errorf("%s: round trip gave %s", what, back)
		}
	}
	for _, name := range item.GetVanillaItemNames() {
		if name == "air" {
			continue
		}
		check(name, item.VanillaItem(name))
	}
	for _, name := range block.GetVanillaBlockNames() {
		if name == "air" {
			continue
		}
		it, err := block.VanillaBlock(name).(interface{ AsItem() (block.Item, error) }).AsItem()
		if err != nil {
			t.Fatal(err)
		}
		check("block "+name, it.(item.Item))
	}
}
