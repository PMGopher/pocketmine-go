package item

import (
	"testing"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/math"
)

var _ Item = (*ItemBlockWallOrFloor)(nil)

func newTestWallSignLikeBlocks() (floor, wall block.Behavior) {
	floorID, err := block.NewBlockIdentifier(1, nil)
	if err != nil {
		panic(err)
	}
	wallID, err := block.NewBlockIdentifier(2, nil)
	if err != nil {
		panic(err)
	}
	typeInfo := block.NewBlockTypeInfo(block.BlockBreakInfoInstant(block.ToolTypeNone, 0), nil, nil)
	return block.NewDirt(floorID, "Floor Variant", typeInfo), block.NewDirt(wallID, "Wall Variant", typeInfo)
}

func TestItemBlockWallOrFloorTakesNameFromFloorVariant(t *testing.T) {
	floor, wall := newTestWallSignLikeBlocks()
	i := NewItemBlockWallOrFloor(NewItemIdentifier(0), floor, wall)
	if i.GetVanillaName() != "Floor Variant" {
		t.Errorf("GetVanillaName() = %q, want %q", i.GetVanillaName(), "Floor Variant")
	}
}

func TestItemBlockWallOrFloorGetBlockForFacePicksVariantByAxis(t *testing.T) {
	floor, wall := newTestWallSignLikeBlocks()
	i := NewItemBlockWallOrFloor(NewItemIdentifier(0), floor, wall)

	up := math.Up
	if got := i.GetBlockForFace(&up); got.GetName() != "Floor Variant" {
		t.Errorf("GetBlockForFace(Up) = %q, want Floor Variant", got.GetName())
	}

	north := math.North
	if got := i.GetBlockForFace(&north); got.GetName() != "Wall Variant" {
		t.Errorf("GetBlockForFace(North) = %q, want Wall Variant", got.GetName())
	}

	if got := i.GetBlockForFace(nil); got.GetName() != "Floor Variant" {
		t.Errorf("GetBlockForFace(nil) = %q, want Floor Variant", got.GetName())
	}
}

func TestItemBlockWallOrFloorCloneDeepCopiesBothVariants(t *testing.T) {
	floor, wall := newTestWallSignLikeBlocks()
	i := NewItemBlockWallOrFloor(NewItemIdentifier(0), floor, wall)

	clone := i.Clone().(*ItemBlockWallOrFloor)
	if clone.FloorVariant == i.FloorVariant || clone.WallVariant == i.WallVariant {
		t.Error("expected Clone to deep-copy both block variants, not share pointers")
	}
}
