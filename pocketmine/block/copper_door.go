package block

import (
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// CopperDoor is a port of pocketmine\block\CopperDoor.
type CopperDoor struct {
	Door
	CopperComponent
}

func NewCopperDoor(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *CopperDoor {
	c := &CopperDoor{
		Door: Door{
			Transparent:               Transparent{NewBlock(idInfo, name, typeInfo)},
			HorizontalFacingComponent: NewHorizontalFacingComponent(),
		},
	}
	c.Init(c)
	return c
}

func (c *CopperDoor) Clone() Behavior {
	cl := *c
	cl.rebind(&cl)
	return &cl
}

func (c *CopperDoor) DescribeBlockItemState(w runtime.DataDescriber) { c.DescribeCopper(w) }

func (c *CopperDoor) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	if player != nil && player.IsSneaking() && c.OnInteractCopper(c.self, c.position, item) {
		// copy copper properties to other half
		otherSide := math.Up
		if c.Top {
			otherSide = math.Down
		}
		if other, ok := c.GetSide(otherSide, 1).(*CopperDoor); ok {
			other.SetOxidation(c.Oxidation)
			other.SetWaxed(c.Waxed)
			if world, err := c.position.GetWorld(); err == nil {
				if err := world.SetBlock(other.GetPosition(), other); err != nil {
					panic(err)
				}
			}
		}
		return true
	}

	return c.Door.OnInteract(item, face, clickVector, player, returnedItems)
}
