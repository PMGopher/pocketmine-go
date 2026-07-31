package block

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// Lantern is a port of pocketmine\block\Lantern.
type Lantern struct {
	Transparent

	LightLevelValue int // readonly
	Hanging         bool
}

func NewLantern(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo, lightLevel int) *Lantern {
	l := &Lantern{Transparent: Transparent{NewBlock(idInfo, name, typeInfo)}, LightLevelValue: lightLevel}
	l.Init(l)
	return l
}

func (l *Lantern) Clone() Behavior {
	c := *l
	c.rebind(&c)
	return &c
}

func (l *Lantern) DescribeBlockOnlyState(w runtime.DataDescriber) { w.Bool(&l.Hanging) }

func (l *Lantern) IsHanging() bool { return l.Hanging }

func (l *Lantern) SetHanging(hanging bool) { l.Hanging = hanging }

func (l *Lantern) GetLightLevel() int { return l.LightLevelValue }

func (l *Lantern) RecalculateCollisionBoxes() []math.AxisAlignedBB {
	upTrim := 8.0 / 16
	downTrim := 0.0
	if l.Hanging {
		upTrim = 6.0 / 16
		downTrim = 2.0 / 16
	}
	bb := math.OneAABB().
		TrimmedCopy(math.Up, upTrim).
		TrimmedCopy(math.Down, downTrim).
		SquashedCopy(math.AxisX, 5.0/16).
		SquashedCopy(math.AxisZ, 5.0/16)
	return []math.AxisAlignedBB{bb}
}

func (l *Lantern) GetSupportType(facing math.Facing) blockutils.SupportType {
	return blockutils.SupportTypeNone
}

func (l *Lantern) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	downSupport := l.canBeSupportedAt(blockReplace, math.Down)
	if !downSupport && !l.canBeSupportedAt(blockReplace, math.Up) {
		return false
	}
	l.Hanging = face == math.Down || !downSupport
	return l.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
}

func (l *Lantern) OnNearbyBlockChange() {
	face := math.Down
	if l.Hanging {
		face = math.Up
	}
	if !l.canBeSupportedAt(l.self, face) {
		if world, err := l.position.GetWorld(); err == nil {
			world.UseBreakOn(l.position.AsVector3())
		}
	}
}

func (l *Lantern) canBeSupportedAt(blk Behavior, face math.Facing) bool {
	return blk.(blockGeometry).GetAdjacentSupportType(face).HasCenterSupport()
}
