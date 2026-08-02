package block

import "math/rand"

const (
	cropGrowthOnHydratedFarmlandBonus       = 3.0
	cropGrowthOnDryFarmlandBonus            = 1.0
	cropGrowthAdjacentHydratedFarmlandBonus = 3.0 / 4
	cropGrowthAdjacentDryFarmlandBonus      = 1.0 / 4
	cropGrowthImproperArrangementDivisor    = 2.0

	CropGrowthMinLightLevel = 9
)

// CropGrowthCalculateMultiplier is a port of pocketmine\block\utils\CropGrowthHelper::calculateMultiplier.
func CropGrowthCalculateMultiplier(blk Behavior) float64 {
	result := 1.0

	pos := blk.GetPosition()
	world, err := pos.GetWorld()
	if err != nil {
		return result
	}
	baseX, baseY, baseZ := pos.FloorX(), pos.FloorY(), pos.FloorZ()

	if farmland, ok := world.GetBlockAt(baseX, baseY-1, baseZ).(*Farmland); ok {
		if farmland.GetWetness() > 0 {
			result += cropGrowthOnHydratedFarmlandBonus
		} else {
			result += cropGrowthOnDryFarmlandBonus
		}
	}

	xRow, zRow, improperArrangement := false, false, false

	for x := -1; x <= 1; x++ {
		for z := -1; z <= 1; z++ {
			if x == 0 && z == 0 {
				continue
			}
			nextFarmland, ok := world.GetBlockAt(baseX+x, baseY-1, baseZ+z).(*Farmland)
			if !ok {
				continue
			}

			if nextFarmland.GetWetness() > 0 {
				result += cropGrowthAdjacentHydratedFarmlandBonus
			} else {
				result += cropGrowthAdjacentDryFarmlandBonus
			}

			if !improperArrangement {
				nextCrop := world.GetBlockAt(baseX+x, baseY, baseZ+z)
				if nextCrop.(blockGeometry).HasSameTypeId(blk) {
					switch 0 {
					case x:
						if zRow {
							improperArrangement = true
						} else {
							xRow = true
						}
					case z:
						if xRow {
							improperArrangement = true
						} else {
							zRow = true
						}
					default:
						improperArrangement = true
					}
				}
			}
		}
	}

	// crops can be arranged in rows, but the rows must not cross and must be spaced apart by at
	// least one block
	if improperArrangement {
		result /= cropGrowthImproperArrangementDivisor
	}

	return result
}

// CropGrowthHasEnoughLight is a port of CropGrowthHelper::hasEnoughLight, with minLevel defaulted
// to CropGrowthMinLightLevel (Go has no default parameters) - callers needing a different level
// should query World.GetPotentialLightAt directly instead.
func CropGrowthHasEnoughLight(blk Behavior) bool {
	pos := blk.GetPosition()
	world, err := pos.GetWorld()
	if err != nil {
		return false
	}
	return world.GetPotentialLightAt(pos.FloorX(), pos.FloorY(), pos.FloorZ()) >= CropGrowthMinLightLevel
}

// CropGrowthCanGrow is a port of CropGrowthHelper::canGrow.
func CropGrowthCanGrow(blk Behavior) bool {
	multiplier := CropGrowthCalculateMultiplier(blk)
	return rand.Intn(int(25/multiplier)+1) == 0 && CropGrowthHasEnoughLight(blk)
}
