package item

import (
	"pocketmine-go/pocketmine/block"
	runtime "pocketmine-go/pocketmine/data/runtime"
)

// CoralFan is a port of pocketmine\item\CoralFan. It reuses block.CoralComponent directly for its
// coral-type/dead state (the same struct backing FloorCoralFan/WallCoralFan block state), matching
// PHP's own reuse of CoralTypeTrait for both Block and Item.
//
// GetBlock (needs VanillaBlocks.CORAL_FAN()/WALL_CORAL_FAN(), the unported block registry) isn't
// ported - GetBlock isn't part of Item here at all yet (see the Item interface's doc comment). Its
// PHP GetFuelTime/GetMaxStackSize both delegate to GetBlock().GetFuelTime()/GetMaxStackSize(), but
// neither FloorCoralFan nor WallCoralFan override those in this port (confirmed by grep), so
// ItemBase's own defaults (0 and 64) already produce the same effective values without needing
// GetBlock at all.
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
