package block

import (
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// CopperTrapdoor is a port of pocketmine\block\CopperTrapdoor.
type CopperTrapdoor struct {
	Trapdoor
	CopperComponent
}

func NewCopperTrapdoor(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *CopperTrapdoor {
	c := &CopperTrapdoor{
		Trapdoor: Trapdoor{
			Transparent:               Transparent{NewBlock(idInfo, name, typeInfo)},
			HorizontalFacingComponent: NewHorizontalFacingComponent(),
		},
	}
	c.Init(c)
	return c
}

func (c *CopperTrapdoor) Clone() Behavior {
	cl := *c
	cl.rebind(&cl)
	return &cl
}

func (c *CopperTrapdoor) DescribeBlockItemState(w runtime.DataDescriber) { c.DescribeCopper(w) }

func (c *CopperTrapdoor) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	if player != nil && player.IsSneaking() && c.OnInteractCopper(c.self, c.position, item) {
		return true
	}
	return c.Trapdoor.OnInteract(item, face, clickVector, player, returnedItems)
}
