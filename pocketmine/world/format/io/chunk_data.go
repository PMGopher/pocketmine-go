package io

import (
	"pocketmine-go/pocketmine/nbt"
	"pocketmine-go/pocketmine/world/format"
)

// ChunkData is a port of pocketmine\world\format\io\ChunkData: a chunk as loaded from or saved to
// a provider (terrain, and the NBT of its entities and tiles).
type ChunkData struct {
	subChunks map[int]*format.SubChunk
	populated bool
	entityNBT []*nbt.CompoundTag
	tileNBT   []*nbt.CompoundTag
}

// NewChunkData is a port of ChunkData::__construct. subChunks is keyed by subchunk Y index.
func NewChunkData(subChunks map[int]*format.SubChunk, populated bool, entityNBT, tileNBT []*nbt.CompoundTag) *ChunkData {
	return &ChunkData{subChunks: subChunks, populated: populated, entityNBT: entityNBT, tileNBT: tileNBT}
}

func (d *ChunkData) GetSubChunks() map[int]*format.SubChunk { return d.subChunks }
func (d *ChunkData) IsPopulated() bool                      { return d.populated }
func (d *ChunkData) GetEntityNBT() []*nbt.CompoundTag       { return d.entityNBT }
func (d *ChunkData) GetTileNBT() []*nbt.CompoundTag         { return d.tileNBT }
