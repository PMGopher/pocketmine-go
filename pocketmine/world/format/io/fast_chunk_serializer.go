package io

import (
	"encoding/binary"
	"fmt"

	"pocketmine-go/pocketmine/world/format"
)

// fastChunkFlagPopulated is FastChunkSerializer::FLAG_POPULATED.
const fastChunkFlagPopulated = 1 << 1

// SerializeTerrain is a port of FastChunkSerializer::serializeTerrain: a chunk's terrain in a fast
// in-memory format for passing between threads (tiles and entities aren't included, like PHP).
func SerializeTerrain(chunk *format.Chunk) []byte {
	var buf []byte
	flags := byte(0)
	if chunk.IsPopulated() {
		flags |= fastChunkFlagPopulated
	}
	buf = append(buf, flags)

	subChunks := chunk.GetSubChunks()
	buf = append(buf, byte(len(subChunks)))
	for y := format.MinSubChunkIndex; y <= format.MaxSubChunkIndex; y++ {
		subChunk, ok := subChunks[y]
		if !ok {
			continue
		}
		buf = append(buf, byte(int8(y)))
		buf = binary.BigEndian.AppendUint32(buf, uint32(subChunk.GetEmptyBlockID()))
		layers := subChunk.GetBlockLayers()
		buf = append(buf, byte(len(layers)))
		for _, blocks := range layers {
			buf = serializePalettedArray(buf, blocks)
		}
		buf = serializePalettedArray(buf, subChunk.GetBiomeArray())
	}
	return buf
}

func serializePalettedArray(buf []byte, array *format.PalettedBlockArray) []byte {
	buf = append(buf, byte(array.GetBitsPerBlock()))
	buf = append(buf, array.GetWordArray()...)
	palette := array.GetPalette()
	buf = binary.BigEndian.AppendUint32(buf, uint32(len(palette)*4))
	for _, v := range palette {
		buf = binary.LittleEndian.AppendUint32(buf, uint32(v)) // pack("L*"): machine (little) endian
	}
	return buf
}

type fastReader struct {
	data   []byte
	offset int
}

func (r *fastReader) bytes(n int) ([]byte, error) {
	if r.offset+n > len(r.data) {
		return nil, fmt.Errorf("unexpected end of fast-serialized chunk data")
	}
	b := r.data[r.offset : r.offset+n]
	r.offset += n
	return b, nil
}

func (r *fastReader) byte() (byte, error) {
	b, err := r.bytes(1)
	if err != nil {
		return 0, err
	}
	return b[0], nil
}

func (r *fastReader) uint32() (uint32, error) {
	b, err := r.bytes(4)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32(b), nil
}

func (r *fastReader) palettedArray() (*format.PalettedBlockArray, error) {
	bitsPerBlock, err := r.byte()
	if err != nil {
		return nil, err
	}
	words, err := r.bytes(format.WordCountForBitsPerBlock(int(bitsPerBlock)) * 4)
	if err != nil {
		return nil, err
	}
	paletteSize, err := r.uint32()
	if err != nil {
		return nil, err
	}
	raw, err := r.bytes(int(paletteSize))
	if err != nil {
		return nil, err
	}
	palette := make([]int32, len(raw)/4)
	for i := range palette {
		palette[i] = int32(binary.LittleEndian.Uint32(raw[i*4:]))
	}
	return format.NewPalettedBlockArrayFromRaw(int(bitsPerBlock), words, palette)
}

// DeserializeTerrain is a port of FastChunkSerializer::deserializeTerrain.
func DeserializeTerrain(data []byte) (*format.Chunk, error) {
	r := &fastReader{data: data}
	flags, err := r.byte()
	if err != nil {
		return nil, err
	}
	terrainPopulated := flags&fastChunkFlagPopulated != 0

	subChunks := map[int]*format.SubChunk{}
	count, err := r.byte()
	if err != nil {
		return nil, err
	}
	for i := 0; i < int(count); i++ {
		yByte, err := r.byte()
		if err != nil {
			return nil, err
		}
		//TODO: why the heck are we using big-endian here?
		airBlockID, err := r.uint32()
		if err != nil {
			return nil, err
		}
		layerCount, err := r.byte()
		if err != nil {
			return nil, err
		}
		layers := make([]*format.PalettedBlockArray, 0, layerCount)
		for j := 0; j < int(layerCount); j++ {
			layer, err := r.palettedArray()
			if err != nil {
				return nil, err
			}
			layers = append(layers, layer)
		}
		biomeArray, err := r.palettedArray()
		if err != nil {
			return nil, err
		}
		subChunks[int(int8(yByte))] = format.NewSubChunk(int32(airBlockID), layers, biomeArray)
	}
	return format.NewChunk(subChunks, terrainPopulated, EmptyStateID(), 0), nil
}
