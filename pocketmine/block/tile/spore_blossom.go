package tile

import (
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
)

// SporeBlossom is a port of pocketmine\block\tile\SporeBlossom.
//
// This exists to force the client to update the spore blossom every tick, which is necessary for
// it to generate particles.
type SporeBlossom struct {
	SpawnableBase
}

func NewSporeBlossom(world World, pos math.Vector3) *SporeBlossom {
	s := &SporeBlossom{SpawnableBase{TileBase: NewTileBase(world, pos)}}
	s.Init(s)
	return s
}

func (s *SporeBlossom) SaveID() string { return "SporeBlossom" }

func (s *SporeBlossom) AddAdditionalSpawnData(nbt *nbt.CompoundTag) {}

func (s *SporeBlossom) ReadSaveData(nbt *nbt.CompoundTag) error { return nil }

func (s *SporeBlossom) WriteSaveData(nbt *nbt.CompoundTag) {}
