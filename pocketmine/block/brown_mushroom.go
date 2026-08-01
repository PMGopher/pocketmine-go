package block

// BrownMushroom is a port of pocketmine\block\BrownMushroom.
type BrownMushroom struct {
	RedMushroom
}

func NewBrownMushroom(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *BrownMushroom {
	b := &BrownMushroom{RedMushroom{Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}}}}
	b.Init(b)
	return b
}

func (b *BrownMushroom) Clone() Behavior {
	c := *b
	c.rebind(&c)
	return &c
}

func (b *BrownMushroom) GetLightLevel() int { return 1 }
