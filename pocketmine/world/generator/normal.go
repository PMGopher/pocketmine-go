package generator

import (
	"math"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/utils"
	"pocketmine-go/pocketmine/world/biome"
	"pocketmine-go/pocketmine/world/format"
	"pocketmine-go/pocketmine/world/generator/biomeselector"
	"pocketmine-go/pocketmine/world/generator/groundcover"
	"pocketmine-go/pocketmine/world/generator/noise"
	"pocketmine-go/pocketmine/world/generator/populator"
)

// normalNoiseSamplingRateY is a port of Normal::NOISE_SAMPLING_RATE_Y.
const normalNoiseSamplingRateY = 8

// Normal is a port of pocketmine\world\generator\normal\Normal: real noise-shaped, biome-varied
// terrain (hills, oceans, deserts, ...), as opposed to Flat's single fixed layer stack.
//
// Two documented simplifications from the PHP original, both consequences of choices already made
// elsewhere in this port rather than new shortcuts:
//   - GroundCover runs as a regular populator (Populators()[0]) instead of PHP's separate
//     "generationPopulators" phase spliced into the middle of GenerateChunk - see Flat's identical
//     GenerateChunk/PopulateChunk split (this port's Generator interface doesn't hand GenerateChunk
//     a World to write through, so GroundCover can't run mid-generation the way PHP's does; running
//     it first thing in PopulateChunk instead is behaviourally equivalent, since nothing else reads
//     the chunk in between either way).
//   - pickBiome's hash mixing uses plain wrapping int64 arithmetic instead of replicating PHP's
//     int-to-float overflow promotion (`$hash *= $hash + 223` can silently become a float in PHP -
//     see pickBiome's own comment on why that's kept there). That quirk exists in upstream only to
//     avoid shifting terrain in already-generated player worlds; this port never reads a PHP-
//     generated world, so there's nothing to stay compatible with - just a different, equally
//     deterministic per-seed hash.
type Normal struct {
	seed        int
	waterHeight int

	random    *utils.Random
	noiseBase *noise.Simplex
	selector  *biomeselector.Selector
	registry  *biome.Registry
	gaussian  *Gaussian

	populators []populator.Populator

	airStateID        int32
	bedrockStateID    int32
	stillWaterStateID int32
	stoneStateID      int32
}

// NewNormal is a port of Normal::__construct.
func NewNormal(seed int) *Normal {
	n := &Normal{seed: seed, waterHeight: 62, gaussian: NewGaussian(2)}

	random := utils.NewRandom(seed)
	n.noiseBase = noise.NewSimplex(random, 4, 1.0/4, 1.0/32)
	random.SetSeed(seed)
	n.random = random

	n.registry = biome.NewRegistry()
	n.selector = biomeselector.New(n.random, n.registry, normalBiomeLookup)
	n.selector.Recalculate()

	n.populators = []populator.Populator{
		groundcover.New(n.registry),
		normalOrePopulator(),
	}

	n.airStateID = int32(block.VanillaAir().GetStateId())
	n.bedrockStateID = int32(block.VanillaBedrock().GetStateId())
	n.stillWaterStateID = int32(block.VanillaWater().GetStateId())
	n.stoneStateID = int32(block.VanillaStone().GetStateId())

	return n
}

// normalOrePopulator builds the same Ore populator (identical 8 OreType tuples) as Flat's
// "decoration" option - see VanillaFlatOreTypes's doc comment for the exact source line, shared
// verbatim by both generators in real PocketMine-MP.
func normalOrePopulator() populator.Populator {
	ore := &populator.Ore{}
	ore.SetOreTypes(VanillaFlatOreTypes())
	return ore
}

