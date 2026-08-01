package tile

import (
	"errors"
	stdmath "math"

	blockutils "pocketmine-go/pocketmine/block/utils"
	pmmath "pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
)

const (
	mobHeadTagSkullType = "SkullType"
	mobHeadTagRotation  = "Rotation"
	mobHeadTagRot       = "Rot"
)

var errInvalidSkullType = errors.New("invalid skull type tag value")

// MobHead is a port of pocketmine\block\tile\MobHead.
//
// Deprecated in the PHP original too - see block.MobHead (not ported yet).
//
// MobHeadTypeIdMap's byte<->MobHeadType mapping is a straightforward 0-6 sequential assignment
// matching declaration order (confirmed against the PHP source), so this uses int(MobHeadType)
// directly rather than porting a whole IntSaveIdMapTrait-based registry for one map.
type MobHead struct {
	SpawnableBase

	MobHeadType blockutils.MobHeadType
	Rotation    int
}

func NewMobHead(world World, pos pmmath.Vector3) *MobHead {
	m := &MobHead{SpawnableBase: SpawnableBase{TileBase: NewTileBase(world, pos)}, MobHeadType: blockutils.MobHeadTypeSkeleton}
	m.Init(m)
	return m
}

func (m *MobHead) SaveID() string { return "Skull" }

// ReadSaveData is a port of MobHead::readSaveData. The legacy "Rot" (0-15 byte) tag is used as a
// fallback if "Rotation" (yaw in degrees) isn't present, matching the PHP original's tag-presence
// check rather than just falling back on a zero value.
func (m *MobHead) ReadSaveData(tag *nbt.CompoundTag) error {
	if skullType, err := tag.GetByte(mobHeadTagSkullType); err == nil {
		if int(skullType) < 0 || int(skullType) > int(blockutils.MobHeadTypePiglin) {
			return errInvalidSkullType
		}
		m.MobHeadType = blockutils.MobHeadType(skullType)
	}

	if yaw, err := tag.GetFloat(mobHeadTagRotation); err == nil {
		m.Rotation = int(stdmath.Floor(float64(yaw)*16/360+0.5)) & 0xf
	} else if rot, err := tag.GetByte(mobHeadTagRot); err == nil && rot >= 0 && rot <= 15 {
		m.Rotation = int(rot)
	}
	return nil
}

func (m *MobHead) WriteSaveData(tag *nbt.CompoundTag) {
	tag.SetByte(mobHeadTagSkullType, nbt.ByteTag(m.MobHeadType))
	tag.SetFloat(mobHeadTagRotation, nbt.FloatTag(float64(m.Rotation)*360.0/16.0))
}

func (m *MobHead) AddAdditionalSpawnData(tag *nbt.CompoundTag) {
	tag.SetByte(mobHeadTagSkullType, nbt.ByteTag(m.MobHeadType))
	tag.SetFloat(mobHeadTagRotation, nbt.FloatTag(float64(m.Rotation)*360.0/16.0))
}

func (m *MobHead) GetMobHeadType() blockutils.MobHeadType { return m.MobHeadType }

func (m *MobHead) SetMobHeadType(t blockutils.MobHeadType) { m.MobHeadType = t }

func (m *MobHead) GetRotation() int { return m.Rotation }

func (m *MobHead) SetRotation(rotation int) { m.Rotation = rotation }
