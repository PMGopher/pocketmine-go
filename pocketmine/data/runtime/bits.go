package runtime

import stdmath "math"

// boundedIntAutoBits computes the number of bits needed to encode a value in [min, max],
// shared by Reader, Writer and SizeCalculator so all three always agree.
func boundedIntAutoBits(min, max int) int {
	return int(stdmath.Log2(float64(max-min))) + 1
}
