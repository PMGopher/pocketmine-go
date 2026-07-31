package block

import (
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// CopperLantern is a port of pocketmine\block\CopperLantern.
type CopperLantern struct {
	Lantern
	CopperComponent
}

func NewCopperLantern(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo, lightLevel int) *CopperLantern {
	c := &CopperLantern{
		Lantern: Lantern{
			Transparent:     Transparent{NewBlock(idInfo, name, typeInfo)},
			LightLevelValue: lightLevel,
		},
	}
	c.Init(c)
	return c
}

func (c *CopperLantern) Clone() Behavior {
	cl := *c
	cl.rebind(&cl)
	return &cl
}

func (c *CopperLantern) DescribeBlockItemState(w runtime.DataDescriber) { c.DescribeCopper(w) }

func (c *CopperLantern) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	return c.OnInteractCopper(c.self, c.position, item)
}
