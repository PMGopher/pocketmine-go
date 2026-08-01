package block

import "pocketmine-go/pocketmine/math"

// ChorusPlant is a port of pocketmine\block\ChorusPlant.
type ChorusPlant struct {
	Flowable

	Connections map[math.Facing]bool
}

func NewChorusPlant(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *ChorusPlant {
	c := &ChorusPlant{Flowable: Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}}, Connections: map[math.Facing]bool{}}
	c.Init(c)
	return c
}

// Clone can't use the usual one-line pattern: Connections is a map, a reference type - see
// Vine.Clone's doc comment for the same reasoning.
func (c *ChorusPlant) Clone() Behavior {
	cl := *c
	cl.Connections = make(map[math.Facing]bool, len(c.Connections))
	for k, v := range c.Connections {
		cl.Connections[k] = v
	}
	cl.rebind(&cl)
	return &cl
}

func (c *ChorusPlant) RecalculateCollisionBoxes() []math.AxisAlignedBB {
	bb := math.OneAABB()
	for _, facing := range math.AllFacing {
		if !c.Connections[facing] {
			bb.Trim(facing, 2.0/16)
		}
	}
	return []math.AxisAlignedBB{bb}
}

func (c *ChorusPlant) ReadStateFromWorld() Behavior {
	c.Block.ReadStateFromWorld()

	c.collisionBoxes = nil
	c.haveCollisionBoxes = false

	connections := map[math.Facing]bool{}
	for _, facing := range math.AllFacing {
		side := c.GetSide(facing, 1)
		switch side.GetTypeId() {
		case END_STONE, CHORUS_FLOWER, c.GetTypeId():
			connections[facing] = true
		}
	}
	c.Connections = connections

	return c.self
}

func (c *ChorusPlant) canBeSupportedBy(blk Behavior) bool {
	return blk.(blockGeometry).HasSameTypeId(c.self) || blk.GetTypeId() == END_STONE
}

func (c *ChorusPlant) canBeSupportedAt(blk Behavior) bool {
	geo := blk.(blockGeometry)
	world, err := blk.GetPosition().GetWorld()
	if err != nil {
		return false
	}
	down := geo.GetSide(math.Down, 1)
	verticalAir := down.GetTypeId() == AIR || geo.GetSide(math.Up, 1).GetTypeId() == AIR

	for _, sidePos := range blk.GetPosition().AsVector3().SidesAroundAxis(math.AxisY, 1) {
		side := world.GetBlockAt(sidePos.FloorX(), sidePos.FloorY(), sidePos.FloorZ())
		if side.GetTypeId() == CHORUS_PLANT {
			if !verticalAir {
				return false
			}
			if c.canBeSupportedBy(side.(blockGeometry).GetSide(math.Down, 1)) {
				return true
			}
		}
	}

	return c.canBeSupportedBy(down)
}

func (c *ChorusPlant) CanBePlacedAt(blockReplace Behavior, clickVector math.Vector3, face math.Facing, isClickedBlock bool) bool {
	return c.canBeSupportedAt(blockReplace) && c.Flowable.CanBePlacedAt(blockReplace, clickVector, face, isClickedBlock)
}

func (c *ChorusPlant) OnNearbyBlockChange() {
	if !c.canBeSupportedAt(c.self) {
		if world, err := c.position.GetWorld(); err == nil {
			world.UseBreakOn(c.position.AsVector3())
		}
	} else {
		c.Flowable.OnNearbyBlockChange()
	}
}

// GetDropsForCompatibleTool should have a 50% chance of dropping VanillaItems.CHORUS_FRUIT() —
// needs real Item construction from the unported item package (see
// Block.GetDropsForCompatibleTool's doc comment), so this returns nil for now.
func (c *ChorusPlant) GetDropsForCompatibleTool(item Item) []Item { return nil }
