package block

// NetherReactor is a port of pocketmine\block\NetherReactor.
type NetherReactor struct {
	Opaque
}

func NewNetherReactor(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *NetherReactor {
	n := &NetherReactor{Opaque{NewBlock(idInfo, name, typeInfo)}}
	n.Init(n)
	return n
}

func (n *NetherReactor) Clone() Behavior {
	c := *n
	c.rebind(&c)
	return &c
}

// GetDropsForCompatibleTool should return [VanillaItems.IRON_INGOT().SetCount(6),
// VanillaItems.DIAMOND().SetCount(3)] — needs real Item construction from the unported item
// package (see Block.GetDropsForCompatibleTool's doc comment), so this returns nil for now.
func (n *NetherReactor) GetDropsForCompatibleTool(item Item) []Item { return nil }
