package item

// Boat is a port of pocketmine\item\Boat. Placing the boat entity is //TODO in PHP too, so only
// the type, fuel and stack size are ported.
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
