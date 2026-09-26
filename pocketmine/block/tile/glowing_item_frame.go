package tile

import "pocketmine-go/pocketmine/math"

// GlowingItemFrame is a port of pocketmine\block\tile\GlowingItemFrame: an ItemFrame saved as
// "GlowItemFrame".
type GlowingItemFrame struct {
	ItemFrame
}

func NewGlowingItemFrame(world World, pos math.Vector3) *GlowingItemFrame {
	g := &GlowingItemFrame{ItemFrame: ItemFrame{itemDropChance: 1.0}}
	g.SpawnableBase = SpawnableBase{TileBase: NewTileBase(world, pos)}
	g.Init(g)
	return g
}

func (g *GlowingItemFrame) SaveID() string { return "GlowItemFrame" }
