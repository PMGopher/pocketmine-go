package block

// Bookshelf is a port of pocketmine\block\Bookshelf.
type Bookshelf struct {
	Opaque
}

func NewBookshelf(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Bookshelf {
	b := &Bookshelf{Opaque{NewBlock(idInfo, name, typeInfo)}}
	b.Init(b)
	return b
}

func (b *Bookshelf) Clone() Behavior {
	c := *b
	c.rebind(&c)
	return &c
}

// GetDropsForCompatibleTool is a port of Bookshelf::getDropsForCompatibleTool.
func (b *Bookshelf) GetDropsForCompatibleTool(item Item) []Item {
	return itemDrops(vanillaItemCount("book", 3))
}

func (b *Bookshelf) IsAffectedBySilkTouch() bool { return true }

func (b *Bookshelf) GetFuelTime() int { return 300 }

func (b *Bookshelf) GetFlameEncouragement() int { return 30 }

func (b *Bookshelf) GetFlammability() int { return 20 }
