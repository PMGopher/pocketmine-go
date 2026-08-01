package block

import (
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// RedstoneWire is a port of pocketmine\block\RedstoneWire.
type RedstoneWire struct {
	Flowable
	AnalogRedstoneSignalEmitterComponent
}

func NewRedstoneWire(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *RedstoneWire {
	r := &RedstoneWire{Flowable: Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}}}
	r.Init(r)
	return r
}

func (r *RedstoneWire) Clone() Behavior {
	c := *r
	c.rebind(&c)
	return &c
}

func (r *RedstoneWire) DescribeBlockOnlyState(w runtime.DataDescriber) { r.DescribeSignalStrength(w) }

// ReadStateFromWorld: connections to nearby redstone components are a TODO in the PHP original
// too.
func (r *RedstoneWire) ReadStateFromWorld() Behavior {
	r.Block.ReadStateFromWorld()
	return r.self
}

func (r *RedstoneWire) canBeSupportedAt(blk Behavior) bool {
	return blk.(blockGeometry).GetAdjacentSupportType(math.Down).HasCenterSupport()
}

func (r *RedstoneWire) CanBePlacedAt(blockReplace Behavior, clickVector math.Vector3, face math.Facing, isClickedBlock bool) bool {
	return r.canBeSupportedAt(blockReplace) && r.Flowable.CanBePlacedAt(blockReplace, clickVector, face, isClickedBlock)
}

func (r *RedstoneWire) OnNearbyBlockChange() {
	if !r.canBeSupportedAt(r.self) {
		if world, err := r.position.GetWorld(); err == nil {
			world.UseBreakOn(r.position.AsVector3())
		}
	} else {
		r.Flowable.OnNearbyBlockChange()
	}
}

// GetDropsForCompatibleTool/AsItem should return VanillaItems.REDSTONE_DUST() — needs the
// unported item package (see Block.GetDropsForCompatibleTool's doc comment), so both are left as
// Block's defaults for now.
