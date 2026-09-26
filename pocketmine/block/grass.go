package block

import (
	"math/rand"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/utils"
	"pocketmine-go/pocketmine/world/sound"

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

// OnInteract is a port of Grass::onInteract: bone meal grows grass and flowers around it, a hoe
// tills it into farmland and a shovel makes a path.
func (g *Grass) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	if g.self.(blockGeometry).GetSide(math.Up, 1).GetTypeId() != AIR {
		return false
	}
	world, err := g.position.GetWorld()
	if err != nil {
		return false
	}
	if isFertilizer(item) {
		item.Pop()
		if GrowGrassFunc != nil {
			GrowGrassFunc(world, g.position.FloorX(), g.position.FloorY(), g.position.FloorZ(), utils.NewRandom(rand.Int()), 8, 2)
		}
		return true
	}
	if face != math.Down {
		var newBlock Behavior
		if isHoe(item) {
			newBlock = VanillaBlock("farmland")
		} else if isShovel(item) {
			newBlock = VanillaBlock("grass_path")
		}
		if newBlock != nil {
			applyDamage(item, 1)
			world.AddSound(g.position.Add(0.5, 0.5, 0.5), sound.ItemUseOnBlockSound{BlockStateID: newBlock.GetStateId()})
			_ = world.SetBlock(g.position, newBlock)
			return true
		}
	}
	return false
}

// GetDropsForCompatibleTool is a port of Grass::getDropsForCompatibleTool.
func (g *Grass) GetDropsForCompatibleTool(item Item) []Item {
	dropped := asItemOrNil(VanillaDirt())
	if dropped == nil {
		return nil
	}
	return []Item{dropped}
}
