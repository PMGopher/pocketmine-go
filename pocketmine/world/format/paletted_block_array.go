// Package format is a port of a slice of pocketmine\world\format: the in-memory chunk/subchunk
// block storage this port's own block package (via block.Behavior.GetStateId()) and the network
// chunk serializer both read and write.
package format

import (
	"encoding/binary"
	"fmt"
)

const (
	subChunkEdgeLength = 16
	subChunkBlockCount = subChunkEdgeLength * subChunkEdgeLength * subChunkEdgeLength // 4096
)

// validBitsPerBlock lists every bits-per-block step a paletted array may use - Bedrock's paletted
// storage format doesn't allow arbitrary bit widths, only these.
var validBitsPerBlock = [...]int{0, 1, 2, 3, 4, 5, 6, 8, 16}

// PalettedBlockArray is a hand-written port of pocketmine\world\format\PalettedBlockArray. There
// is no PHP source to port from - the real class is implemented natively (the chunkutils2 PHP
// extension); tests/phpstan/stubs/chunkutils2.stub only documents a couple of static-analysis-only
// methods, not the actual algorithm. This implements the well-documented Bedrock paletted storage
// format directly instead: up to 4096 (16x16x16) entries in "XZY order" (matching
// pocketmine\world\format\Chunk's own doc comment), each storing a PALETTE INDEX (not the raw
// value) packed into bitsPerBlock bits within a little-endian uint32 word array, with a "uniform"
// fast path (bitsPerBlock 0, no words at all) for a layer that's a single repeated value - most
// importantly, a subchunk layer that's 100% air.
type PalettedBlockArray struct {
	bitsPerBlock int
	palette      []int32
	// index is a reverse lookup (value -> index in palette), rebuilt whenever the palette changes,
	// so Set doesn't need to linear-scan the palette on every call.
	index map[int32]int
	// words holds every position's PALETTE INDEX, bitsPerBlock bits each, packed LSB-first into
	// little-endian uint32 words - nil when bitsPerBlock is 0.
	words []uint32
}

// blockIndex computes a position's index within a 16x16x16 layer ("XZY order").
func blockIndex(x, y, z int) int {
	return (x << 8) | (z << 4) | y
}

// NewPalettedBlockArray returns a new uniform (single-value) array, matching
// `new PalettedBlockArray($fillValue)`.
func NewPalettedBlockArray(fillValue int32) *PalettedBlockArray {
	return &PalettedBlockArray{
		bitsPerBlock: 0,
		palette:      []int32{fillValue},
		index:        map[int32]int{fillValue: 0},
	}
}

func (p *PalettedBlockArray) GetBitsPerBlock() int { return p.bitsPerBlock }

// Clone is a port of PalettedBlockArray's implicit PHP `clone` semantics (native extension objects
// are copied by value on `clone` too) - a deep copy, since palette/index/words are all reference
// types in Go.
func (p *PalettedBlockArray) Clone() *PalettedBlockArray {
	c := &PalettedBlockArray{
		bitsPerBlock: p.bitsPerBlock,
		palette:      make([]int32, len(p.palette)),
		index:        make(map[int32]int, len(p.index)),
		words:        make([]uint32, len(p.words)),
	}
	copy(c.palette, p.palette)
	for k, v := range p.index {
		c.index[k] = v
	}
	copy(c.words, p.words)
	return c
}

// GetPalette is a port of PalettedBlockArray::getPalette. Returns a copy - callers must not be
// able to corrupt the array's internal palette/index invariant by mutating the result.
func (p *PalettedBlockArray) GetPalette() []int32 {
	result := make([]int32, len(p.palette))
	copy(result, p.palette)
	return result
}

// GetWordArray returns the packed word array as little-endian bytes, matching what
// ChunkSerializer::serializeSubChunk writes via `$stream->writeByteArray($blocks->getWordArray())`
// (PalettedBlockArray::getWordArray returns a raw byte string in the PHP original).
func (p *PalettedBlockArray) GetWordArray() []byte {
	buf := make([]byte, len(p.words)*4)
	for i, w := range p.words {
		binary.LittleEndian.PutUint32(buf[i*4:], w)
	}
	return buf
}

// Get is a port of PalettedBlockArray::get.
func (p *PalettedBlockArray) Get(x, y, z int) int32 {
	if p.bitsPerBlock == 0 {
		return p.palette[0]
	}
	return p.palette[p.paletteIndexAt(blockIndex(x, y, z))]
}

