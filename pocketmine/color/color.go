package color

// Color is a port of pocketmine\color\Color.
type Color struct {
	r, g, b, a uint8
}

// NewColor mirrors the Color constructor, with A defaulting to 0xff (opaque) as in PHP.
func NewColor(r, g, b uint8, a ...uint8) Color {
	alpha := uint8(0xff)
	if len(a) > 0 {
		alpha = a[0]
	}
	return Color{r: r, g: g, b: b, a: alpha}
}

func (c Color) GetA() uint8 { return c.a }
func (c Color) GetR() uint8 { return c.r }
func (c Color) GetG() uint8 { return c.g }
func (c Color) GetB() uint8 { return c.b }

// Mix mixes the supplied list of colours together to produce a result colour.
func Mix(first Color, rest ...Color) Color {
	colors := append([]Color{first}, rest...)
	var a, r, g, b int
	for _, c := range colors {
		a += int(c.a)
		r += int(c.r)
		g += int(c.g)
		b += int(c.b)
	}
	count := len(colors)
	return NewColor(uint8(r/count), uint8(g/count), uint8(b/count), uint8(a/count))
}

// FromRGB returns a Color from the supplied RGB colour code (24-bit).
func FromRGB(code int32) Color {
	return NewColor(uint8(code>>16), uint8(code>>8), uint8(code))
}

// FromARGB returns a Color from the supplied ARGB colour code (32-bit).
func FromARGB(code int32) Color {
	return NewColor(uint8(code>>16), uint8(code>>8), uint8(code), uint8(code>>24))
}

// ToARGB returns an ARGB 32-bit colour value.
func (c Color) ToARGB() int32 {
	return int32(c.a)<<24 | int32(c.r)<<16 | int32(c.g)<<8 | int32(c.b)
}

// FromRGBA returns a Color from the supplied RGBA colour code (32-bit).
func FromRGBA(code int32) Color {
	return NewColor(uint8(code>>24), uint8(code>>16), uint8(code>>8), uint8(code))
}

// ToRGBA returns an RGBA 32-bit colour value.
func (c Color) ToRGBA() int32 {
	return int32(c.r)<<24 | int32(c.g)<<16 | int32(c.b)<<8 | int32(c.a)
}

func (c Color) Equals(other Color) bool {
	return c == other
}
