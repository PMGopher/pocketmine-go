package block

// Mycelium is a port of pocketmine\block\Mycelium.
type Mycelium struct {
	Opaque
}

func NewMycelium(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Mycelium {
	m := &Mycelium{Opaque{NewBlock(idInfo, name, typeInfo)}}
	m.Init(m)
	return m
}

func (m *Mycelium) Clone() Behavior {
	c := *m
	c.rebind(&c)
	return &c
}

func (m *Mycelium) IsAffectedBySilkTouch() bool { return true }

func (m *Mycelium) TicksRandomly() bool { return true }

// OnRandomTick should pick a random nearby Dirt block (of DirtType::NORMAL, with a Transparent
// block above it) and spread mycelium onto it via BlockEventHelper::spread - needs the unported
// block registry (VanillaBlocks) and BlockEventHelper, so this is a no-op for now (see
// Block.GetDropsForCompatibleTool's doc comment for the same category of gap).
func (m *Mycelium) OnRandomTick() {}

// GetDropsForCompatibleTool should return [VanillaBlocks.DIRT().AsItem()] — needs the unported
// block registry and real Item construction (see Block.GetDropsForCompatibleTool's doc comment),
// so this returns nil for now.
func (m *Mycelium) GetDropsForCompatibleTool(item Item) []Item { return nil }
