package block

import "pocketmine-go/pocketmine/math"

// GrassPath is a port of pocketmine\block\GrassPath.
type GrassPath struct {
	Transparent
}

func NewGrassPath(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *GrassPath {
	g := &GrassPath{Transparent{NewBlock(idInfo, name, typeInfo)}}
	g.Init(g)
	return g
}

func (g *GrassPath) Clone() Behavior {
	c := *g
	c.rebind(&c)
	return &c
}

func (g *GrassPath) RecalculateCollisionBoxes() []math.AxisAlignedBB {
	return []math.AxisAlignedBB{math.OneAABB().TrimmedCopy(math.Up, 1.0/16)}
}

// OnNearbyBlockChange should turn this block into VanillaBlocks.DIRT() when a solid block is
// placed above it — needs the unported block registry (VanillaBlocks), so this is a no-op for
// now (see Block.GetDropsForCompatibleTool's doc comment for the same category of gap).
func (g *GrassPath) OnNearbyBlockChange() {}

func (g *GrassPath) IsAffectedBySilkTouch() bool { return true }

// GetDropsForCompatibleTool should return [VanillaBlocks.DIRT().AsItem()] — needs the unported
// block registry and real Item construction (see Block.GetDropsForCompatibleTool's doc comment),
// so this returns nil for now.
func (g *GrassPath) GetDropsForCompatibleTool(item Item) []Item { return nil }
