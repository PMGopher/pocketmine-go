package format

// LightArray is a hand-written port of pocketmine\world\format\LightArray - like
// PalettedBlockArray, there's no PHP source to port from in this checkout (the class is
// referenced throughout SubChunk.php/the light engine but the file itself isn't present in this
// snapshot), so this implements the well-documented standard Bedrock/Java "nibble array" light
// storage format directly: one 4-bit (0-15) light level per position, two positions packed per
// byte, in the same XZY position order PalettedBlockArray uses (see blockIndex).
type LightArray struct {
	data [subChunkBlockCount / 2]byte
}

// NewLightArrayFilled is a port of `LightArray::fill`.
func NewLightArrayFilled(value int) *LightArray {
	l := &LightArray{}
	b := packNibble(value)
	for i := range l.data {
		l.data[i] = b
	}
	return l
}

func packNibble(value int) byte {
	n := byte(value & 0xf)
	return n | n<<4
}

// Get is a port of LightArray::get.
func (l *LightArray) Get(x, y, z int) int {
	i := blockIndex(x, y, z)
	b := l.data[i>>1]
	if i&1 == 0 {
		return int(b & 0x0f)
	}
	return int(b >> 4)
}

// Set is a port of LightArray::set.
func (l *LightArray) Set(x, y, z, value int) {
	i := blockIndex(x, y, z)
	byteIdx := i >> 1
	nibble := byte(value & 0xf)
	if i&1 == 0 {
		l.data[byteIdx] = (l.data[byteIdx] &^ 0x0f) | nibble
	} else {
		l.data[byteIdx] = (l.data[byteIdx] &^ 0xf0) | (nibble << 4)
	}
}

// IsUniform is a port of LightArray::isUniform - used by SubChunk::collectGarbage to decide
// whether a light array can be dropped back to nil (see SubChunk's doc comment on the same
// pattern for block layers).
func (l *LightArray) IsUniform(value int) bool {
	b := packNibble(value)
	for _, v := range l.data {
		if v != b {
			return false
		}
	}
	return true
}

// GetData returns the raw packed nibble bytes (2048 bytes for a full 16x16x16 array) - matching
// what a real Bedrock LevelDB/network light array is stored as, if this port ever needs to
// serialize one (not done yet - see this package's doc comment on ChunkSerializer not needing
// light data over the network).
func (l *LightArray) GetData() []byte {
	out := make([]byte, len(l.data))
	copy(out, l.data[:])
	return out
}

// Clone is a port of LightArray's implicit PHP `clone`.
func (l *LightArray) Clone() *LightArray {
	c := &LightArray{}
	c.data = l.data
	return c
}
