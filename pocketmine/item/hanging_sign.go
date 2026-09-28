package item

import (
	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/math"
)

// HangingSign is a port of pocketmine\item\HangingSign.
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

// GetPlacementTransaction is a port of HangingSign::getPlacementTransaction.
func (h *HangingSign) GetPlacementTransaction(blockReplace, blockClicked block.Behavior, face math.Facing, clickVector math.Vector3, player Player) *block.BlockTransactionImpl {
	if face != math.Down {
		return TryPlacementTransaction(h, h.WallVariant.Clone(), blockReplace, blockClicked, face, clickVector, player)
	}
	//ceiling edges sign has stricter placement conditions than ceiling center sign, so try that first
	var ceilingEdgeTx *block.BlockTransactionImpl
	if player == nil || !player.IsSneaking() {
		ceilingEdgeTx = TryPlacementTransaction(h, h.EdgePointCeilingVariant.Clone(), blockReplace, blockClicked, face, clickVector, player)
	}
	if ceilingEdgeTx != nil {
		return ceilingEdgeTx
	}
	return TryPlacementTransaction(h, h.CenterPointCeilingVariant.Clone(), blockReplace, blockClicked, face, clickVector, player)
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
