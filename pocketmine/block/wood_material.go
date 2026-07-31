package block

import blockutils "pocketmine-go/pocketmine/block/utils"

// WoodMaterial is a port of pocketmine\block\utils\WoodMaterial.
type WoodMaterial interface {
	GetWoodType() blockutils.WoodType
}

// WoodTypeComponent is a port of pocketmine\block\utils\WoodTypeTrait. The wood type is immutable
// once constructed (matching the PHP original's comment) and isn't part of block state — each
// wood variant of a block is its own distinct registered block type, not a state-bit variation —
// so there's no DescribeXState method here, just the getter.
type WoodTypeComponent struct {
	WoodType blockutils.WoodType
}

func NewWoodTypeComponent(woodType blockutils.WoodType) WoodTypeComponent {
	return WoodTypeComponent{WoodType: woodType}
}

func (w *WoodTypeComponent) GetWoodType() blockutils.WoodType { return w.WoodType }
