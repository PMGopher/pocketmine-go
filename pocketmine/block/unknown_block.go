package block

import runtime "pocketmine-go/pocketmine/data/runtime"

// UnknownBlock is a port of pocketmine\block\UnknownBlock - represents a block which is
// unrecognized or not implemented.
type UnknownBlock struct {
	Transparent
	stateData int
}

func NewUnknownBlock(idInfo *BlockIdentifier, typeInfo *BlockTypeInfo, stateData int) *UnknownBlock {
	u := &UnknownBlock{
		Transparent: Transparent{NewBlock(idInfo, "Unknown", typeInfo)},
		stateData:   stateData,
	}
	u.Init(u)
	return u
}

func (u *UnknownBlock) Clone() Behavior {
	c := *u
	c.rebind(&c)
	return &c
}

// DescribeBlockItemState uses the raw type/state data instead of a real state describer, so no
// information (like colour) is lost - this might be an improperly registered plugin block.
func (u *UnknownBlock) DescribeBlockItemState(w runtime.DataDescriber) {
	w.Int(InternalStateDataBits, &u.stateData)
}

func (u *UnknownBlock) CanBePlaced() bool { return false }

func (u *UnknownBlock) GetDrops(item Item) []Item { return nil }
