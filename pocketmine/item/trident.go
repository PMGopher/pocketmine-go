package item

// Trident is a port of pocketmine\item\Trident (its use methods are in item_use.go).
type Trident struct {
	Tool
}

func NewTrident(identifier ItemIdentifier, name string) *Trident {
	t := &Trident{}
	t.Init(t, identifier, name)
	return t
}

func (t *Trident) Clone() Item {
	c := *t
	c.rebind(&c)
	return &c
}

func (t *Trident) GetMaxDurability() int { return 251 }

func (t *Trident) GetAttackPoints() int { return 9 }
