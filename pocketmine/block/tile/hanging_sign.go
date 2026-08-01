package tile

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/math"
)

// HangingSign is a port of pocketmine\block\tile\HangingSign.
//
// Deprecated in the PHP original too.
type HangingSign struct {
	Sign
}

func NewHangingSign(world World, pos math.Vector3) *HangingSign {
	h := &HangingSign{Sign: Sign{
		SpawnableBase: SpawnableBase{TileBase: NewTileBase(world, pos)},
		Text:          blockutils.NewSignText(nil, nil, false),
		BackText:      blockutils.NewSignText(nil, nil, false),
	}}
	h.Init(h)
	return h
}

func (h *HangingSign) SaveID() string { return "HangingSign" }
