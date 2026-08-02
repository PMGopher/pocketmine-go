package noise

import (
	"pocketmine-go/pocketmine/utils"
)

// Simplex is a port of pocketmine\world\generator\noise\Simplex, folded together with its abstract
// Noise base class (Simplex is the only Noise subclass this port needs - BiomeSelector and Normal
// are its only two callers, both only ever construct a Simplex - so there's no separate abstract
// base type here, just Simplex's own fields for what Noise::__construct held).
type Simplex struct {
	octaves     int
	persistence float64
	expansion   float64

	offsetX, offsetY, offsetZ float64
	perm                      [512]int
}

var grad3 = [12][3]float64{
	{1, 1, 0}, {-1, 1, 0}, {1, -1, 0}, {-1, -1, 0},
	{1, 0, 1}, {-1, 0, 1}, {1, 0, -1}, {-1, 0, -1},
	{0, 1, 1}, {0, -1, 1}, {0, 1, -1}, {0, -1, -1},
}

const (
	simplexF2  = 0.5 * (1.7320508075688772 - 1) // 0.5 * (sqrt(3) - 1)
	simplexG2  = (3 - 1.7320508075688772) / 6   // (3 - sqrt(3)) / 6
	simplexG22 = simplexG2*2.0 - 1
	simplexF3  = 1.0 / 3.0
	simplexG3  = 1.0 / 6.0
)

// NewSimplex is a port of Simplex::__construct.
func NewSimplex(random *utils.Random, octaves int, persistence, expansion float64) *Simplex {
	s := &Simplex{octaves: octaves, persistence: persistence, expansion: expansion}

	s.offsetX = random.NextFloat() * 256
	s.offsetY = random.NextFloat() * 256
	s.offsetZ = random.NextFloat() * 256

	for i := 0; i < 256; i++ {
		s.perm[i] = random.NextBoundedInt(256)
	}
	for i := 0; i < 256; i++ {
		pos := random.NextBoundedInt(256-i) + i
		old := s.perm[i]
		s.perm[i] = s.perm[pos]
		s.perm[pos] = old
		s.perm[i+256] = s.perm[i]
	}

	// This dummy call is necessary to produce the same RNG state as PHP's Simplex constructor -
	// see Simplex.php's identical comment (a leftover offsetW read from before a refactor, kept
	// only so the RNG stream downstream stays byte-for-byte identical to the original).
	random.NextSignedInt()

	return s
}

