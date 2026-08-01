package block

// Nylium is a port of pocketmine\block\Nylium.
type Nylium struct {
	Opaque

	// Vegetation mirrors the PHP constructor's Block[] $vegetation - the blocks that can be grown
	// on this Nylium block using Bone Meal.
	Vegetation []Behavior
}

func NewNylium(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo, vegetation []Behavior) *Nylium {
	n := &Nylium{Opaque: Opaque{NewBlock(idInfo, name, typeInfo)}, Vegetation: vegetation}
	n.Init(n)
	return n
}

func (n *Nylium) Clone() Behavior {
	c := *n
	c.rebind(&c)
	return &c
}

func (n *Nylium) IsAffectedBySilkTouch() bool { return true }

func (n *Nylium) TicksRandomly() bool { return true }

// OnRandomTick should revert to Netherrack when covered, or spread onto nearby Netherrack via
// BlockEventHelper - needs the unported block registry (VanillaBlocks) and BlockEventHelper, so
// this is a no-op for now (same gap category as Mycelium/Grass's OnRandomTick).
func (n *Nylium) OnRandomTick() {}

// OnInteract should fertilizer-grow Vegetation blocks nearby - needs a Fertilizer item marker and
// World.GetBlock(Position)/IsInWorld, none ported yet. Block's default OnInteract (return false)
// already matches this gap, so there's nothing to override here.

// GetDropsForCompatibleTool should return [VanillaBlocks.NETHERRACK().AsItem()] — needs the
// unported block registry and real Item construction (see Block.GetDropsForCompatibleTool's doc
// comment), so this returns nil for now.
func (n *Nylium) GetDropsForCompatibleTool(item Item) []Item { return nil }
