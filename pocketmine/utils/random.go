package utils

import "time"

// Random is a port of pocketmine\utils\Random: an XorShift128 PRNG used for fast seeded values.
// Most of the logic was adapted from the XorShift128Engine in the php-random library.
//
// This relies on Go's int being 64-bit (true on amd64/arm64, PocketMine's own supported targets),
// matching PHP's 64-bit int, so the bit-shift/mask arithmetic below produces identical results.
const (
	randomX = 123456789
	randomY = 362436069
	randomZ = 521288629
	randomW = 88675123
)

type Random struct {
	x, y, z, w int
	seed       int
}

// NewRandom creates a Random seeded with the given value, or the current time if seed is -1.
func NewRandom(seed int) *Random {
	r := &Random{}
	if seed == -1 {
		seed = int(time.Now().Unix())
	}
	r.SetSeed(seed)
	return r
}

func (r *Random) SetSeed(seed int) {
	r.seed = seed
	r.x = randomX ^ seed
	r.y = (randomY ^ (seed << 17)) | (((seed >> 15) & 0x7fffffff) & 0xffffffff)
	r.z = (randomZ ^ (seed << 31)) | (((seed >> 1) & 0x7fffffff) & 0xffffffff)
	r.w = (randomW ^ (seed << 18)) | (((seed >> 14) & 0x7fffffff) & 0xffffffff)
}

func (r *Random) GetSeed() int {
	return r.seed
}

// NextInt returns a 31-bit integer (not signed).
func (r *Random) NextInt() int {
	return r.NextSignedInt() & 0x7fffffff
}

// NextSignedInt returns a 32-bit integer (signed).
func (r *Random) NextSignedInt() int {
	t := (r.x ^ (r.x << 11)) & 0xffffffff

	r.x = r.y
	r.y = r.z
	r.z = r.w
	r.w = (r.w ^ ((r.w >> 19) & 0x7fffffff) ^ (t ^ ((t >> 8) & 0x7fffffff))) & 0xffffffff

	return signInt(r.w)
}

// NextFloat returns a float between 0.0 and 1.0 (inclusive).
func (r *Random) NextFloat() float64 {
	return float64(r.NextInt()) / 0x7fffffff
}

// NextSignedFloat returns a float between -1.0 and 1.0 (inclusive).
func (r *Random) NextSignedFloat() float64 {
	return float64(r.NextSignedInt()) / 0x7fffffff
}

// NextBoolean returns a random boolean.
func (r *Random) NextBoolean() bool {
	return (r.NextSignedInt() & 0x01) == 0
}

// NextRange returns a random integer between start and end (inclusive).
func (r *Random) NextRange(start, end int) int {
	return start + (r.NextInt() % (end + 1 - start))
}

func (r *Random) NextBoundedInt(bound int) int {
	return r.NextInt() % bound
}

// signInt sign-extends the low 32 bits of value, mirroring pocketmine/binaryutils Binary::signInt().
func signInt(value int) int {
	return value << 32 >> 32
}
