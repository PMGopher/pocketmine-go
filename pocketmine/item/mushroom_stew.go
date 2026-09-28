package item

// MushroomStew is a port of pocketmine\item\MushroomStew (its use is in item_use.go).
type MushroomStew struct {
	Food
}

func NewMushroomStew(identifier ItemIdentifier, name string) *MushroomStew {
	m := &MushroomStew{}
	m.Init(m, identifier, name)
	return m
}

func (m *MushroomStew) Clone() Item {
	c := *m
	c.rebind(&c)
	return &c
}

func (m *MushroomStew) GetMaxStackSize() int { return 1 }

func (m *MushroomStew) GetFoodRestore() int { return 6 }

func (m *MushroomStew) GetSaturationRestore() float64 { return 7.2 }
