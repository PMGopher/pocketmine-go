package item

// Totem is a port of pocketmine\item\Totem. Its actual "prevent death" behaviour lives on the
// Human/Player side in PHP (Totem itself is just a stack-size-1 marker item), so there's nothing
// else to port here.
type Totem struct {
	ItemBase
}

func NewTotem(identifier ItemIdentifier, name string) *Totem {
	t := &Totem{}
	t.Init(t, identifier, name)
	return t
}

func (t *Totem) Clone() Item {
	c := *t
	c.rebind(&c)
	return &c
}

func (t *Totem) GetMaxStackSize() int { return 1 }
