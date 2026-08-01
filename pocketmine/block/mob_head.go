package block

import (
	stdmath "math"

	"pocketmine-go/pocketmine/block/tile"
	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

const (
	MobHeadMinRotation = 0
	MobHeadMaxRotation = 15
)

// MobHead is a port of pocketmine\block\MobHead.
//
// WriteStateToWorld's tile sync (writing Rotation/MobHeadType back to the tile.MobHead on
// placement) is skipped: there's no WriteStateToWorld hook in Behavior yet - same gap already
// documented on block.Note/RedstoneComparator/BaseBanner.
type MobHead struct {
	Flowable

	MobHeadType blockutils.MobHeadType
	Facing      math.Facing
	Rotation    int
}

func NewMobHead(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *MobHead {
	m := &MobHead{
		Flowable:    Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}},
		MobHeadType: blockutils.MobHeadTypeSkeleton,
		Facing:      math.North,
	}
	m.Init(m)
	return m
}

func (m *MobHead) Clone() Behavior {
	c := *m
	c.rebind(&c)
	return &c
}

func (m *MobHead) DescribeBlockItemState(w runtime.DataDescriber) {
	t := int(m.MobHeadType)
	w.BoundedIntAuto(int(blockutils.MobHeadTypeSkeleton), int(blockutils.MobHeadTypePiglin), &t)
	m.MobHeadType = blockutils.MobHeadType(t)
}

func (m *MobHead) DescribeBlockOnlyState(w runtime.DataDescriber) {
	w.FacingExcept(&m.Facing, math.Down)
}

// ReadStateFromWorld is a port of MobHead::readStateFromWorld.
func (m *MobHead) ReadStateFromWorld() Behavior {
	m.Block.ReadStateFromWorld()

	world, err := m.position.GetWorld()
	if err != nil {
		return m.self
	}
	t, _ := world.GetTile(m.position)
	if headTile, ok := t.(*tile.MobHead); ok {
		m.MobHeadType = headTile.GetMobHeadType()
		m.Rotation = headTile.GetRotation()
	}
	return m.self
}

func (m *MobHead) GetMobHeadType() blockutils.MobHeadType { return m.MobHeadType }

func (m *MobHead) SetMobHeadType(t blockutils.MobHeadType) { m.MobHeadType = t }

func (m *MobHead) GetFacing() math.Facing { return m.Facing }

// SetFacing panics if facing is Down, mirroring the PHP original's InvalidArgumentException (a
// programmer error at the call site).
func (m *MobHead) SetFacing(facing math.Facing) {
	if facing == math.Down {
		panic("Skull may not face DOWN")
	}
	m.Facing = facing
}

func (m *MobHead) GetRotation() int { return m.Rotation }

// SetRotation panics if out of range, mirroring the PHP original's InvalidArgumentException.
func (m *MobHead) SetRotation(rotation int) {
	if rotation < MobHeadMinRotation || rotation > MobHeadMaxRotation {
		panic("Rotation must be in range 0 ... 15")
	}
	m.Rotation = rotation
}

func (m *MobHead) RecalculateCollisionBoxes() []math.AxisAlignedBB {
	box := math.OneAABB().ContractedCopy(0.25, 0, 0.25).TrimmedCopy(math.Up, 0.5)
	if m.Facing != math.Up {
		box = box.OffsetTowardsCopy(math.Opposite(m.Facing), 0.25).OffsetTowardsCopy(math.Up, 0.25)
	}
	return []math.AxisAlignedBB{box}
}

func (m *MobHead) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	if face == math.Down {
		return false
	}
	m.Facing = face
	if player != nil && face == math.Up {
		m.Rotation = int(stdmath.Floor(player.GetYaw()*16/360+0.5)) & 0xf
	}
	return m.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
}
