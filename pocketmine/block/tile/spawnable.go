package tile

import "pocketmine-go/pocketmine/nbt"

// spawnableShaper lets SpawnableBase reach a concrete tile's SaveID/AddAdditionalSpawnData -
// same self-dispatch shape as nameableShaper.
type spawnableShaper interface {
	SaveID() string
	AddAdditionalSpawnData(nbt *nbt.CompoundTag)
}

// SpawnableBase is a port of pocketmine\block\tile\Spawnable. The PHP original caches the spawn
// compound as a CacheableNbt (a pre-encoded varint/little-endian byte buffer, for the network
// layer) - since network/mcpe/protocol isn't ported, this just caches the *nbt.CompoundTag
// itself, which is the part relevant to anything other than actually sending packets.
type SpawnableBase struct {
	TileBase

	spawnCompoundCache *nbt.CompoundTag
}

func (s *SpawnableBase) ClearSpawnCompoundCache() { s.spawnCompoundCache = nil }

func (s *SpawnableBase) GetSpawnCompound(self spawnableShaper) *nbt.CompoundTag {
	n := nbt.NewCompoundTag()
	n.SetString(TagID, nbt.StringTag(self.SaveID()))
	n.SetInt(TagX, nbt.IntTag(s.position.FloorX()))
	n.SetInt(TagY, nbt.IntTag(s.position.FloorY()))
	n.SetInt(TagZ, nbt.IntTag(s.position.FloorZ()))
	self.AddAdditionalSpawnData(n)
	return n
}

func (s *SpawnableBase) GetSerializedSpawnCompound(self spawnableShaper) *nbt.CompoundTag {
	if s.spawnCompoundCache == nil {
		s.spawnCompoundCache = s.GetSpawnCompound(self)
	}
	return s.spawnCompoundCache
}
