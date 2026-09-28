package item

import (
	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/math"
)

// GlassBottle is a port of pocketmine\item\GlassBottle.
type GlassBottle struct {
	ItemBase
}

func NewGlassBottle(identifier ItemIdentifier, name string) *GlassBottle {
	g := &GlassBottle{}
	g.Init(g, identifier, name)
	return g
}

func (g *GlassBottle) Clone() Item {
	c := *g
	c.rebind(&c)
	return &c
}

// OnInteractBlock is a port of GlassBottle::onInteractBlock: clicking water fills the bottle.
func (g *GlassBottle) OnInteractBlock(player Player, blockReplace, blockClicked block.Behavior, face math.Facing, clickVector math.Vector3, returnedItems *[]Item) ItemUseResult {
	if blockClicked.GetTypeId() != block.WATER {
		return ItemUseResultNone
	}
	g.Pop()
	potion := VanillaItem("potion")
	potion.(*Potion).SetType(PotionTypeWater)
	*returnedItems = append(*returnedItems, potion)

	return ItemUseResultSuccess
}
