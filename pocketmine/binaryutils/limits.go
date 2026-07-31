package binaryutils

// Limits is a port of pocketmine\utils\Limits.
//
// Go already exposes these via math.MaxInt8/math.MinInt8 etc. in the standard library; these
// constants exist only so ported code that referenced Limits::* by name has a direct equivalent.
const (
	Uint8Max = 0xff
	Int8Min  = -0x7f - 1
	Int8Max  = 0x7f

	Uint16Max = 0xffff
	Int16Min  = -0x7fff - 1
	Int16Max  = 0x7fff

	Uint32Max = 0xffffffff
	Int32Min  = -0x7fffffff - 1
	Int32Max  = 0x7fffffff

	Uint64Max = 0xffffffffffffffff
	Int64Min  = -0x7fffffffffffffff - 1
	Int64Max  = 0x7fffffffffffffff
)
