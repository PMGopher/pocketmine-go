package player

import (
	"iter"
	"math"
)

// SelectChunks is a port of ChunkSelector::selectChunks: yields chunk coordinates in expanding
// "ring" order outward from (centerX, centerZ) - closer chunks first, matching real PHP's own
// per-ring generator semantics.
//
// Real PHP yields `$subRadius => chunkHash` (the ring index as the generator key, a
// World::chunkHash-packed value). This port yields chunk coordinates directly instead of a packed
// hash (matching this port's established preference for native [2]int map/iteration keys
// elsewhere, e.g. World's own chunkKey), and drops the ring-index key entirely: the one real
// caller (WorldManager::generateWorld's background generation) already discards it via
// `iterator_to_array(..., preserve_keys: false)`, and Go's iteration order itself already
// preserves the same ring-by-ring ordering, so nothing is lost by not exposing it separately.
func SelectChunks(radius, centerX, centerZ int) iter.Seq[[2]int] {
	return func(yield func([2]int) bool) {
		for subRadius := 0; subRadius < radius; subRadius++ {
			subRadiusSquared := subRadius * subRadius
			nextSubRadiusSquared := (subRadius + 1) * (subRadius + 1)
			minX := int(float64(subRadius) / math.Sqrt2)

			lastZ := 0

			for x := subRadius; x >= minX; x-- {
				for z := lastZ; z <= x; z++ {
					distanceSquared := x*x + z*z
					if distanceSquared < subRadiusSquared {
						continue
					} else if distanceSquared >= nextSubRadiusSquared {
						break
					}

					lastZ = z
					// If the chunk is in the radius, others at the same offsets in different
					// quadrants are also guaranteed to be.

					if !yield([2]int{centerX + x, centerZ + z}) { // top right quadrant
						return
					}
					if !yield([2]int{centerX - x - 1, centerZ + z}) { // top left quadrant
						return
					}
					if !yield([2]int{centerX + x, centerZ - z - 1}) { // bottom right quadrant
						return
					}
					if !yield([2]int{centerX - x - 1, centerZ - z - 1}) { // bottom left quadrant
						return
					}

					if x != z {
						if !yield([2]int{centerX + z, centerZ + x}) { // top right quadrant mirror
							return
						}
						if !yield([2]int{centerX - z - 1, centerZ + x}) { // top left quadrant mirror
							return
						}
						if !yield([2]int{centerX + z, centerZ - x - 1}) { // bottom right quadrant mirror
							return
						}
						if !yield([2]int{centerX - z - 1, centerZ - x - 1}) { // bottom left quadrant mirror
							return
						}
					}
				}
			}
		}
	}
}
