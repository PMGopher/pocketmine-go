package block

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// CopperBulb is a port of pocketmine\block\CopperBulb.
type CopperBulb struct {
	Opaque
	CopperComponent
	PoweredByRedstoneComponent
	LightableComponent
}

func NewCopperBulb(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *CopperBulb {
	c := &CopperBulb{Opaque: Opaque{NewBlock(idInfo, name, typeInfo)}}
	c.Init(c)
	return c
}

func (c *CopperBulb) Clone() Behavior {
	cl := *c
	cl.rebind(&cl)
	return &cl
}

func (c *CopperBulb) DescribeBlockOnlyState(w runtime.DataDescriber) {
	w.Bool(&c.Lit)
	w.Bool(&c.Powered)
}

func (c *CopperBulb) DescribeBlockItemState(w runtime.DataDescriber) { c.DescribeCopper(w) }

func (c *CopperBulb) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	return c.OnInteractCopper(c.self, c.position, item)
}

// TogglePowered is a port of CopperBulb::togglePowered.
func (c *CopperBulb) TogglePowered(powered bool) {
	if powered == c.Powered {
		return
	}
	if powered {
		c.Lit = !c.Lit
	}
	c.Powered = powered
}

func (c *CopperBulb) GetLightLevel() int {
	if !c.Lit {
		return 0
	}
	switch c.Oxidation {
	case blockutils.CopperOxidationNone:
		return 15
	case blockutils.CopperOxidationExposed:
		return 12
	case blockutils.CopperOxidationWeathered:
		return 8
	case blockutils.CopperOxidationOxidized:
		return 4
	}
	return 0
}
