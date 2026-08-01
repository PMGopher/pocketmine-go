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
//
// createFakeUpdatePacket (the ring-animation hack) needs BlockActorDataPacket, from the unported
// network/mcpe/protocol, so it's not ported here - see block.Bell.Ring's doc comment for the
// block-side half of the same gap.
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
