package tile

import "pocketmine-go/pocketmine/nbt"

// spawnableShaper lets SpawnableBase reach a concrete tile's SaveID/AddAdditionalSpawnData -
// same self-dispatch shape as nameableShaper.
type spawnableShaper interface {
	SaveID() string
	AddAdditionalSpawnData(nbt *nbt.CompoundTag)
}

// SpawnableBase is a port of pocketmine\block\tile\Spawnable. The PHP original caches the spawn
// compound as a CacheableNbt (pre-encoded network NBT); this caches the *nbt.CompoundTag, which
// the network code encodes with gophertunnel when it sends it.
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

// Spawnable is pocketmine\block\tile\Spawnable as an interface: every tile embedding SpawnableBase
// satisfies it.
type Spawnable interface {
	Tile
	spawnableShaper
	ClearSpawnCompoundCache()
	GetSerializedSpawnCompound(self spawnableShaper) *nbt.CompoundTag
}

// SerializedSpawnCompound is `$tile instanceof Spawnable ? $tile->getSerializedSpawnCompound()`:
// the compound sent to clients to spawn the tile (cached until ClearSpawnCompoundCache).
func SerializedSpawnCompound(t Tile) (*nbt.CompoundTag, bool) {
	s, ok := t.(Spawnable)
	if !ok {
		return nil, false
	}
	return s.GetSerializedSpawnCompound(s), true
}

// RenderUpdateBugWorkaroundStateProperties is a port of
// Spawnable::getRenderUpdateBugWorkaroundStateProperties: the fake block state properties (network
// state values: uint8 for bytes, int32 for ints) sent before the real state so that the client
// re-renders the block with the tile's new data. Empty for most tiles.
func RenderUpdateBugWorkaroundStateProperties(t Tile, b Block) map[string]any {
	if w, ok := t.(interface {
		GetRenderUpdateBugWorkaroundStateProperties(b Block) map[string]any
	}); ok {
		return w.GetRenderUpdateBugWorkaroundStateProperties(b)
	}
	return nil
}
