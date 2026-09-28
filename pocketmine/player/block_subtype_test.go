package player

import (
	"testing"

	"pocketmine-go/pocketmine/block"
	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/math"
)

// These cover blocks whose Go type embeds the PHP base class (WoodenDoor embeds Door, ...): the
// base's `instanceof` checks must match them too.

func TestPlacingAndOpeningAWoodenDoor(t *testing.T) {
	p := newTestPlayer(t, 1, math.NewVector3(0.5, 70, 0.5))
	w := p.GetWorld()
	setBlockAt(t, p, 0, 69, 2, block.VanillaStone())
	setBlockAt(t, p, 0, 70, 2, block.VanillaAir())
	setBlockAt(t, p, 0, 71, 2, block.VanillaAir())

	p.GetInventory().SetItemInHand(parsed(t, "oak_door"))
	if !p.InteractBlock(math.NewVector3(0, 69, 2), math.Up, math.NewVector3(0.5, 1, 0.5)) {
		t.Fatal("placing the door failed")
	}
	bottom, ok := w.GetBlockAt(0, 70, 2).(*block.WoodenDoor)
	top, ok2 := w.GetBlockAt(0, 71, 2).(*block.WoodenDoor)
	if !ok || !ok2 || !top.Top {
		t.Fatalf("door halves = %T %T", w.GetBlockAt(0, 70, 2), w.GetBlockAt(0, 71, 2))
	}

	p.GetInventory().SetItemInHand(item.VanillaAir())
	bottom.OnInteract(item.VanillaAir(), math.North, math.NewVector3(0.5, 0.5, 0.5), p, nil)
	if !w.GetBlockAt(0, 70, 2).(*block.WoodenDoor).Open || !w.GetBlockAt(0, 71, 2).(*block.WoodenDoor).Open {
		t.Error("opening the bottom half didn't open both halves")
	}
}

func TestWoodenSlabsCombine(t *testing.T) {
	p := newTestPlayer(t, 1, math.NewVector3(0.5, 70, 0.5))
	w := p.GetWorld()
	setBlockAt(t, p, 0, 69, 2, block.VanillaStone())
	setBlockAt(t, p, 0, 70, 2, block.VanillaAir())

	p.GetInventory().SetItemInHand(parsed(t, "oak_slab"))
	p.InteractBlock(math.NewVector3(0, 69, 2), math.Up, math.NewVector3(0.5, 1, 0.5))
	p.GetInventory().SetItemInHand(parsed(t, "oak_slab"))
	p.InteractBlock(math.NewVector3(0, 70, 2), math.Up, math.NewVector3(0.5, 0.5, 0.5))
	slab, ok := w.GetBlockAt(0, 70, 2).(*block.WoodenSlab)
	if !ok || slab.SlabTypeValue != blockutils.SlabTypeDouble {
		t.Errorf("two oak slabs didn't make a double slab: %v", w.GetBlockAt(0, 70, 2))
	}
}

func TestPlacingTallGrass(t *testing.T) {
	p := newTestPlayer(t, 1, math.NewVector3(0.5, 70, 0.5))
	w := p.GetWorld()
	setBlockAt(t, p, 0, 69, 2, block.VanillaGrass())
	setBlockAt(t, p, 0, 70, 2, block.VanillaAir())
	setBlockAt(t, p, 0, 71, 2, block.VanillaAir())

	p.GetInventory().SetItemInHand(parsed(t, "double_tallgrass"))
	if !p.InteractBlock(math.NewVector3(0, 69, 2), math.Up, math.NewVector3(0.5, 1, 0.5)) {
		t.Fatal("placing double tall grass failed")
	}
	if _, ok := w.GetBlockAt(0, 71, 2).(*block.DoubleTallGrass); !ok {
		t.Errorf("top half = %T, want DoubleTallGrass", w.GetBlockAt(0, 71, 2))
	}
}

func parsed(t *testing.T, name string) item.Item {
	t.Helper()
	it, ok := item.GetStringToItemParser().Parse(name)
	if !ok {
		t.Fatalf("no item %q", name)
	}
	return it
}
