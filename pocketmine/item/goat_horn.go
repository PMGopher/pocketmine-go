package item

import (
	runtime "pocketmine-go/pocketmine/data/runtime"
)

const itemCooldownTagGoatHorn = "goat_horn"

// GoatHorn is a port of pocketmine\item\GoatHorn. OnClickAir (playing the horn sound) and
// CanStartUsingItem need a real Player/World - see the Item interface's doc comment on
// Player/Entity-interaction methods.
type GoatHorn struct {
	ItemBase

	HornType GoatHornType
}

func NewGoatHorn(identifier ItemIdentifier, name string) *GoatHorn {
	g := &GoatHorn{HornType: GoatHornTypePonder}
	g.Init(g, identifier, name)
	return g
}

func (g *GoatHorn) Clone() Item {
	c := *g
	c.rebind(&c)
	return &c
}

func (g *GoatHorn) GetHornType() GoatHornType { return g.HornType }

func (g *GoatHorn) SetHornType(t GoatHornType) { g.HornType = t }

func (g *GoatHorn) GetMaxStackSize() int { return 1 }

func (g *GoatHorn) GetCooldownTicks() int { return 140 }

func (g *GoatHorn) GetCooldownTag() (string, bool) { return itemCooldownTagGoatHorn, true }

func (g *GoatHorn) describeState(w runtime.DataDescriber) {
	t := int(g.HornType)
	w.BoundedIntAuto(int(GoatHornTypePonder), int(GoatHornTypeDream), &t)
	g.HornType = GoatHornType(t)
}
