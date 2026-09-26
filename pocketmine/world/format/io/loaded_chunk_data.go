package io

// LoadedChunkData fixer flags, a port of LoadedChunkData::FIXER_FLAG_*.
const (
	FixerFlagNone = 0
	FixerFlagAll  = ^0
)

// LoadedChunkData is a port of pocketmine\world\format\io\LoadedChunkData: a chunk read from a
// provider, and whether it was upgraded from an older format (and so needs saving).
type LoadedChunkData struct {
	data       *ChunkData
	upgraded   bool
	fixerFlags int
}

// NewLoadedChunkData is a port of LoadedChunkData::__construct.
func NewLoadedChunkData(data *ChunkData, upgraded bool, fixerFlags int) *LoadedChunkData {
	return &LoadedChunkData{data: data, upgraded: upgraded, fixerFlags: fixerFlags}
}

func (d *LoadedChunkData) GetData() *ChunkData { return d.data }
func (d *LoadedChunkData) IsUpgraded() bool    { return d.upgraded }
func (d *LoadedChunkData) GetFixerFlags() int  { return d.fixerFlags }
