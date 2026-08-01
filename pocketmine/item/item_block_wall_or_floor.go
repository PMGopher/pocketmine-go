package item

import (
	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/math"
)

// ItemBlockWallOrFloor is a port of pocketmine\item\ItemBlockWallOrFloor: items (like signs and
// banners) that place a different block variant depending on whether the clicked face is
// vertical (floor) or horizontal (wall).
//
// RuntimeBlockStateRegistry::fromStateId (not ported - same block-registry gap as
// ItemBlock.GetBlock's doc comment) isn't used here - NewItemBlockWallOrFloor takes both block
// variants directly instead of round-tripping them through state IDs.
type ItemBlockWallOrFloor struct {
	ItemBase

	FloorVariant block.Behavior
	WallVariant  block.Behavior
}

func NewItemBlockWallOrFloor(identifier ItemIdentifier, floorVariant, wallVariant block.Behavior) *ItemBlockWallOrFloor {
	i := &ItemBlockWallOrFloor{FloorVariant: floorVariant, WallVariant: wallVariant}
	i.Init(i, identifier, floorVariant.GetName())
	return i
}

func (i *ItemBlockWallOrFloor) Clone() Item {
	c := *i
	c.FloorVariant = i.FloorVariant.Clone()
	c.WallVariant = i.WallVariant.Clone()
	c.rebind(&c)
	return &c
}

// GetBlockForFace is a port of ItemBlockWallOrFloor::getBlock. Named differently from the rest of
// this port's GetBlock() (no-argument) convention because this type's variant genuinely depends
// on the clicked face - callers that don't have a face should use GetBlock instead.
func (i *ItemBlockWallOrFloor) GetBlockForFace(clickedFace *math.Facing) block.Behavior {
	if clickedFace != nil && math.FacingAxis(*clickedFace) != math.AxisY {
		return i.WallVariant.Clone()
	}
	return i.FloorVariant.Clone()
}

func (i *ItemBlockWallOrFloor) GetBlock() block.Behavior { return i.FloorVariant.Clone() }

func (i *ItemBlockWallOrFloor) GetFuelTime() int { return i.FloorVariant.GetFuelTime() }

func (i *ItemBlockWallOrFloor) GetMaxStackSize() int { return i.FloorVariant.GetMaxStackSize() }