// normalBiomeLookup is a port of the anonymous BiomeSelector subclass Normal's constructor
// defines inline in PHP - Go has no anonymous subclassing, so this is a plain function passed to
// biomeselector.New instead (see biomeselector.LookupFunc's doc comment).
func normalBiomeLookup(temperature, rainfall float64) int {
	switch {
	case rainfall < 0.25:
		switch {
		case temperature < 0.7:
			return biome.IDOcean
		case temperature < 0.85:
			return biome.IDRiver
		default:
			return biome.IDSwampland
		}
	case rainfall < 0.60:
		switch {
		case temperature < 0.25:
			return biome.IDIcePlains
		case temperature < 0.75:
			return biome.IDPlains
		default:
			return biome.IDDesert
		}
	case rainfall < 0.80:
		switch {
		case temperature < 0.25:
			return biome.IDTaiga
		case temperature < 0.75:
			return biome.IDForest
		default:
			return biome.IDBirchForest
		}
	default:
		switch {
		case temperature < 0.20:
			return biome.IDExtremeHills
		case temperature < 0.40:
			return biome.IDExtremeHillsEdge
		default:
			return biome.IDRiver
		}
	}
}

// pickBiome is a port of Normal::pickBiome - see this type's doc comment on the hash-mixing
// deviation from PHP.
func (n *Normal) pickBiome(x, z int) *biome.Biome {
	hash := int64(x)*2345803 ^ int64(z)*9236449 ^ int64(n.seed)
	hash *= hash + 223
	xNoise := int((hash >> 20) & 3)
	zNoise := int((hash >> 22) & 3)
	if xNoise == 3 {
		xNoise = 1
	}
	if zNoise == 3 {
		zNoise = 1
	}
	return n.selector.PickBiome(float64(x+xNoise-1), float64(z+zNoise-1))
}

// GenerateChunk is a port of Normal::generateChunk, minus the GroundCover generation-populator
// phase - see this type's doc comment on why that moved to PopulateChunk instead.
func (n *Normal) GenerateChunk(chunkX, chunkZ int) *format.Chunk {
	n.random.SetSeed(0xdeadbeef ^ (chunkX << 8) ^ chunkZ ^ n.seed)

	baseX := chunkX * format.SubChunkEdgeLength
	baseZ := chunkZ * format.SubChunkEdgeLength

	biomeArray, minHeights, maxHeights := n.generateBiomes(baseX, baseZ)

	lowestNoiseBlock := int(math.Floor(min2D(minHeights)))
	highestNoiseBlock := int(math.Ceil(max2D(maxHeights)))

	// getFastNoise3D expects the inputs to be aligned with the sampling rate, otherwise the
	// samples will be taken from different coordinates than intended.
	noiseMin := floorDivInt(lowestNoiseBlock, normalNoiseSamplingRateY) * normalNoiseSamplingRateY
	noiseMax := ceilDivInt(highestNoiseBlock, normalNoiseSamplingRateY) * normalNoiseSamplingRateY

	noiseCube := n.noiseBase.GetFastNoise3D(
		format.SubChunkEdgeLength, noiseMax-noiseMin, format.SubChunkEdgeLength,
		4, normalNoiseSamplingRateY, 4,
		chunkX*format.SubChunkEdgeLength, noiseMin, chunkZ*format.SubChunkEdgeLength,
	)

	minNoiseSubChunk := floorDivInt(noiseMin, format.SubChunkEdgeLength)

	subChunks := map[int]*format.SubChunk{}
	for y := format.MinSubChunkIndex; y <= format.MaxSubChunkIndex; y++ {
		var blocks []*format.PalettedBlockArray
		if y >= 0 && y < minNoiseSubChunk {
			// Everything above 0 and below noiseMin is always solid stone, which can be
			// flood-filled instead of setting the blocks one at a time - this is vastly faster.
			blocks = []*format.PalettedBlockArray{format.NewPalettedBlockArray(n.stoneStateID)}
		}
		subChunks[y] = format.NewSubChunk(n.airStateID, blocks, biomeArray.Clone())
	}
	chunk := format.NewChunk(subChunks, false, n.airStateID, int32(biome.IDOcean))

	for x := 0; x < format.SubChunkEdgeLength; x++ {
		for z := 0; z < format.SubChunkEdgeLength; z++ {
			chunk.SetBlockStateID(x, 0, z, n.bedrockStateID)

			minSum := minHeights[x][z]
			maxSum := maxHeights[x][z]
			maxBlockY := math.Max(maxSum, float64(n.waterHeight))
			smoothHeight := (maxSum - minSum) / 2

			// Everything below minSum is always solid stone - the subchunks below minNoiseSubChunk
			// were already flood-filled above, so only the gap in this column needs filling here.
			for y := minNoiseSubChunk * format.SubChunkEdgeLength; float64(y) < minSum; y++ {
				chunk.SetBlockStateID(x, y, z, n.stoneStateID)
			}
			for y := int(math.Floor(minSum)); float64(y) <= maxBlockY; y++ {
				var noiseValue float64
				if y > noiseMax {
					// noiseValue would anyway be <= 0 above maxSum, since the smoothing term is >= 1.
					noiseValue = -1
				} else {
					noiseValue = noiseCube[x][z][y-noiseMin] - 1/smoothHeight*(float64(y)-smoothHeight-minSum)
				}

				if noiseValue > 0 {
					chunk.SetBlockStateID(x, y, z, n.stoneStateID)
				} else if y <= n.waterHeight {
					chunk.SetBlockStateID(x, y, z, n.stillWaterStateID)
				}
			}
		}
	}

	return chunk
}

