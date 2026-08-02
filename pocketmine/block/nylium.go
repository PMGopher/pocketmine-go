package block

import (
	"math/rand"

	"pocketmine-go/pocketmine/math"
)

// Nylium is a port of pocketmine\block\Nylium.
type Nylium struct {
	Opaque

	// Vegetation mirrors the PHP constructor's Block[] $vegetation - the blocks that can be grown
	// on this Nylium block using Bone Meal.
	Vegetation []Behavior
}

func NewNylium(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo, vegetation []Behavior) *Nylium {
	n := &Nylium{Opaque: Opaque{NewBlock(idInfo, name, typeInfo)}, Vegetation: vegetation}
	n.Init(n)
	return n
}

func (n *Nylium) Clone() Behavior {
	c := *n
	c.rebind(&c)
	return &c
}

func (n *Nylium) IsAffectedBySilkTouch() bool { return true }

func (n *Nylium) TicksRandomly() bool { return true }

// OnRandomTick is a port of Nylium::onRandomTick.
func (n *Nylium) OnRandomTick() {
	world, err := n.position.GetWorld()
	if err != nil {
		return
	}
	if !n.self.(blockGeometry).GetSide(math.Up, 1).IsTransparent() {
		Spread(n.self, VanillaNetherrack(), n.self)
		return
	}

	pos := n.position.AsVector3()
	for i := 0; i < 4; i++ {
		rx := pos.FloorX() - 1 + rand.Intn(3)
		ry := pos.FloorY() - 2 + rand.Intn(3)
		rz := pos.FloorZ() - 1 + rand.Intn(3)
		n.tryConvertNetherrackAt(world, rx, ry, rz)
	}
}

// tryConvertNetherrackAt is the deterministic (given world coordinates) rest of the spread loop in
// Nylium::onRandomTick, split out from the random position sample above so it's directly testable
// - same pattern as Grass.trySpreadOnto/Mycelium.trySpreadOnto.
func (n *Nylium) tryConvertNetherrackAt(world World, x, y, z int) {
	blk := world.GetBlockAt(x, y, z)
	if blk.GetTypeId() == NETHERRACK && world.GetBlockAt(x, y+1, z).IsTransparent() {
		Spread(blk, n.self.Clone(), n.self)
	}
}

// OnInteract should fertilizer-grow Vegetation blocks nearby - needs a Fertilizer item marker and
// World.GetBlock(Position)/IsInWorld, none ported yet. Block's default OnInteract (return false)
// already matches this gap, so there's nothing to override here.

// GetDropsForCompatibleTool should return [VanillaBlocks.NETHERRACK().AsItem()] — needs the
// unported block registry and real Item construction (see Block.GetDropsForCompatibleTool's doc
// comment), so this returns nil for now.
func (n *Nylium) GetDropsForCompatibleTool(item Item) []Item { return nil }
