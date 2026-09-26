package convert

import (
	"testing"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/data/bedrock"
)

func TestBlockTranslatorTranslatesVanillaBlocks(t *testing.T) {
	tr := NewBlockTranslator()
	for name, want := range map[string]string{
		"air":            "minecraft:air",
		"stone":          "minecraft:stone",
		"oak_planks":     "minecraft:oak_planks",
		"chest":          "minecraft:chest",
		"crafting_table": "minecraft:crafting_table",
	} {
		id := tr.InternalIDToNetworkID(block.VanillaBlock(name).GetStateId())
		if got := bedrock.BlockStates()[id].Name; got != want {
			t.Errorf("%s: network state %s, want %s", name, got, want)
		}
	}
}

func TestBlockTranslatorKeepsState(t *testing.T) {
	tr := NewBlockTranslator()
	b := block.VanillaBlock("bedrock").(*block.Bedrock)
	b.SetBurnsForever(true)
	data := bedrock.BlockStates()[tr.InternalIDToNetworkID(b.GetStateId())]
	if data.Name != "minecraft:bedrock" || data.States["infiniburn_bit"] != uint8(1) {
		t.Errorf("bedrock(burns forever) translated to %s %v", data.Name, data.States)
	}
}

func TestBlockTranslatorFallback(t *testing.T) {
	tr := NewBlockTranslator()
	if got := tr.InternalIDToNetworkID(1 << 30); got != tr.FallbackStateID() {
		t.Errorf("an unknown state translated to %d, want the info_update fallback %d", got, tr.FallbackStateID())
	}
}
