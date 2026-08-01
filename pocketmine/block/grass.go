package block

// Grass is a port of pocketmine\block\Grass.
type Grass struct {
	Opaque
}

func NewGrass(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Grass {
	g := &Grass{Opaque{NewBlock(idInfo, name, typeInfo)}}
	g.Init(g)
	return g
}

func (g *Grass) Clone() Behavior {
	c := *g
	c.rebind(&c)
	return &c
}

func (g *Grass) IsAffectedBySilkTouch() bool { return true }

func (g *Grass) TicksRandomly() bool { return true }

// OnRandomTick should die back to Dirt in low light, or spread onto nearby eligible Dirt blocks
// in bright light, via BlockEventHelper - needs the unported block registry (VanillaBlocks) and
// BlockEventHelper, so this is a no-op for now (see Block.GetDropsForCompatibleTool's doc comment
// for the same category of gap).
func (g *Grass) OnRandomTick() {}

// OnInteract should grow tall grass with Fertilizer, or till into Farmland/GrassPath with a
// Hoe/Shovel - needs the unported Fertilizer/Hoe/Shovel item markers, the block registry, and the
// TallGrass world-gen object. Block's default OnInteract (return false) already matches this gap,
// so there's nothing to override here.

// GetDropsForCompatibleTool should return [VanillaBlocks.DIRT().AsItem()] — needs the unported
// block registry and real Item construction (see Block.GetDropsForCompatibleTool's doc comment),
// so this returns nil for now.
func (g *Grass) GetDropsForCompatibleTool(item Item) []Item { return nil }
