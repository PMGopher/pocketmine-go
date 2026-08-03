package leveldb

import (
	"fmt"

	"pocketmine-go/pocketmine/binaryutils"
	"pocketmine-go/pocketmine/data/bedrock"
	"pocketmine-go/pocketmine/nbt"
	"pocketmine-go/pocketmine/world/format"
)

// nbtMaxDepth mirrors the depth limit real PocketMine-MP's NBT reader defaults to for world data -
// block state compounds are only ever 2 levels deep (root -> states), so this is generous rather
// than exact.
const nbtMaxDepth = 512

// StateLookup resolves a chunk's bare internal state ID to the persistent (name + states) form a
// world save stores - the disk equivalent of convert.BlockStateSerializer. Supplied by the caller
// (World) rather than imported directly, since format/io/leveldb has no reason to depend on the
// block/convert packages beyond this one seam.
type StateLookup func(stateID int32) (bedrock.BlockStateData, error)

// StateResolver is the reverse of StateLookup: persistent (name + states) back to this port's own
// internal state ID. Returns ok=false for any blockstate this port doesn't have a registered
// template for (see World.stateTemplates) - the same "not supported yet" outcome
// BlockTranslator.InternalIDToNetworkID falls back on for the network direction.
type StateResolver func(data bedrock.BlockStateData) (int32, bool)

// SerializeSubChunk is a port of LevelDB::serializeBlockPalette plus the version/layer-count
// envelope from LevelDB::writeChunk's subchunk loop - the exact on-disk shape ChunkVersion::
// PALETTED_MULTI (8) has always used: identical to this port's own network SerializeSubChunk
// envelope (version byte, layer count, per-layer bitsPerBlock/words/palette-count), except each
// palette entry is a persistent NBT blockstate compound instead of a raw network runtime ID -
// worlds must stay loadable across game updates that change runtime ID assignments, so the save
// format never uses them.
func SerializeSubChunk(subChunk *format.SubChunk, lookup StateLookup) ([]byte, error) {
	layers := subChunk.GetBlockLayers()
	buf := []byte{subChunkVersion, byte(len(layers))}

	for _, layer := range layers {
		bitsPerBlock := layer.GetBitsPerBlock()
		buf = append(buf, byte(bitsPerBlock<<1))
		buf = append(buf, layer.GetWordArray()...)

		palette := layer.GetPalette()
		if bitsPerBlock != 0 {
			buf = append(buf, binaryutils.WriteLInt(int32(len(palette)))...)
		}

		roots := make([]*nbt.TreeRoot, 0, len(palette))
		for _, stateID := range palette {
			data, err := lookup(stateID)
			if err != nil {
				return nil, fmt.Errorf("leveldb: serializing palette entry (internal state %d): %w", stateID, err)
			}
			tag, err := blockStateToNBT(data)
			if err != nil {
				return nil, err
			}
			root, err := nbt.NewTreeRoot(tag, "")
			if err != nil {
				return nil, err
			}
			roots = append(roots, root)
		}
		encoded, err := nbt.NewLittleEndianSerializer().WriteMultiple(roots)
		if err != nil {
			return nil, fmt.Errorf("leveldb: encoding subchunk block palette: %w", err)
		}
		buf = append(buf, encoded...)
	}

	return buf, nil
}

// DeserializeSubChunkLayers is the reverse of SerializeSubChunk's palette encoding, returning just
// the block layers (not a full *format.SubChunk - the biome array lives under a separate LevelDB
// key, see chunk.go's LoadChunk, which combines both into the actual SubChunk). emptyBlockID is
// the internal state ID substituted for any block state this port doesn't recognise (see
// StateResolver's doc comment) - the same "honest, non-crashing fallback" this port uses elsewhere
// (BlockTranslator's fallback state, World.GetBlockAt's unknown-state handling), rather than
// failing the whole subchunk load over one unfamiliar block.
func DeserializeSubChunkLayers(data []byte, emptyBlockID int32, resolve StateResolver) ([]*format.PalettedBlockArray, error) {
	if len(data) < 2 {
		return nil, fmt.Errorf("leveldb: subchunk data too short (%d bytes)", len(data))
	}
	version := data[0]
	if version != subChunkVersion {
		return nil, fmt.Errorf("leveldb: unsupported subchunk version %d (only PALETTED_MULTI=%d is supported)", version, subChunkVersion)
	}
	layerCount := int(data[1])
	offset := 2

	layers := make([]*format.PalettedBlockArray, 0, layerCount)
	deserializer := nbt.NewLittleEndianSerializer()
	for i := 0; i < layerCount; i++ {
		if offset >= len(data) {
			return nil, fmt.Errorf("leveldb: subchunk data truncated before layer %d header", i)
		}
		bitsPerBlock := int(data[offset]) >> 1
		offset++

		var wordBytes []byte
		if bitsPerBlock != 0 {
			wordLen := format.WordCountForBitsPerBlock(bitsPerBlock) * 4
			if offset+wordLen > len(data) {
				return nil, fmt.Errorf("leveldb: subchunk data truncated in layer %d word array", i)
			}
			wordBytes = data[offset : offset+wordLen]
			offset += wordLen
		}

		paletteCount := 1
		if bitsPerBlock != 0 {
			if offset+4 > len(data) {
				return nil, fmt.Errorf("leveldb: subchunk data truncated before layer %d palette count", i)
			}
			n, err := binaryutils.ReadLInt(data[offset : offset+4])
			if err != nil {
				return nil, err
			}
			paletteCount = int(n)
			offset += 4
		}

		palette := make([]int32, paletteCount)
		for p := 0; p < paletteCount; p++ {
			root, newOffset, err := deserializer.Read(data, offset, nbtMaxDepth)
			if err != nil {
				return nil, fmt.Errorf("leveldb: decoding layer %d palette entry %d: %w", i, p, err)
			}
			offset = newOffset

			compound, ok := root.GetTag().(*nbt.CompoundTag)
			if !ok {
				return nil, fmt.Errorf("leveldb: layer %d palette entry %d is not a compound tag", i, p)
			}
			blockState, err := nbtToBlockState(compound)
			if err != nil {
				return nil, err
			}
			stateID, ok := resolve(blockState)
			if !ok {
				stateID = emptyBlockID
			}
			palette[p] = stateID
		}

		layer, err := format.NewPalettedBlockArrayFromRaw(bitsPerBlock, wordBytes, palette)
		if err != nil {
			return nil, fmt.Errorf("leveldb: rebuilding layer %d: %w", i, err)
		}
		layers = append(layers, layer)
	}

	return layers, nil
}
