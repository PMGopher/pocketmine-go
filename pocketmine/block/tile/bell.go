package tile

import (
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
)

const (
	BellTagDirection = "Direction"
	BellTagRinging   = "Ringing"
	BellTagTicks     = "Ticks"
)

// Bell is a port of pocketmine\block\tile\Bell.
type Bell struct {
	SpawnableBase

	Ringing bool
	Facing  int
	Ticks   int
}

func NewBell(world World, pos math.Vector3) *Bell {
	b := &Bell{SpawnableBase: SpawnableBase{TileBase: NewTileBase(world, pos)}, Facing: int(math.North)}
	b.Init(b)
	return b
}

func (b *Bell) SaveID() string { return "Bell" }

func (b *Bell) IsRinging() bool { return b.Ringing }

func (b *Bell) SetRinging(ringing bool) { b.Ringing = ringing }

func (b *Bell) GetFacing() int { return b.Facing }

func (b *Bell) SetFacing(facing int) { b.Facing = facing }

func (b *Bell) GetTicks() int { return b.Ticks }

func (b *Bell) SetTicks(ticks int) { b.Ticks = ticks }

func (b *Bell) AddAdditionalSpawnData(tag *nbt.CompoundTag) {
	ringing := nbt.ByteTag(0)
	if b.Ringing {
		ringing = 1
	}
	tag.SetByte(BellTagRinging, ringing)
	tag.SetInt(BellTagDirection, nbt.IntTag(b.Facing))
	tag.SetInt(BellTagTicks, nbt.IntTag(b.Ticks))
}

func (b *Bell) ReadSaveData(tag *nbt.CompoundTag) error {
	b.Ringing = tag.GetByteOr(BellTagRinging, 0) != 0
	b.Facing = int(tag.GetIntOr(BellTagDirection, nbt.IntTag(math.North)))
	b.Ticks = int(tag.GetIntOr(BellTagTicks, 0))
	return nil
}

func (b *Bell) WriteSaveData(tag *nbt.CompoundTag) {
	ringing := nbt.ByteTag(0)
	if b.Ringing {
		ringing = 1
	}
	tag.SetByte(BellTagRinging, ringing)
	tag.SetInt(BellTagDirection, nbt.IntTag(b.Facing))
	tag.SetInt(BellTagTicks, nbt.IntTag(b.Ticks))
}

// CreateFakeUpdateCompound is a port of Bell::createFakeUpdatePacket: the spawn compound of a
// ringing bell hit on bellHitFace, which the block sends as a BlockActorDataPacket (this package
// has no network code; see block.BroadcastTileDataFunc).
func (b *Bell) CreateFakeUpdateCompound(bellHitFace math.Facing) *nbt.CompoundTag {
	tag := b.GetSpawnCompound(b)
	tag.SetByte(BellTagRinging, 1)
	var direction int
	switch bellHitFace {
	case math.South:
		direction = 0
	case math.West:
		direction = 1
	case math.North:
		direction = 2
	case math.East:
		direction = 3
	default:
		panic("Unreachable")
	}
	tag.SetInt(BellTagDirection, nbt.IntTag(direction))
	tag.SetInt(BellTagTicks, 0)
	return tag
}
