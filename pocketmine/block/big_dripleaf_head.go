package block

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/world/sound"
)

// BigDripleafHead is a port of pocketmine\block\BigDripleafHead.
type BigDripleafHead struct {
	BaseBigDripleaf

	LeafState blockutils.DripleafState
}

func NewBigDripleafHead(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *BigDripleafHead {
	b := &BigDripleafHead{
		BaseBigDripleaf: BaseBigDripleaf{
			Transparent:               Transparent{NewBlock(idInfo, name, typeInfo)},
			HorizontalFacingComponent: NewHorizontalFacingComponent(),
		},
		LeafState: blockutils.DripleafStateStable,
	}
	b.Init(b)
	return b
}

func (b *BigDripleafHead) Clone() Behavior {
	c := *b
	c.rebind(&c)
	return &c
}

func (b *BigDripleafHead) DescribeBlockOnlyState(w runtime.DataDescriber) {
	b.BaseBigDripleaf.DescribeBlockOnlyState(w)
	state := int(b.LeafState)
	w.BoundedIntAuto(int(blockutils.DripleafStateStable), int(blockutils.DripleafStateFullTilt), &state)
	b.LeafState = blockutils.DripleafState(state)
}

func (b *BigDripleafHead) IsHead() bool { return true }

func (b *BigDripleafHead) GetLeafState() blockutils.DripleafState { return b.LeafState }

func (b *BigDripleafHead) SetLeafState(leafState blockutils.DripleafState) {
	b.LeafState = leafState
}

func (b *BigDripleafHead) HasEntityCollision() bool { return true }

func (b *BigDripleafHead) setTiltAndScheduleTick(tilt blockutils.DripleafState) {
	world, err := b.position.GetWorld()
	if err != nil {
		return
	}
	b.LeafState = tilt
	if err := world.SetBlock(b.position, b.self); err != nil {
		panic(err)
	}
	if delay, ok := tilt.GetScheduledUpdateDelayTicks(); ok {
		world.ScheduleDelayedBlockUpdate(b.position.AsVector3(), delay)
	}
}

func (b *BigDripleafHead) getLeafTopOffset() float64 {
	switch b.LeafState {
	case blockutils.DripleafStateStable, blockutils.DripleafStateUnstable:
		return 1.0 / 16
	case blockutils.DripleafStatePartialTilt:
		return 3.0 / 16
	default:
		return 0
	}
}

func (b *BigDripleafHead) OnEntityInside(entity Entity) bool {
	if _, isProjectile := entity.(Projectile); !isProjectile && b.LeafState == blockutils.DripleafStateStable {
		pos := b.position.AsVector3()
		intersection := math.OneAABB().OffsetCopy(pos.X, pos.Y, pos.Z).TrimmedCopy(math.Down, 1-b.getLeafTopOffset())
		if entity.GetBoundingBox().IntersectsWith(intersection, 0) {
			b.setTiltAndScheduleTick(blockutils.DripleafStateUnstable)
			return false
		}
	}
	return true
}

func (b *BigDripleafHead) OnProjectileHit(projectile Projectile, hitResult math.RayTraceResult) {
	if b.LeafState != blockutils.DripleafStateFullTilt {
		b.setTiltAndScheduleTick(blockutils.DripleafStateFullTilt)
		if world, err := b.position.GetWorld(); err == nil {
			world.AddSound(b.position.AsVector3(), sound.DripleafTiltDownSound{})
		}
	}
}

func (b *BigDripleafHead) OnScheduledUpdate() {
	if b.LeafState == blockutils.DripleafStateStable {
		return
	}
	world, err := b.position.GetWorld()
	if err != nil {
		return
	}
	if b.LeafState == blockutils.DripleafStateFullTilt {
		b.LeafState = blockutils.DripleafStateStable
		if err := world.SetBlock(b.position, b.self); err != nil {
			panic(err)
		}
		world.AddSound(b.position.AsVector3(), sound.DripleafTiltUpSound{})
		return
	}

	next := blockutils.DripleafStatePartialTilt
	if b.LeafState == blockutils.DripleafStatePartialTilt {
		next = blockutils.DripleafStateFullTilt
	}
	b.setTiltAndScheduleTick(next)
	world.AddSound(b.position.AsVector3(), sound.DripleafTiltDownSound{})
}

func (b *BigDripleafHead) RecalculateCollisionBoxes() []math.AxisAlignedBB {
	if b.LeafState == blockutils.DripleafStateFullTilt {
		return nil
	}
	return []math.AxisAlignedBB{
		math.OneAABB().TrimmedCopy(math.Down, 11.0/16).TrimmedCopy(math.Up, b.getLeafTopOffset()),
	}
}