// PopulateChunk is a port of Normal::populateChunk.
func (n *Normal) PopulateChunk(world block.World, chunkX, chunkZ int) {
	n.random.SetSeed(0xdeadbeef ^ (chunkX << 8) ^ chunkZ ^ n.seed)
	for _, p := range n.populators {
		p.Populate(world, chunkX, chunkZ, n.random)
	}

	pos := block.NewPosition(float64(chunkX*format.SubChunkEdgeLength+7), 7, float64(chunkZ*format.SubChunkEdgeLength+7), world)
	if chunk, ok := world.GetOrLoadChunkAtPosition(pos); ok {
		b := n.registry.GetBiome(int(chunk.GetBiomeID(7, 7, 7)))
		b.PopulateChunk(world, chunkX, chunkZ, n.random)
	}
}

// generateBiomes is a port of Normal::generateBiomes.
func (n *Normal) generateBiomes(baseX, baseZ int) (biomeArray *format.PalettedBlockArray, minHeights, maxHeights [16][16]float64) {
	padding := n.gaussian.SmoothSize
	start := -padding
	end := format.SubChunkEdgeLength + padding
	rangeSize := end - start

	biomeCache := make([][]*biome.Biome, rangeSize)
	for i := range biomeCache {
		biomeCache[i] = make([]*biome.Biome, rangeSize)
	}

	biomeArray = format.NewPalettedBlockArray(int32(biome.IDOcean))

	uniformSet, isUniform, uniformID := false, true, 0

	for x := start; x < end; x++ {
		absoluteX := baseX + x
		for z := start; z < end; z++ {
			absoluteZ := baseZ + z

			b := n.pickBiome(absoluteX, absoluteZ)
			biomeCache[x-start][z-start] = b
			biomeID := b.ID()

			if !uniformSet {
				uniformSet = true
				uniformID = biomeID
			} else if isUniform && biomeID != uniformID {
				isUniform = false
			}

			if x >= 0 && x < format.SubChunkEdgeLength && z >= 0 && z < format.SubChunkEdgeLength {
				for y := 0; y < 16; y++ {
					biomeArray.Set(x, y, z, int32(biomeID))
				}
			}
		}
	}

	if !isUniform {
		minHeights, maxHeights = n.gaussianSmoothElevation(start, end, biomeCache)
	} else {
		// If every biome in the blurred area is the same, blurring can be skipped entirely.
		b := n.registry.GetBiome(uniformID)
		minElevation := float64(b.MinElevation() - 1)
		maxElevation := float64(b.MaxElevation())
		for x := 0; x < format.SubChunkEdgeLength; x++ {
			for z := 0; z < format.SubChunkEdgeLength; z++ {
				minHeights[x][z] = minElevation
				maxHeights[x][z] = maxElevation
			}
		}
	}

	return biomeArray, minHeights, maxHeights
}

