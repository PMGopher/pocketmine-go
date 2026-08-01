package item

// BlazeRod is a port of pocketmine\item\BlazeRod.
type BlazeRod struct {
	ItemBase
}

func NewBlazeRod(identifier ItemIdentifier, name string) *BlazeRod {
	b := &BlazeRod{}
	b.Init(b, identifier, name)
	return b
}

func (b *BlazeRod) Clone() Item {
	c := *b
	c.rebind(&c)
	return &c
}

func (b *BlazeRod) GetFuelTime() int { return 2400 }
