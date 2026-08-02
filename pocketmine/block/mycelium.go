package block

import (
	"math/rand"

	blockutils "pocketmine-go/pocketmine/block/utils"
)

// Mycelium is a port of pocketmine\block\Mycelium.
type Mycelium struct {
	Opaque
}

func NewMycelium(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Mycelium {
	m := &Mycelium{Opaque{NewBlock(idInfo, name, typeInfo)}}
	m.Init(m)
	return m
}

func (m *Mycelium) Clone() Behavior {
	c := *m
	c.rebind(&c)
	return &c
}

func (m *Mycelium) IsAffectedBySilkTouch() bool { return true }

func (m *Mycelium) TicksRandomly() bool { return true }

// OnRandomTick is a port of Mycelium::onRandomTick. The "//TODO: light levels" comment in the PHP
// original is copied as-is - not a gap introduced by this port.
func (m *Mycelium) OnRandomTick() {
	world, err := m.position.GetWorld()
	if err != nil {
		return
	}
	pos := m.position.AsVector3()
	x := pos.FloorX() - 1 + rand.Intn(3)
	y := pos.FloorY() - 2 + rand.Intn(5)
	z := pos.FloorZ() - 1 + rand.Intn(3)
	m.trySpreadOnto(world, x, y, z)
}

// trySpreadOnto is the deterministic (given world coordinates) rest of Mycelium::onRandomTick,
// split out from the random position sample above so it's directly testable - same pattern as
// Grass.trySpreadOnto.
func (m *Mycelium) trySpreadOnto(world World, x, y, z int) {
	blk := world.GetBlockAt(x, y, z)
	if dirt, ok := blk.(*Dirt); ok && dirt.GetDirtType() == blockutils.DirtTypeNormal {
		if world.GetBlockAt(x, y+1, z).IsTransparent() {
			Spread(dirt, VanillaMycelium(), m.self)
		}
	}
}

// GetDropsForCompatibleTool is a port of Mycelium::getDropsForCompatibleTool.
func (m *Mycelium) GetDropsForCompatibleTool(item Item) []Item {
	dropped := asItemOrNil(VanillaDirt())
	if dropped == nil {
		return nil
	}
	return []Item{dropped}
}
