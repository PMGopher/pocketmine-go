package block

import (
	"math/rand"

	blockutils "pocketmine-go/pocketmine/block/utils"
)

// Grass is a port of pocketmine\block\Grass.
type Grass struct {
	Opaque
}

func NewGrass(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Grass {
	g := &Grass{Opaque{NewBlock(idInfo, name, typeInfo)}}
	g.Init(g)
	return g
}

func (g *Grass) Clone() Behavior {
	c := *g
	c.rebind(&c)
	return &c
}

func (g *Grass) IsAffectedBySilkTouch() bool { return true }

func (g *Grass) TicksRandomly() bool { return true }

// OnRandomTick is a port of Grass::onRandomTick.
func (g *Grass) OnRandomTick() {
	world, err := g.position.GetWorld()
	if err != nil {
		return
	}
	pos := g.position.AsVector3()
	x, y, z := pos.FloorX(), pos.FloorY(), pos.FloorZ()

	lightAbove := world.GetFullLightAt(x, y+1, z)
	if lightAbove < 4 && world.GetBlockAt(x, y+1, z).GetLightFilter() >= 2 {
		// grass dies
		Spread(g.self, VanillaDirt(), g.self)
	} else if lightAbove >= 9 {
		// try grass spread
		for i := 0; i < 4; i++ {
			rx := x - 1 + rand.Intn(3)
			ry := y - 3 + rand.Intn(5)
			rz := z - 1 + rand.Intn(3)
			g.trySpreadOnto(world, rx, ry, rz)
		}
	}
}

// trySpreadOnto is the deterministic (given world coordinates) rest of Grass::onRandomTick's
// spread loop, split out from the random position sampling above so it's directly testable -
// same "extract a helper for the decision-and-act logic" pattern as
// BuddingAmethyst.tryGrowBud.
func (g *Grass) trySpreadOnto(world World, x, y, z int) {
	b := world.GetBlockAt(x, y, z)
	dirt, ok := b.(*Dirt)
	if !ok || dirt.GetDirtType() != blockutils.DirtTypeNormal ||
		world.GetFullLightAt(x, y+1, z) < 4 ||
		world.GetBlockAt(x, y+1, z).GetLightFilter() >= 2 {
		return
	}
	Spread(dirt, VanillaGrass(), g.self)
}

// OnInteract should grow tall grass with Fertilizer, or till into Farmland/GrassPath with a
// Hoe/Shovel - needs the unported Fertilizer/Hoe/Shovel item markers, the block registry, and the
// TallGrass world-gen object. Block's default OnInteract (return false) already matches this gap,
// so there's nothing to override here.

// GetDropsForCompatibleTool should return [VanillaBlocks.DIRT().AsItem()] — needs the unported
// block registry and real Item construction (see Block.GetDropsForCompatibleTool's doc comment),
// so this returns nil for now.
func (g *Grass) GetDropsForCompatibleTool(item Item) []Item { return nil }