// GetNoise3D is a port of Simplex::getNoise3D.
func (s *Simplex) GetNoise3D(x, y, z float64) float64 {
	x += s.offsetX
	y += s.offsetY
	z += s.offsetZ

	sk := (x + y + z) * simplexF3
	// PHP's `(int) ($x + $s)` truncates toward zero, not floor - Go's int() conversion does the
	// same thing, so this deliberately isn't math.Floor (which would be wrong for negative coords).
	i := int(x + sk)
	j := int(y + sk)
	k := int(z + sk)
	t := float64(i+j+k) * simplexG3
	x0 := x - (float64(i) - t)
	y0 := y - (float64(j) - t)
	z0 := z - (float64(k) - t)

	var i1, j1, k1, i2, j2, k2 int
	if x0 >= y0 {
		if y0 >= z0 {
			i1, j1, k1, i2, j2, k2 = 1, 0, 0, 1, 1, 0
		} else if x0 >= z0 {
			i1, j1, k1, i2, j2, k2 = 1, 0, 0, 1, 0, 1
		} else {
			i1, j1, k1, i2, j2, k2 = 0, 0, 1, 1, 0, 1
		}
	} else {
		if y0 < z0 {
			i1, j1, k1, i2, j2, k2 = 0, 0, 1, 0, 1, 1
		} else if x0 < z0 {
			i1, j1, k1, i2, j2, k2 = 0, 1, 0, 0, 1, 1
		} else {
			i1, j1, k1, i2, j2, k2 = 0, 1, 0, 1, 1, 0
		}
	}

	x1 := x0 - float64(i1) + simplexG3
	y1 := y0 - float64(j1) + simplexG3
	z1 := z0 - float64(k1) + simplexG3
	x2 := x0 - float64(i2) + 2.0*simplexG3
	y2 := y0 - float64(j2) + 2.0*simplexG3
	z2 := z0 - float64(k2) + 2.0*simplexG3
	x3 := x0 - 1.0 + 3.0*simplexG3
	y3 := y0 - 1.0 + 3.0*simplexG3
	z3 := z0 - 1.0 + 3.0*simplexG3

	ii := i & 255
	jj := j & 255
	kk := k & 255

	n := 0.0
	perm := s.perm

	if t0 := 0.6 - x0*x0 - y0*y0 - z0*z0; t0 > 0 {
		gi0 := grad3[perm[ii+perm[jj+perm[kk]]]%12]
		n += t0 * t0 * t0 * t0 * (gi0[0]*x0 + gi0[1]*y0 + gi0[2]*z0)
	}
	if t1 := 0.6 - x1*x1 - y1*y1 - z1*z1; t1 > 0 {
		gi1 := grad3[perm[ii+i1+perm[jj+j1+perm[kk+k1]]]%12]
		n += t1 * t1 * t1 * t1 * (gi1[0]*x1 + gi1[1]*y1 + gi1[2]*z1)
	}
	if t2 := 0.6 - x2*x2 - y2*y2 - z2*z2; t2 > 0 {
		gi2 := grad3[perm[ii+i2+perm[jj+j2+perm[kk+k2]]]%12]
		n += t2 * t2 * t2 * t2 * (gi2[0]*x2 + gi2[1]*y2 + gi2[2]*z2)
	}
	if t3 := 0.6 - x3*x3 - y3*y3 - z3*z3; t3 > 0 {
		gi3 := grad3[perm[ii+1+perm[jj+1+perm[kk+1]]]%12]
		n += t3 * t3 * t3 * t3 * (gi3[0]*x3 + gi3[1]*y3 + gi3[2]*z3)
	}

	return 32.0 * n
}

// GetNoise2D is a port of Simplex::getNoise2D.
func (s *Simplex) GetNoise2D(x, y float64) float64 {
	x += s.offsetX
	y += s.offsetY

	sk := (x + y) * simplexF2
	i := int(x + sk)
	j := int(y + sk)
	t := float64(i+j) * simplexG2
	x0 := x - (float64(i) - t)
	y0 := y - (float64(j) - t)

	var i1, j1 int
	if x0 > y0 {
		i1, j1 = 1, 0
	} else {
		i1, j1 = 0, 1
	}

	x1 := x0 - float64(i1) + simplexG2
	y1 := y0 - float64(j1) + simplexG2
	x2 := x0 + simplexG22
	y2 := y0 + simplexG22

	ii := i & 255
	jj := j & 255

	n := 0.0

	if t0 := 0.5 - x0*x0 - y0*y0; t0 > 0 {
		gi0 := grad3[s.perm[ii+s.perm[jj]]%12]
		n += t0 * t0 * t0 * t0 * (gi0[0]*x0 + gi0[1]*y0)
	}
	if t1 := 0.5 - x1*x1 - y1*y1; t1 > 0 {
		gi1 := grad3[s.perm[ii+i1+s.perm[jj+j1]]%12]
		n += t1 * t1 * t1 * t1 * (gi1[0]*x1 + gi1[1]*y1)
	}
	if t2 := 0.5 - x2*x2 - y2*y2; t2 > 0 {
		gi2 := grad3[s.perm[ii+1+s.perm[jj+1]]%12]
		n += t2 * t2 * t2 * t2 * (gi2[0]*x2 + gi2[1]*y2)
	}

	return 70.0 * n
}

// Noise2D is a port of Noise::noise2D.
func (s *Simplex) Noise2D(x, z float64, normalized bool) float64 {
	result, amp, freq, max := 0.0, 1.0, 1.0, 0.0

	x *= s.expansion
	z *= s.expansion

	for i := 0; i < s.octaves; i++ {
		result += s.GetNoise2D(x*freq, z*freq) * amp
		max += amp
		freq *= 2
		amp *= s.persistence
	}

	if normalized {
		result /= max
	}
	return result
}

