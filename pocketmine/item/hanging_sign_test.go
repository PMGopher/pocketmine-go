package item

import (
	"testing"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/math"
)

var _ Item = (*HangingSign)(nil)

func newTestHangingSignVariants() (center, edge, wall block.Behavior) {
	typeInfo := block.NewBlockTypeInfo(block.BlockBreakInfoInstant(block.ToolTypeNone, 0), nil, nil)
	centerID, _ := block.NewBlockIdentifier(1, nil)
	edgeID, _ := block.NewBlockIdentifier(2, nil)
	wallID, _ := block.NewBlockIdentifier(3, nil)
	return block.NewDirt(centerID, "Center", typeInfo), block.NewDirt(edgeID, "Edge", typeInfo), block.NewDirt(wallID, "Wall", typeInfo)
}

func TestHangingSignGetBlockForFacePicksCenterOnlyForDown(t *testing.T) {
	center, edge, wall := newTestHangingSignVariants()
	h := NewHangingSign(NewItemIdentifier(0), "Hanging Sign", center, edge, wall)

	down := math.Down
	if got := h.GetBlockForFace(&down); got.GetName() != "Center" {
		t.Errorf("GetBlockForFace(Down) = %q, want Center", got.GetName())
	}

	up := math.Up
	if got := h.GetBlockForFace(&up); got.GetName() != "Wall" {
		t.Errorf("GetBlockForFace(Up) = %q, want Wall", got.GetName())
	}

	if got := h.GetBlockForFace(nil); got.GetName() != "Wall" {
		t.Errorf("GetBlockForFace(nil) = %q, want Wall", got.GetName())
	}
}

func TestHangingSignMaxStackSizeAndFuelTime(t *testing.T) {
	center, edge, wall := newTestHangingSignVariants()
	h := NewHangingSign(NewItemIdentifier(0), "Hanging Sign", center, edge, wall)

	if h.GetMaxStackSize() != 16 {
		t.Errorf("GetMaxStackSize() = %d, want 16", h.GetMaxStackSize())
	}
	if h.GetFuelTime() != 200 {
		t.Errorf("GetFuelTime() = %d, want 200", h.GetFuelTime())
	}
}
