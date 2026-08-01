package item

import "testing"

// These types are all pure ItemBlock-style leaves in the PHP original (their only content is a
// GetBlock() override, which needs the unported block registry - see StringItem's doc comment),
// so this just confirms each constructs correctly and carries the right type ID and count/name
// defaults through the shared ItemBase machinery.
func TestItemBlockOnlyLeaves(t *testing.T) {
	cases := []struct {
		name string
		item Item
		want int
	}{
		{"Redstone", NewRedstone(NewItemIdentifier(REDSTONE_DUST), "Redstone"), REDSTONE_DUST},
		{"CocoaBeans", NewCocoaBeans(NewItemIdentifier(COCOA_BEANS), "Cocoa Beans"), COCOA_BEANS},
		{"MelonSeeds", NewMelonSeeds(NewItemIdentifier(MELON_SEEDS), "Melon Seeds"), MELON_SEEDS},
		{"PumpkinSeeds", NewPumpkinSeeds(NewItemIdentifier(PUMPKIN_SEEDS), "Pumpkin Seeds"), PUMPKIN_SEEDS},
		{"WheatSeeds", NewWheatSeeds(NewItemIdentifier(WHEAT_SEEDS), "Wheat Seeds"), WHEAT_SEEDS},
		{"BeetrootSeeds", NewBeetrootSeeds(NewItemIdentifier(BEETROOT_SEEDS), "Beetroot Seeds"), BEETROOT_SEEDS},
		{"PitcherPod", NewPitcherPod(NewItemIdentifier(PITCHER_POD), "Pitcher Pod"), PITCHER_POD},
		{"TorchflowerSeeds", NewTorchflowerSeeds(NewItemIdentifier(TORCHFLOWER_SEEDS), "Torchflower Seeds"), TORCHFLOWER_SEEDS},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.item.GetTypeId(); got != c.want {
				t.Errorf("GetTypeId() = %d, want %d", got, c.want)
			}
			if c.item.GetCount() != 1 {
				t.Errorf("GetCount() = %d, want 1 (default)", c.item.GetCount())
			}
			clone := c.item.Clone()
			clone.SetCount(5)
			if c.item.GetCount() != 1 {
				t.Error("expected Clone() mutation not to affect the original")
			}
		})
	}
}
