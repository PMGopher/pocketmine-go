package block

import (
	stdmath "math"

	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// FloorCoralFan is a port of pocketmine\block\FloorCoralFan.
type FloorCoralFan struct {
	BaseCoral

	Axis math.Axis
}

func NewFloorCoralFan(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *FloorCoralFan {
	f := &FloorCoralFan{BaseCoral: BaseCoral{Transparent: Transparent{NewBlock(idInfo, name, typeInfo)}}, Axis: math.AxisX}
	f.Init(f)
	return f
}

func (f *FloorCoralFan) Clone() Behavior {
	c := *f
	c.rebind(&c)
	return &c
}

func (f *FloorCoralFan) DescribeBlockOnlyState(w runtime.DataDescriber) { w.HorizontalAxis(&f.Axis) }

func (f *FloorCoralFan) GetAxis() math.Axis { return f.Axis }

// SetAxis panics for a non-horizontal axis, mirroring the PHP original's InvalidArgumentException
// (a programmer error at the call site).
func (f *FloorCoralFan) SetAxis(axis math.Axis) {
	if axis != math.AxisX && axis != math.AxisZ {
		panic("Axis must be X or Z only")
	}
	f.Axis = axis
}

func (f *FloorCoralFan) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	if player != nil {
		playerBlockPos := player.GetPosition().Floor()
		direction := blockReplace.GetPosition().AsVector3().SubtractVector(playerBlockPos).Normalize()
		angle := stdmath.Atan2(direction.Z, direction.X) * 180 / stdmath.Pi

		// TODO: this produces Z axis 75% of the time, because any negative angle will produce Z
		// axis - this is a bug in vanilla (matching the PHP original's own note).
		// https://bugs.mojang.com/browse/MCPE-125311
		if angle <= 45 || 315 <= angle || (135 <= angle && angle <= 225) {
			f.Axis = math.AxisZ
		}
	}

	f.Dead = !f.isCoveredWithWater()

	return f.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
}

func (f *FloorCoralFan) canBeSupportedAt(blk Behavior) bool {
	return blk.(blockGeometry).GetAdjacentSupportType(math.Down).HasCenterSupport()
}

func (f *FloorCoralFan) CanBePlacedAt(blockReplace Behavior, clickVector math.Vector3, face math.Facing, isClickedBlock bool) bool {
	return f.canBeSupportedAt(blockReplace) && f.Transparent.CanBePlacedAt(blockReplace, clickVector, face, isClickedBlock)
}

func (f *FloorCoralFan) OnNearbyBlockChange() {
	if !f.canBeSupportedAt(f.self) {
		if world, err := f.position.GetWorld(); err == nil {
			world.UseBreakOn(f.position.AsVector3())
		}
	} else {
		f.BaseCoral.OnNearbyBlockChange()
	}
}

// AsItem should return VanillaItems.CORAL_FAN().SetCoralType(f.CoralType).SetDead(f.Dead) — needs
// the unported item package (see Block.GetDropsForCompatibleTool's doc comment), so it's left as
// Block's default for now.
