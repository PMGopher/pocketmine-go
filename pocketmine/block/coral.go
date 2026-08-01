package block

import "pocketmine-go/pocketmine/math"

// Coral is a port of pocketmine\block\Coral.
type Coral struct {
	BaseCoral
}

func NewCoral(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Coral {
	c := &Coral{BaseCoral{Transparent: Transparent{NewBlock(idInfo, name, typeInfo)}}}
	c.Init(c)
	return c
}

func (c *Coral) Clone() Behavior {
	cl := *c
	cl.rebind(&cl)
	return &cl
}

func (c *Coral) canBeSupportedAt(blk Behavior) bool {
	return blk.(blockGeometry).GetAdjacentSupportType(math.Down).HasCenterSupport()
}

func (c *Coral) CanBePlacedAt(blockReplace Behavior, clickVector math.Vector3, face math.Facing, isClickedBlock bool) bool {
	return c.canBeSupportedAt(blockReplace) && c.Transparent.CanBePlacedAt(blockReplace, clickVector, face, isClickedBlock)
}

func (c *Coral) OnNearbyBlockChange() {
	if !c.canBeSupportedAt(c.self) {
		if world, err := c.position.GetWorld(); err == nil {
			world.UseBreakOn(c.position.AsVector3())
		}
	} else {
		c.BaseCoral.OnNearbyBlockChange()
	}
}
