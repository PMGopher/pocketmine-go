package item

// TurtleHelmet is a port of pocketmine\item\TurtleHelmet. Its only override in PHP is OnTickWorn
// (granting Water Breathing while worn out of water) - needs a real Living/Human entity, so it's
// skipped per the Item interface's doc comment on Player/Entity-interaction methods. Everything
// else (defense points, durability, etc.) comes from the embedded Armor exactly as with any other
// armor piece.
type TurtleHelmet struct {
	Armor
}

func NewTurtleHelmet(identifier ItemIdentifier, name string, info ArmorTypeInfo) *TurtleHelmet {
	t := &TurtleHelmet{Armor: Armor{ArmorInfo: info}}
	t.Init(t, identifier, name)
	return t
}

func (t *TurtleHelmet) Clone() Item {
	c := *t
	c.rebind(&c)
	return &c
}
