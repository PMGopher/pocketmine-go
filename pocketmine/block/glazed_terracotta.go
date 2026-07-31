package block

import (
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// GlazedTerracotta is a port of pocketmine\block\GlazedTerracotta.
type GlazedTerracotta struct {
	Opaque
	ColorComponent
	HorizontalFacingComponent
}

func NewGlazedTerracotta(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *GlazedTerracotta {
	g := &GlazedTerracotta{
		Opaque:                    Opaque{NewBlock(idInfo, name, typeInfo)},
		ColorComponent:            NewColorComponent(),
		HorizontalFacingComponent: NewHorizontalFacingComponent(),
	}
	g.Init(g)
	return g
}

func (g *GlazedTerracotta) Clone() Behavior {
	c := *g
	c.rebind(&c)
	return &c
}

func (g *GlazedTerracotta) DescribeBlockItemState(w runtime.DataDescriber) { g.DescribeColor(w) }

func (g *GlazedTerracotta) DescribeBlockOnlyState(w runtime.DataDescriber) {
	g.DescribeHorizontalFacing(w)
}

// Place is a port of pocketmine\block\utils\FacesOppositePlacingPlayerTrait::place.
func (g *GlazedTerracotta) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	if player != nil {
		g.Facing = math.Opposite(player.GetHorizontalFacing())
	}
	return g.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
}
