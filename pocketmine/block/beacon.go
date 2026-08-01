package block

// Beacon is a port of pocketmine\block\Beacon.
//
// The PHP original itself is a //TODO stub beyond GetLightLevel - beacon power/effect logic isn't
// implemented upstream either.
type Beacon struct {
	Transparent
}

func NewBeacon(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Beacon {
	b := &Beacon{Transparent{NewBlock(idInfo, name, typeInfo)}}
	b.Init(b)
	return b
}

func (b *Beacon) Clone() Behavior {
	c := *b
	c.rebind(&c)
	return &c
}

func (b *Beacon) GetLightLevel() int { return 15 }
