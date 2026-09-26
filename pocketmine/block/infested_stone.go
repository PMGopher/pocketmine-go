package block

// InfestedStone is a port of pocketmine\block\InfestedStone.
//
// The PHP original stores the imitated block's state ID and resolves it back to a real Block via
// RuntimeBlockStateRegistry in GetImitatedBlock.
type InfestedStone struct {
	Opaque

	ImitatedStateID int
}

func NewInfestedStone(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo, imitated Behavior) *InfestedStone {
	i := &InfestedStone{
		Opaque:          Opaque{NewBlock(idInfo, name, typeInfo)},
		ImitatedStateID: imitated.GetStateId(),
	}
	i.Init(i)
	return i
}

func (i *InfestedStone) Clone() Behavior {
	c := *i
	c.rebind(&c)
	return &c
}

func (i *InfestedStone) GetImitatedStateID() int { return i.ImitatedStateID }

func (i *InfestedStone) GetDropsForCompatibleTool(item Item) []Item { return nil }

// GetImitatedBlock is a port of InfestedStone::getImitatedBlock.
func (i *InfestedStone) GetImitatedBlock() Behavior {
	return GetRuntimeBlockStateRegistry().FromStateId(i.ImitatedStateID)
}

// GetSilkTouchDrops is a port of InfestedStone::getSilkTouchDrops.
func (i *InfestedStone) GetSilkTouchDrops(item Item) []Item {
	return itemDrops(asItemOrNil(i.GetImitatedBlock()))
}

func (i *InfestedStone) IsAffectedBySilkTouch() bool { return true }
