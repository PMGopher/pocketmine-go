package block

// InfestedStone is a port of pocketmine\block\InfestedStone.
//
// The PHP original stores the imitated block's state ID and resolves it back to a real Block via
// RuntimeBlockStateRegistry (an unported world/block-state registry) in GetImitatedBlock. That
// resolution isn't implemented yet — only the raw state ID is stored/exposed for now.
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

// GetSilkTouchDrops should return [i.GetImitatedBlock().AsItem()] — needs
// RuntimeBlockStateRegistry (see the type doc comment), so this returns nil for now.
func (i *InfestedStone) GetSilkTouchDrops(item Item) []Item { return nil }

func (i *InfestedStone) IsAffectedBySilkTouch() bool { return true }
