package format

// HeightArray is a port of pocketmine\world\format\HeightArray: a 16x16 (one Y value per chunk
// column) heightmap, used by the light engine to know the topmost sky-light-blocking block in
// each column without rescanning the whole column every time.
type HeightArray struct {
	values [256]int
}

// NewHeightArrayFilled is a port of `HeightArray::fill`.
func NewHeightArrayFilled(value int) *HeightArray {
	h := &HeightArray{}
	for i := range h.values {
		h.values[i] = value
	}
	return h
}

// heightIndex is a port of HeightArray::idx (ZZZZXXXX bit order). Panics on an out-of-range x/z,
// matching the PHP original's InvalidArgumentException (a programmer error at the call site).
func heightIndex(x, z int) int {
	if x < 0 || x >= SubChunkEdgeLength || z < 0 || z >= SubChunkEdgeLength {
		panic("format: x and z must be in the range 0-15")
	}
	return (z << 4) | x
}

// Get is a port of HeightArray::get.
func (h *HeightArray) Get(x, z int) int { return h.values[heightIndex(x, z)] }

// Set is a port of HeightArray::set.
func (h *HeightArray) Set(x, z, height int) { h.values[heightIndex(x, z)] = height }

// GetValues is a port of HeightArray::getValues (ZZZZXXXX key bit order, matching heightIndex).
func (h *HeightArray) GetValues() [256]int { return h.values }

// SetValues replaces every value at once - used when rebuilding the whole heightmap
// (SkyLightUpdate::recalculateChunk's `$chunk->setHeightMapArray($newHeightMap->getValues())`).
func (h *HeightArray) SetValues(values [256]int) { h.values = values }

// Clone is a port of HeightArray's implicit PHP `clone` (deep-copies the backing SplFixedArray).
func (h *HeightArray) Clone() *HeightArray {
	c := &HeightArray{}
	c.values = h.values
	return c
}