// Noise3D is a port of Noise::noise3D.
func (s *Simplex) Noise3D(x, y, z float64, normalized bool) float64 {
	result, amp, freq, max := 0.0, 1.0, 1.0, 0.0

	x *= s.expansion
	y *= s.expansion
	z *= s.expansion

	for i := 0; i < s.octaves; i++ {
		result += s.GetNoise3D(x*freq, y*freq, z*freq) * amp
		max += amp
		freq *= 2
		amp *= s.persistence
	}

	if normalized {
		result /= max
	}
	return result
}

// GetFastNoise3D is a port of Noise::getFastNoise3D: samples noise3D on a coarse grid
// (xSamplingRate/ySamplingRate/zSamplingRate apart) and trilinearly interpolates the rest, far
// cheaper than sampling every position directly. Returns a [x][z][y]float64 cube matching the PHP
// original's array shape (indexed x, then z, then y).
func (s *Simplex) GetFastNoise3D(xSize, ySize, zSize, xSamplingRate, ySamplingRate, zSamplingRate, x, y, z int) [][][]float64 {
	noiseArray := make([][][]float64, xSize+1)
	for i := range noiseArray {
		noiseArray[i] = make([][]float64, zSize+1)
		for j := range noiseArray[i] {
			noiseArray[i][j] = make([]float64, ySize+1)
		}
	}

	for xx := 0; xx <= xSize; xx += xSamplingRate {
		for zz := 0; zz <= zSize; zz += zSamplingRate {
			for yy := 0; yy <= ySize; yy += ySamplingRate {
				noiseArray[xx][zz][yy] = s.Noise3D(float64(x+xx), float64(y+yy), float64(z+zz), true)
			}
		}
	}

	xLerpStep := 1 / float64(xSamplingRate)
	yLerpStep := 1 / float64(ySamplingRate)
	zLerpStep := 1 / float64(zSamplingRate)

	for leftX := 0; leftX < xSize; leftX += xSamplingRate {
		rightX := leftX + xSamplingRate
		for leftZ := 0; leftZ < zSize; leftZ += zSamplingRate {
			rightZ := leftZ + zSamplingRate
			for leftY := 0; leftY < ySize; leftY += ySamplingRate {
				rightY := leftY + ySamplingRate

				c000 := noiseArray[leftX][leftZ][leftY]
				c100 := noiseArray[rightX][leftZ][leftY]
				c001 := noiseArray[leftX][leftZ][rightY]
				c101 := noiseArray[rightX][leftZ][rightY]
				c010 := noiseArray[leftX][rightZ][leftY]
				c110 := noiseArray[rightX][rightZ][leftY]
				c011 := noiseArray[leftX][rightZ][rightY]
				c111 := noiseArray[rightX][rightZ][rightY]

				for xStep := 0; xStep < xSamplingRate; xStep++ {
					xx := leftX + xStep
					dx2 := float64(xStep) * xLerpStep
					dx1 := 1 - dx2

					x00 := c000*dx1 + c100*dx2
					x01 := c001*dx1 + c101*dx2
					x10 := c010*dx1 + c110*dx2
					x11 := c011*dx1 + c111*dx2

					for zStep := 0; zStep < zSamplingRate; zStep++ {
						zz := leftZ + zStep
						dz2 := float64(zStep) * zLerpStep
						dz1 := 1 - dz2

						z0 := x00*dz1 + x10*dz2
						z1 := x01*dz1 + x11*dz2

						yStart := 0
						if xStep == 0 && zStep == 0 {
							yStart = 1
						}
						for yStep := yStart; yStep < ySamplingRate; yStep++ {
							yy := leftY + yStep
							dy2 := float64(yStep) * yLerpStep
							dy1 := 1 - dy2

							noiseArray[xx][zz][yy] = dy1*z0 + dy2*z1
						}
					}
				}
			}
		}
	}

	return noiseArray
}
