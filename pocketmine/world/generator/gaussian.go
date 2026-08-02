package generator

import "math"

// Gaussian is a port of pocketmine\world\generator\Gaussian: a precomputed 1D/2D Gaussian
// smoothing kernel used to blur biome elevation across chunk borders.
type Gaussian struct {
	SmoothSize int

	Kernel1D    []float64
	WeightSum1D float64

	Kernel    [][]float64
	WeightSum float64
}

// NewGaussian is a port of Gaussian::__construct.
func NewGaussian(smoothSize int) *Gaussian {
	g := &Gaussian{SmoothSize: smoothSize}

	bellSize := 1 / float64(smoothSize)
	bellHeight := 2 * float64(smoothSize)

	size := 2*smoothSize + 1
	g.Kernel1D = make([]float64, size)
	g.Kernel = make([][]float64, size)

	for sx := -smoothSize; sx <= smoothSize; sx++ {
		bx := bellSize * float64(sx)

		g.Kernel1D[sx+smoothSize] = math.Sqrt(bellHeight) * math.Exp(-(bx*bx)/2)

		row := make([]float64, size)
		for sz := -smoothSize; sz <= smoothSize; sz++ {
			bz := bellSize * float64(sz)
			row[sz+smoothSize] = bellHeight * math.Exp(-(bx*bx+bz*bz)/2)
		}
		g.Kernel[sx+smoothSize] = row
	}

	for _, v := range g.Kernel1D {
		g.WeightSum1D += v
	}
	g.WeightSum = g.WeightSum1D * g.WeightSum1D

	return g
}
