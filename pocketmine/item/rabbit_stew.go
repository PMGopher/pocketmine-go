package item

// RabbitStew is a port of pocketmine\item\RabbitStew. GetResidue (should return
// VanillaItems.BOWL()) isn't ported - see Food's doc comment for why.
type RabbitStew struct {
	Food
}

func NewRabbitStew(identifier ItemIdentifier, name string) *RabbitStew {
	r := &RabbitStew{}
	r.Init(r, identifier, name)
	return r
}

func (r *RabbitStew) Clone() Item {
	c := *r
	c.rebind(&c)
	return &c
}

func (r *RabbitStew) GetMaxStackSize() int { return 1 }

func (r *RabbitStew) GetFoodRestore() int { return 10 }

func (r *RabbitStew) GetSaturationRestore() float64 { return 12 }
