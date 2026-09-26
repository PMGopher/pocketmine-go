package item

import (
	"pocketmine-go/pocketmine/block"
	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/math"
	runtime "pocketmine-go/pocketmine/data/runtime"
)

// CoralFan is a port of pocketmine\item\CoralFan. It reuses block.CoralComponent directly for its
// coral-type/dead state (the same struct backing FloorCoralFan/WallCoralFan block state), matching
// PHP's own reuse of CoralTypeTrait for both Block and Item.
//
type CoralFan struct {
	ItemBase
	block.CoralComponent
}

func NewCoralFan(identifier ItemIdentifier, name string) *CoralFan {
	c := &CoralFan{}
	c.Init(c, identifier, name)
	return c
}

func (c *CoralFan) Clone() Item {
	cl := *c
	cl.rebind(&cl)
	return &cl
}

func (c *CoralFan) describeState(w runtime.DataDescriber) { c.DescribeCoral(w) }

// coralSetter is the part of FloorCoralFan/WallCoralFan GetBlockForFace sets.
type coralSetter interface {
	SetCoralType(coralType blockutils.CoralType)
	SetDead(dead bool)
}

// GetBlockForFace is a port of CoralFan::getBlock: a wall fan when clicking a side face.
func (c *CoralFan) GetBlockForFace(clickedFace *math.Facing) block.Behavior {
	var blk block.Behavior
	if clickedFace != nil && math.FacingAxis(*clickedFace) != math.AxisY {
		blk = block.VanillaBlock("wall_coral_fan")
	} else {
		blk = block.VanillaBlock("coral_fan")
	}
	setter := blk.(coralSetter)
	setter.SetCoralType(c.CoralType)
	setter.SetDead(c.Dead)
	return blk
}

func (c *CoralFan) GetBlock() block.Behavior { return c.GetBlockForFace(nil) }

func (c *CoralFan) GetFuelTime() int { return c.GetBlock().GetFuelTime() }

func (c *CoralFan) GetMaxStackSize() int { return c.GetBlock().GetMaxStackSize() }
