package block

import runtime "pocketmine-go/pocketmine/data/runtime"

// DyedShulkerBox is a port of pocketmine\block\DyedShulkerBox.
type DyedShulkerBox struct {
	ShulkerBox
	ColorComponent
}

func NewDyedShulkerBox(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *DyedShulkerBox {
	d := &DyedShulkerBox{
		ShulkerBox: ShulkerBox{
			Opaque:          Opaque{NewBlock(idInfo, name, typeInfo)},
			FacingComponent: NewFacingComponent(),
		},
		ColorComponent: NewColorComponent(),
	}
	d.Init(d)
	return d
}

func (d *DyedShulkerBox) Clone() Behavior {
	c := *d
	c.rebind(&c)
	return &c
}

func (d *DyedShulkerBox) DescribeBlockItemState(w runtime.DataDescriber) { d.DescribeColor(w) }
