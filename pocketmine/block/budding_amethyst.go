package block

import (
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/world/sound"
)

// BuddingAmethyst is a port of pocketmine\block\BuddingAmethyst.
type BuddingAmethyst struct {
	Opaque
}

func NewBuddingAmethyst(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *BuddingAmethyst {
	b := &BuddingAmethyst{Opaque{NewBlock(idInfo, name, typeInfo)}}
	b.Init(b)
	return b
}

func (b *BuddingAmethyst) Clone() Behavior {
	c := *b
	c.rebind(&c)
	return &c
}

func (b *BuddingAmethyst) OnProjectileHit(projectile Projectile, hitResult math.RayTraceResult) {
	world, err := b.position.GetWorld()
	if err != nil {
		return
	}
	world.AddSound(b.position.AsVector3(), sound.AmethystBlockChimeSound{})
	world.AddSound(b.position.AsVector3(), sound.BlockPunchSound{BlockStateID: b.GetStateId()})
}

func (b *BuddingAmethyst) TicksRandomly() bool { return true }

// OnRandomTick should grow an adjacent AmethystCluster bud (or advance an existing one's stage) —
// needs BlockEventHelper and the block registry (VanillaBlocks), neither ported yet, so this is a
// no-op for now.
func (b *BuddingAmethyst) OnRandomTick() {}

// GetDropsForCompatibleTool deliberately returns nothing, matching the PHP original's
// `return [];` (this isn't a not-yet-ported gap - budding amethyst can't be harvested at all).
func (b *BuddingAmethyst) GetDropsForCompatibleTool(item Item) []Item { return nil }
