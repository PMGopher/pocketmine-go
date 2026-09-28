package io

import (
	"testing"

	"pocketmine-go/pocketmine/data/bedrock"
)

// Stairs and panes saved by 1.26.30 (states stamped 1.26.30.0, without 1.26.50's minecraft:corner
// and minecraft:connection_* properties) must upgrade and deserialize to real blocks instead of
// the "update!" block.
func TestPre12650StatesUpgradeToRealBlocks(t *testing.T) {
	const v12630 = 1<<24 | 26<<16 | 30<<8
	for _, data := range []bedrock.BlockStateData{
		{Name: "minecraft:stone_brick_stairs", Version: v12630, States: map[string]any{"upside_down_bit": uint8(0), "weirdo_direction": int32(2)}},
		{Name: "minecraft:oak_stairs", Version: v12630, States: map[string]any{"upside_down_bit": uint8(0), "weirdo_direction": int32(0)}},
		{Name: "minecraft:glass_pane", Version: v12630, States: map[string]any{}},
		{Name: "minecraft:oak_stairs", Version: bedrock.CurrentBlockStateVersion, States: map[string]any{"upside_down_bit": uint8(0), "weirdo_direction": int32(1)}},
	} {
		upgraded := GetBlockDataUpgrader().GetBlockStateUpgrader().Upgrade(data)
		if _, err := GetBlockStateDeserializer().Deserialize(upgraded); err != nil {
			t.Errorf("%s %v: %v (upgraded states %v)", data.Name, data.States, err, upgraded.States)
		}
	}
}