// gaussianSmoothElevation is a port of Normal::gaussianSmoothElevation. biomeCache is indexed
// [x-start][z-start], matching generateBiomes' own padded storage.
func (n *Normal) gaussianSmoothElevation(start, end int, biomeCache [][]*biome.Biome) (minHeights, maxHeights [16][16]float64) {
	smoothSize := n.gaussian.SmoothSize
	rangeSize := end - start

	// Blur along the X axis first. While the padding corners don't need smoothing themselves, their
	// contribution must still be included in the Z padding below, otherwise chunk-corner artifacts
	// appear.
	minHeightsX := make([][]float64, format.SubChunkEdgeLength)
	maxHeightsX := make([][]float64, format.SubChunkEdgeLength)
	for x := 0; x < format.SubChunkEdgeLength; x++ {
		minHeightsX[x] = make([]float64, rangeSize)
		maxHeightsX[x] = make([]float64, rangeSize)
		for z := start; z < end; z++ {
			minSum, maxSum := 0.0, 0.0
			for sx := -smoothSize; sx <= smoothSize; sx++ {
				weight := n.gaussian.Kernel1D[sx+smoothSize]
				adjacent := biomeCache[x+sx-start][z-start]

				minSum += (float64(adjacent.MinElevation()) - 1) * weight
				maxSum += float64(adjacent.MaxElevation()) * weight
			}
			minHeightsX[x][z-start] = minSum / n.gaussian.WeightSum1D
			maxHeightsX[x][z-start] = maxSum / n.gaussian.WeightSum1D
		}
	}

	// Then the Z axis, using the blurred values from the previous loop.
	for x := 0; x < format.SubChunkEdgeLength; x++ {
		for z := 0; z < format.SubChunkEdgeLength; z++ {
			minSum, maxSum := 0.0, 0.0
			for sz := -smoothSize; sz <= smoothSize; sz++ {
				weight := n.gaussian.Kernel1D[sz+smoothSize]
				zIdx := z + sz - start

				minSum += minHeightsX[x][zIdx] * weight
				maxSum += maxHeightsX[x][zIdx] * weight
			}
			minHeights[x][z] = minSum / n.gaussian.WeightSum1D
			maxHeights[x][z] = maxSum / n.gaussian.WeightSum1D
		}
	}

	return minHeights, maxHeights
}

func min2D(a [16][16]float64) float64 {
	m := a[0][0]
	for _, row := range a {
		for _, v := range row {
			if v < m {
				m = v
			}
		}
	}
	return m
}

func max2D(a [16][16]float64) float64 {
	m := a[0][0]
	for _, row := range a {
		for _, v := range row {
			if v > m {
				m = v
			}
		}
	}
	return m
}

// floorDivInt and ceilDivInt are integer floor/ceiling division, matching PHP's
// `(int) floor($a / $b)` / `(int) ceil($a / $b)` for possibly-negative $a (Go's native `/` on ints
// truncates toward zero, which is wrong for negative numerators - e.g. -20/8 must floor to -3, not
// truncate to -2).
func floorDivInt(a, b int) int {
	q := a / b
	if a%b != 0 && (a < 0) != (b < 0) {
		q--
	}
	return q
}

func ceilDivInt(a, b int) int {
	q := a / b
	if a%b != 0 && (a < 0) == (b < 0) {
		q++
	}
	return q
}
