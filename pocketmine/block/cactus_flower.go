package block

import "pocketmine-go/pocketmine/math"

// CactusFlower is a port of pocketmine\block\CactusFlower.
type CactusFlower struct {
	Flowable
}

func NewCactusFlower(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *CactusFlower {
	c := &CactusFlower{Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}}}
	c.Init(c)
	return c
}

func (c *CactusFlower) Clone() Behavior {
	cl := *c
	cl.rebind(&cl)
	return &cl
}

func (c *CactusFlower) canBeSupportedAt(blk Behavior) bool {
	support := blk.(blockGeometry).GetSide(math.Down, 1)
	return support.GetSupportType(math.Up).HasCenterSupport() || support.GetTypeId() == CACTUS
}

func (c *CactusFlower) CanBePlacedAt(blockReplace Behavior, clickVector math.Vector3, face math.Facing, isClickedBlock bool) bool {
	return c.canBeSupportedAt(blockReplace) && c.Flowable.CanBePlacedAt(blockReplace, clickVector, face, isClickedBlock)
}

func (c *CactusFlower) OnNearbyBlockChange() {
	if !c.canBeSupportedAt(c.self) {
		if world, err := c.position.GetWorld(); err == nil {
			world.UseBreakOn(c.position.AsVector3())
		}
	} else {
		c.Flowable.OnNearbyBlockChange()
	}
}

func (c *CactusFlower) GetFlameEncouragement() int { return 60 }

func (c *CactusFlower) GetFlammability() int { return 100 }