func (p *PalettedBlockArray) paletteIndexAt(i int) int {
	blocksPerWord := 32 / p.bitsPerBlock
	wordIndex := i / blocksPerWord
	offset := uint((i % blocksPerWord) * p.bitsPerBlock)
	mask := uint32(1)<<uint(p.bitsPerBlock) - 1
	return int((p.words[wordIndex] >> offset) & mask)
}

// Set is a port of PalettedBlockArray::set.
func (p *PalettedBlockArray) Set(x, y, z int, value int32) {
	idx, ok := p.index[value]
	if !ok {
		idx = len(p.palette)
		p.palette = append(p.palette, value)
		p.index[value] = idx
		if needed := bitsPerBlockFor(len(p.palette)); needed != p.bitsPerBlock {
			p.repack(needed)
		}
	}
	p.setPaletteIndexAt(blockIndex(x, y, z), idx)
}

func (p *PalettedBlockArray) setPaletteIndexAt(i, paletteIdx int) {
	if p.bitsPerBlock == 0 {
		return // only one palette entry is possible; nothing to store
	}
	blocksPerWord := 32 / p.bitsPerBlock
	wordIndex := i / blocksPerWord
	offset := uint((i % blocksPerWord) * p.bitsPerBlock)
	mask := uint32(1)<<uint(p.bitsPerBlock) - 1
	p.words[wordIndex] = (p.words[wordIndex] &^ (mask << offset)) | (uint32(paletteIdx) << offset)
}

// bitsPerBlockFor returns the smallest valid bitsPerBlock step that can index paletteSize
// distinct palette entries.
func bitsPerBlockFor(paletteSize int) int {
	if paletteSize <= 1 {
		return 0
	}
	for _, bits := range validBitsPerBlock[1:] {
		if paletteSize <= 1<<uint(bits) {
			return bits
		}
	}
	panic(fmt.Sprintf("format: palette of size %d exceeds the maximum bitsPerBlock (16)", paletteSize))
}

func wordCountFor(bitsPerBlock int) int {
	blocksPerWord := 32 / bitsPerBlock
	return (subChunkBlockCount + blocksPerWord - 1) / blocksPerWord
}

// repack rebuilds the word array at a new bitsPerBlock, decoding every existing entry at the old
// width and re-encoding it at the new one.
func (p *PalettedBlockArray) repack(newBitsPerBlock int) {
	old := p.words
	oldBits := p.bitsPerBlock

	p.bitsPerBlock = newBitsPerBlock
	if newBitsPerBlock == 0 {
		p.words = nil
		return
	}
	p.words = make([]uint32, wordCountFor(newBitsPerBlock))

	if oldBits == 0 {
		return // every position was (implicitly) palette index 0, which the zeroed words already mean
	}
	oldBlocksPerWord := 32 / oldBits
	oldMask := uint32(1)<<uint(oldBits) - 1
	for i := 0; i < subChunkBlockCount; i++ {
		oldWordIndex := i / oldBlocksPerWord
		oldOffset := uint((i % oldBlocksPerWord) * oldBits)
		paletteIdx := int((old[oldWordIndex] >> oldOffset) & oldMask)
		p.setPaletteIndexAt(i, paletteIdx)
	}
}

// CollectGarbage rebuilds the palette to contain only entries some position actually still uses,
// shrinking bitsPerBlock accordingly - matching the native chunkutils2 extension's
// PalettedBlockArray::collectGarbage, which SubChunk::collectGarbage relies on to decide whether a
// layer has become redundant (collapsed to a single, empty-matching value) and can be dropped.
func (p *PalettedBlockArray) CollectGarbage() {
	if p.bitsPerBlock == 0 {
		return // already maximally compact: one entry, no words
	}

	indices := make([]int, subChunkBlockCount)
	used := make(map[int32]bool, len(p.palette))
	for i := 0; i < subChunkBlockCount; i++ {
		idx := p.paletteIndexAt(i)
		indices[i] = idx
		used[p.palette[idx]] = true
	}

	newPalette := make([]int32, 0, len(used))
	newIndex := make(map[int32]int, len(used))
	remap := make([]int, len(p.palette)) // old palette index -> new palette index
	for oldIdx, value := range p.palette {
		if !used[value] {
			continue
		}
		newIdx := len(newPalette)
		newPalette = append(newPalette, value)
		newIndex[value] = newIdx
		remap[oldIdx] = newIdx
	}

	p.palette = newPalette
	p.index = newIndex
	p.bitsPerBlock = bitsPerBlockFor(len(newPalette))
	if p.bitsPerBlock == 0 {
		p.words = nil
		return
	}
	p.words = make([]uint32, wordCountFor(p.bitsPerBlock))
	for i := 0; i < subChunkBlockCount; i++ {
		p.setPaletteIndexAt(i, remap[indices[i]])
	}
}
