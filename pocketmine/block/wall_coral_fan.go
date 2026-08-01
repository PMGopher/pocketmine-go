package block

import (
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// WallCoralFan is a port of pocketmine\block\WallCoralFan.
type WallCoralFan struct {
	BaseCoral
	HorizontalFacingComponent
}

func NewWallCoralFan(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *WallCoralFan {
	w := &WallCoralFan{
		BaseCoral:                 BaseCoral{Transparent: Transparent{NewBlock(idInfo, name, typeInfo)}},
		HorizontalFacingComponent: NewHorizontalFacingComponent(),
	}
	w.Init(w)
	return w
}

func (w *WallCoralFan) Clone() Behavior {
	c := *w
	c.rebind(&c)
	return &c
}

func (w *WallCoralFan) DescribeBlockOnlyState(d runtime.DataDescriber) { w.DescribeHorizontalFacing(d) }

func (w *WallCoralFan) wallCanBeSupportedAt(blk Behavior, face math.Facing) bool {
	return blk.(blockGeometry).GetAdjacentSupportType(face).HasCenterSupport()
}

func (w *WallCoralFan) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	axis := math.FacingAxis(face)
	if (axis != math.AxisX && axis != math.AxisZ) || !w.wallCanBeSupportedAt(blockReplace, math.Opposite(face)) {
		return false
	}
	w.Facing = face
	w.Dead = !w.isCoveredWithWater()
	return w.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
}

func (w *WallCoralFan) OnNearbyBlockChange() {
	if !w.wallCanBeSupportedAt(w.self, math.Opposite(w.Facing)) {
		if world, err := w.position.GetWorld(); err == nil {
			world.UseBreakOn(w.position.AsVector3())
		}
	} else {
		w.BaseCoral.OnNearbyBlockChange()
	}
}

// AsItem should return VanillaItems.CORAL_FAN().SetCoralType(w.CoralType).SetDead(w.Dead) — needs
// the unported item package (see Block.GetDropsForCompatibleTool's doc comment), so it's left as
// Block's default for now.
