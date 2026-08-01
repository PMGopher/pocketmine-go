package item

import (
	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/math"
)

// HangingSign is a port of pocketmine\item\HangingSign. GetPlacementTransaction (choosing between
// the wall/edge-ceiling/center-ceiling ready block variants based on clicked face and sneak state)
// isn't ported - it needs world.BlockTransaction and tryPlacementTransaction, neither of which
// exist in this port yet (see Item's doc comment for the same category of gap).
type HangingSign struct {
	ItemBase

	CenterPointCeilingVariant block.Behavior
	EdgePointCeilingVariant   block.Behavior
	WallVariant               block.Behavior
}

func NewHangingSign(identifier ItemIdentifier, name string, centerPointCeilingVariant, edgePointCeilingVariant, wallVariant block.Behavior) *HangingSign {
	h := &HangingSign{
		CenterPointCeilingVariant: centerPointCeilingVariant,
		EdgePointCeilingVariant:   edgePointCeilingVariant,
		WallVariant:               wallVariant,
	}
	h.Init(h, identifier, name)
	return h
}

func (h *HangingSign) Clone() Item {
	c := *h
	c.CenterPointCeilingVariant = h.CenterPointCeilingVariant.Clone()
	c.EdgePointCeilingVariant = h.EdgePointCeilingVariant.Clone()
	c.WallVariant = h.WallVariant.Clone()
	c.rebind(&c)
	return &c
}

// GetBlockForFace is a port of HangingSign::getBlock - "we don't have enough information here to
// decide which ceiling type to use" (the PHP original's own comment), so Facing::DOWN always
// picks the center-point ceiling variant.
func (h *HangingSign) GetBlockForFace(clickedFace *math.Facing) block.Behavior {
	if clickedFace != nil && *clickedFace == math.Down {
		return h.CenterPointCeilingVariant.Clone()
	}
	return h.WallVariant.Clone()
}

func (h *HangingSign) GetBlock() block.Behavior { return h.WallVariant.Clone() }

func (h *HangingSign) GetMaxStackSize() int { return 16 }

func (h *HangingSign) GetFuelTime() int { return 200 }
