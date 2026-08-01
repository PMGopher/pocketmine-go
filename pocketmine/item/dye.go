package item

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
)

// Dye is a port of pocketmine\item\Dye. Its GetColor() method is what satisfies block.Dye (the
// forward-compatible marker declared in block/base_sign.go), so BaseSign's dye-recoloring
// OnInteract branch now works with a real Dye instance instead of only a hypothetical one.
type Dye struct {
	ItemBase

	Color blockutils.DyeColor
}

func NewDye(identifier ItemIdentifier, name string) *Dye {
	d := &Dye{Color: blockutils.DyeColorBlack}
	d.Init(d, identifier, name)
	return d
}

func (d *Dye) Clone() Item {
	c := *d
	c.rebind(&c)
	return &c
}

func (d *Dye) GetColor() blockutils.DyeColor { return d.Color }

func (d *Dye) SetColor(color blockutils.DyeColor) { d.Color = color }

func (d *Dye) describeState(w runtime.DataDescriber) {
	col := int(d.Color)
	w.BoundedIntAuto(int(blockutils.DyeColorWhite), int(blockutils.DyeColorBlack), &col)
	d.Color = blockutils.DyeColor(col)
}
