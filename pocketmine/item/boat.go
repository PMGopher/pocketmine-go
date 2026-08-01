package item

// Boat is a port of pocketmine\item\Boat. Placing/spawning the boat entity (marked //TODO even in
// the PHP original) needs the unported entity package, so only the type/fuel/stack-size state is
// ported here.
type Boat struct {
	ItemBase

	BoatTypeValue BoatType
}

func NewBoat(identifier ItemIdentifier, name string, boatType BoatType) *Boat {
	b := &Boat{BoatTypeValue: boatType}
	b.Init(b, identifier, name)
	return b
}

func (b *Boat) Clone() Item {
	c := *b
	c.rebind(&c)
	return &c
}

func (b *Boat) GetType() BoatType { return b.BoatTypeValue }

func (b *Boat) GetFuelTime() int { return 1200 }

func (b *Boat) GetMaxStackSize() int { return 1 }
