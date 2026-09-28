package world

import (
	"testing"

	"pocketmine-go/pocketmine/block"
	_ "pocketmine-go/pocketmine/world/format/io"
)

func TestAdventureListAllows(t *testing.T) {
	stone := block.VanillaBlock("stone")
	if adventureListAllows(nil, stone) {
		t.Error("an empty list allowed a block")
	}
	if !adventureListAllows(map[string]string{"minecraft:stone": "minecraft:stone"}, stone) {
		t.Error("minecraft:stone didn't allow stone")
	}
	if adventureListAllows(map[string]string{"dirt": "dirt", "not an item": "not an item"}, stone) {
		t.Error("dirt allowed stone")
	}
}
