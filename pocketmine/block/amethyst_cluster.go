package block

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/world/sound"
)

const (
	AmethystClusterStageSmallBud  = 0
	AmethystClusterStageMediumBud = 1
	AmethystClusterStageLargeBud  = 2
	AmethystClusterStageCluster   = 3
)

// AmethystCluster is a port of pocketmine\block\AmethystCluster.
type AmethystCluster struct {
	Transparent
	FacingComponent

	Stage int
}

func NewAmethystCluster(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *AmethystCluster {
	a := &AmethystCluster{
		Transparent:     Transparent{NewBlock(idInfo, name, typeInfo)},
		FacingComponent: NewFacingComponent(),
		Stage:           AmethystClusterStageCluster,
	}
	a.Init(a)
	return a
}

func (a *AmethystCluster) Clone() Behavior {
	c := *a
	c.rebind(&c)
	return &c
}

func (a *AmethystCluster) DescribeBlockItemState(w runtime.DataDescriber) {
	w.BoundedIntAuto(AmethystClusterStageSmallBud, AmethystClusterStageCluster, &a.Stage)
}

func (a *AmethystCluster) GetStage() int { return a.Stage }

// SetStage panics if stage is out of range, mirroring the PHP original's
// \InvalidArgumentException (a programmer error at the call site).
func (a *AmethystCluster) SetStage(stage int) {
	if stage < AmethystClusterStageSmallBud || stage > AmethystClusterStageCluster {
		panic("Size must be in range 0 ... 3")
	}
	a.Stage = stage
}

func (a *AmethystCluster) GetLightLevel() int {
	switch a.Stage {
	case AmethystClusterStageSmallBud:
		return 1
	case AmethystClusterStageMediumBud:
		return 2
	case AmethystClusterStageLargeBud:
		return 4
	case AmethystClusterStageCluster:
		return 5
	default:
		panic("invalid stage")
	}
}

func (a *AmethystCluster) RecalculateCollisionBoxes() []math.AxisAlignedBB {
	myAxis := math.FacingAxis(a.Facing)
	squash := 3.0 / 16
	if a.Stage == AmethystClusterStageSmallBud {
		squash = 4.0 / 16
	}

	box := math.OneAABB()
	for _, axis := range []math.Axis{math.AxisY, math.AxisZ, math.AxisX} {
		if axis != myAxis {
			box.Squash(axis, squash)
		}
	}

	trimAmount := 7.0 / 16
	if a.Stage != AmethystClusterStageCluster {
		trimAmount = float64(a.Stage+3) / 16
	}
	box.Trim(a.Facing, 1-trimAmount)

	return []math.AxisAlignedBB{box}
}

func (a *AmethystCluster) canBeSupportedAt(blk Behavior, facing math.Facing) bool {
	return blk.(blockGeometry).GetAdjacentSupportType(facing) == blockutils.SupportTypeFull
}

func (a *AmethystCluster) GetSupportType(facing math.Facing) blockutils.SupportType {
	return blockutils.SupportTypeNone
}

func (a *AmethystCluster) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	if !a.canBeSupportedAt(blockReplace, math.Opposite(face)) {
		return false
	}
	a.Facing = face
	return a.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
}

func (a *AmethystCluster) OnNearbyBlockChange() {
	if !a.canBeSupportedAt(a.self, math.Opposite(a.Facing)) {
		if world, err := a.position.GetWorld(); err == nil {
			world.UseBreakOn(a.position.AsVector3())
		}
	}
}

func (a *AmethystCluster) OnProjectileHit(projectile Projectile, hitResult math.RayTraceResult) {
	world, err := a.position.GetWorld()
	if err != nil {
		return
	}
	world.AddSound(a.position.AsVector3(), sound.AmethystBlockChimeSound{})
	world.AddSound(a.position.AsVector3(), sound.BlockPunchSound{BlockStateID: a.GetStateId()})
}

func (a *AmethystCluster) IsAffectedBySilkTouch() bool { return true }

// GetDropsForCompatibleTool/GetDropsForIncompatibleTool should return amethyst shards scaled via
// FortuneDropHelper — needs real Item construction from the unported item package (see
// Block.GetDropsForCompatibleTool's doc comment), so these return nil for now.
func (a *AmethystCluster) GetDropsForCompatibleTool(item Item) []Item { return nil }

func (a *AmethystCluster) GetDropsForIncompatibleTool(item Item) []Item { return nil }
