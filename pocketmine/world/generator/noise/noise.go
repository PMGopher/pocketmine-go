// Package noise is a port of pocketmine\world\generator\noise: the Simplex noise generator used
// by both Normal's terrain shaping and BiomeSelector's climate maps.
package noise

// LinearLerp is a port of Noise::linearLerp.
func LinearLerp(x, x1, x2, q0, q1 float64) float64 {
	return ((x2-x)/(x2-x1))*q0 + ((x-x1)/(x2-x1))*q1
}

// BilinearLerp is a port of Noise::bilinearLerp.
func BilinearLerp(x, y, q00, q01, q10, q11, x1, x2, y1, y2 float64) float64 {
	dx1 := (x2 - x) / (x2 - x1)
	dx2 := (x - x1) / (x2 - x1)

	return ((y2-y)/(y2-y1))*(dx1*q00+dx2*q10) + ((y-y1)/(y2-y1))*(dx1*q01+dx2*q11)
}
