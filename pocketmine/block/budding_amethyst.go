package block

import (
	"math/rand"

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

// OnRandomTick is a port of BuddingAmethyst::onRandomTick, minus the water-logging TODO already
// marked as unimplemented in the PHP original too.
func (b *BuddingAmethyst) OnRandomTick() {
	if rand.Intn(5) != 0 { // mt_rand(1, 5) === 1
		return
	}
	face := math.AllFacing[rand.Intn(len(math.AllFacing))]
	b.tryGrowBud(face)
}

// tryGrowBud is the deterministic (given a face) rest of BuddingAmethyst::onRandomTick, split out
// from the two random rolls above so it's directly testable.
func (b *BuddingAmethyst) tryGrowBud(face math.Facing) {
	adjacent := b.self.(blockGeometry).GetSide(face, 1)

	newStage := -1
	if adjacent.GetTypeId() == AIR {
		newStage = AmethystClusterStageSmallBud
	} else if cluster, ok := adjacent.(*AmethystCluster); ok &&
		cluster.Stage < AmethystClusterStageCluster && cluster.Facing == face {
		newStage = cluster.Stage + 1
	}
	if newStage == -1 {
		return
	}

	newCluster := VanillaAmethystCluster().(*AmethystCluster)
	newCluster.Stage = newStage
	newCluster.SetFacing(face)
	Grow(adjacent, newCluster, nil)
}

// GetDropsForCompatibleTool deliberately returns nothing, matching the PHP original's
// `return [];` (this isn't a not-yet-ported gap - budding amethyst can't be harvested at all).
func (b *BuddingAmethyst) GetDropsForCompatibleTool(item Item) []Item { return nil }
