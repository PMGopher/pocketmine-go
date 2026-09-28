package item

import (
	"pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/world/sound"
)

const itemCooldownTagGoatHorn = "goat_horn"

// GoatHorn is a port of pocketmine\item\GoatHorn.
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

// OnClickAir is a port of GoatHorn::onClickAir: the horn's sound is played where the player is.
func (g *GoatHorn) OnClickAir(player Player, directionVector math.Vector3, returnedItems *[]Item) ItemUseResult {
	if world := entityWorld(player); world != nil {
		world.AddSound(player.GetPosition(), sound.GoatHornSound{HornType: int(g.HornType)})
	}
	return ItemUseResultSuccess
}
