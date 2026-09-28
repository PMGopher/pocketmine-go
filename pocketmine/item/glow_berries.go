package item

import "pocketmine-go/pocketmine/block"

// GlowBerries is a port of pocketmine\item\GlowBerries.
type GlowBerries struct {
	Food
}

func NewGlowBerries(identifier ItemIdentifier, name string) *GlowBerries {
	g := &GlowBerries{}
	g.Init(g, identifier, name)
	return g
}

func (g *GlowBerries) Clone() Item {
	cl := *g
	cl.rebind(&cl)
	return &cl
}

func (g *GlowBerries) GetFoodRestore() int { return 2 }

func (g *GlowBerries) GetSaturationRestore() float64 { return 0.4 }

// GetBlock is a port of GlowBerries::getBlock.
func (x *GlowBerries) GetBlock() block.Behavior { return block.VanillaBlock("cave_vines") }
