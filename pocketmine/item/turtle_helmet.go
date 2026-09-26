package item

import "pocketmine-go/pocketmine/entity/effect"

// TurtleHelmet is a port of pocketmine\item\TurtleHelmet.
type TurtleHelmet struct {
	Armor
}

func NewTurtleHelmet(identifier ItemIdentifier, name string, info ArmorTypeInfo, enchantmentTags ...string) *TurtleHelmet {
	t := &TurtleHelmet{Armor: Armor{ArmorInfo: info}}
	t.Init(t, identifier, name)
	t.enchantmentTags = enchantmentTags
	return t
}

func (t *TurtleHelmet) Clone() Item {
	c := *t
	c.rebind(&c)
	return &c
}

// OnTickWorn is a port of TurtleHelmet::onTickWorn: grants Water Breathing to a Human wearing it
// out of water.
func (t *TurtleHelmet) OnTickWorn(entity Living) bool {
	if isHuman(entity) && !entity.IsUnderwater() {
		entity.GetEffects().Add(effect.NewEffectInstanceFull(effect.VanillaWaterBreathing(), intPtr(200), 0, false, false, nil, false))
		return true
	}
	return false
}

func intPtr(v int) *int { return &v }
