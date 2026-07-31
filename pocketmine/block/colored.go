package block

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
)

// Colored is a port of pocketmine\block\utils\Colored.
type Colored interface {
	GetColor() blockutils.DyeColor
	SetColor(color blockutils.DyeColor)
}

// ColorComponent is a port of pocketmine\block\utils\ColoredTrait. Note the PHP trait's
// describeBlockItemState puts color in block-ITEM state (not block-only state) — the dye color
// survives AsItem(), unlike facing/age/etc.
type ColorComponent struct {
	Color blockutils.DyeColor
}

func NewColorComponent() ColorComponent { return ColorComponent{Color: blockutils.DyeColorWhite} }

func (c *ColorComponent) DescribeColor(w runtime.DataDescriber) {
	col := int(c.Color)
	w.BoundedIntAuto(int(blockutils.DyeColorWhite), int(blockutils.DyeColorBlack), &col)
	c.Color = blockutils.DyeColor(col)
}

func (c *ColorComponent) GetColor() blockutils.DyeColor { return c.Color }

func (c *ColorComponent) SetColor(color blockutils.DyeColor) { c.Color = color }
