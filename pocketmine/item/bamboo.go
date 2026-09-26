package item

import (
	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/math"
)

// Bamboo is a port of pocketmine\item\Bamboo: placing it plants a bamboo sapling.
type Bamboo struct {
	ItemBase
}

func NewBamboo(identifier ItemIdentifier, name string) *Bamboo {
	b := &Bamboo{}
	b.Init(b, identifier, name)
	return b
}

func (b *Bamboo) Clone() Item {
	c := *b
	c.rebind(&c)
	return &c
}

func (b *Bamboo) GetFuelTime() int { return 50 }

func (b *Bamboo) GetBlock() block.Behavior { return block.VanillaBlock("bamboo_sapling") }

func (b *Bamboo) GetBlockForFace(clickedFace *math.Facing) block.Behavior { return b.GetBlock() }
