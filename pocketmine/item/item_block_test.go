package item

import (
	"testing"

	"pocketmine-go/pocketmine/block"
	blockutils "pocketmine-go/pocketmine/block/utils"
)

var _ Item = (*ItemBlock)(nil)

func newTestDirtBlock() block.Behavior {
	idInfo, err := block.NewBlockIdentifier(3, nil)
	if err != nil {
		panic(err)
	}
	typeInfo := block.NewBlockTypeInfo(block.BlockBreakInfoInstant(block.ToolTypeShovel, 0), nil, nil)
	return block.NewDirt(idInfo, "Dirt", typeInfo)
}

func newTestAirBlock() block.Behavior {
	idInfo, err := block.NewBlockIdentifier(block.AIR, nil)
	if err != nil {
		panic(err)
	}
	typeInfo := block.NewBlockTypeInfo(block.BlockBreakInfoInstant(block.ToolTypeNone, 0), nil, nil)
	return block.NewAir(idInfo, "Air", typeInfo)
}

func TestItemBlockTakesNameFromBlock(t *testing.T) {
	ib := NewItemBlock(NewItemIdentifier(0), newTestDirtBlock())
	if ib.GetVanillaName() != "Dirt" {
		t.Errorf("GetVanillaName() = %q, want %q", ib.GetVanillaName(), "Dirt")
	}
}

func TestItemBlockGetBlockReturnsAClone(t *testing.T) {
	dirt := newTestDirtBlock().(*block.Dirt)
	dirt.DirtTypeValue = blockutils.DirtTypeCoarse

	ib := NewItemBlock(NewItemIdentifier(0), dirt)
	got := ib.GetBlock().(*block.Dirt)

	if got.DirtTypeValue != blockutils.DirtTypeCoarse {
		t.Errorf("GetBlock().DirtTypeValue = %v, want Coarse", got.DirtTypeValue)
	}

	got.DirtTypeValue = blockutils.DirtTypeRooted
	if dirt.DirtTypeValue != blockutils.DirtTypeCoarse {
		t.Error("expected mutating the block returned by GetBlock() not to affect the original")
	}
}

func TestItemBlockIsNullWhenWrappingAir(t *testing.T) {
	ib := NewItemBlock(NewItemIdentifier(0), newTestAirBlock())
	if !ib.IsNull() {
		t.Error("expected an ItemBlock wrapping Air to be null")
	}
}

func TestItemBlockIsNotNullWhenWrappingDirt(t *testing.T) {
	ib := NewItemBlock(NewItemIdentifier(0), newTestDirtBlock())
	if ib.IsNull() {
		t.Error("expected an ItemBlock wrapping Dirt not to be null")
	}
}

func TestItemBlockCloneDeepCopiesWrappedBlock(t *testing.T) {
	dirt := newTestDirtBlock().(*block.Dirt)
	ib := NewItemBlock(NewItemIdentifier(0), dirt)

	clone := ib.Clone().(*ItemBlock)
	clone.Block.(*block.Dirt).DirtTypeValue = blockutils.DirtTypeCoarse

	if dirt.DirtTypeValue != blockutils.DirtTypeNormal {
		t.Error("expected cloning the ItemBlock not to affect the original wrapped block")
	}
}

func TestItemBlockSatisfiesItemBlockLikeMarker(t *testing.T) {
	ib := NewItemBlock(NewItemIdentifier(0), newTestDirtBlock())
	if _, ok := interface{}(ib).(block.ItemBlockLike); !ok {
		t.Error("expected ItemBlock to satisfy block.ItemBlockLike")
	}
}
