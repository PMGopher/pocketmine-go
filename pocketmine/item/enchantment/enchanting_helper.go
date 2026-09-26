package enchantment

import (
	"math"
	"math/rand/v2"

	"pocketmine-go/pocketmine/binaryutils"
	"pocketmine-go/pocketmine/utils"
)

// EnchantingHelper is a port of pocketmine\item\enchantment\EnchantingHelper: helpers for
// enchanting with the enchanting table. EnchantingHelper::enchantItem needs VanillaItems, so it
// lives in the item package (item.EnchantItem).

const maxBookshelfCount = 15

// EnchantableItem is the part of item.Item GenerateOptions uses.
type EnchantableItem interface {
	TaggedItem
	IsNull() bool
	GetEnchantability() int
}

// TableSurroundings tells GenerateOptions what is around an enchanting table: whether the block
// at the offset (dx, dy, dz) from the table is air or a bookshelf. It stands in for PHP's
// Position (this package can't import block or world).
type TableSurroundings func(dx, dy, dz int) (air, bookshelf bool)

// GenerateSeed is a port of EnchantingHelper::generateSeed: a new random seed for enchant option
// randomization, in the signed 32-bit range.
func GenerateSeed() int {
	return int(rand.Int64N(int64(binaryutils.Int32Max)-int64(binaryutils.Int32Min)+1) + int64(binaryutils.Int32Min))
}

// GenerateOptions is a port of EnchantingHelper::generateOptions.
func GenerateOptions(table TableSurroundings, input EnchantableItem, seed int) []*EnchantingOption {
	if input.IsNull() || input.HasEnchantments() {
		return nil
	}

	random := utils.NewRandom(seed)

	bookshelfCount := countBookshelves(table)
	baseRequiredLevel := random.NextRange(1, 8) + (bookshelfCount >> 1) + random.NextRange(0, bookshelfCount)
	topRequiredLevel := int(math.Floor(math.Max(float64(baseRequiredLevel)/3, 1)))
	middleRequiredLevel := int(math.Floor(float64(baseRequiredLevel)*2/3 + 1))
	bottomRequiredLevel := max(baseRequiredLevel, bookshelfCount*2)

	return []*EnchantingOption{
		createOption(random, input, topRequiredLevel),
		createOption(random, input, middleRequiredLevel),
		createOption(random, input, bottomRequiredLevel),
	}
}

// countBookshelves is a port of EnchantingHelper::countBookshelves.
func countBookshelves(table TableSurroundings) int {
	bookshelfCount := 0

	for x := -2; x <= 2; x++ {
	next:
		for z := -2; z <= 2; z++ {
			// We only check blocks at a distance of 2 blocks from the enchanting table
			if abs(x) != 2 && abs(z) != 2 {
				continue
			}

			// Ensure the space between the bookshelf stack at this X/Z and the enchanting table is empty
			for y := 0; y <= 1; y++ {
				// Calculate the coordinates of the space between the bookshelf and the enchanting table
				spaceX := max(min(x, 1), -1)
				spaceZ := max(min(z, 1), -1)
				if air, _ := table(spaceX, y, spaceZ); !air {
					continue next
				}
			}

			// Finally, check the number of bookshelves at the current position
			for y := 0; y <= 1; y++ {
				if _, bookshelf := table(x, y, z); bookshelf {
					bookshelfCount++
					if bookshelfCount == maxBookshelfCount {
						return bookshelfCount
					}
				}
			}
		}
	}

	return bookshelfCount
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// createOption is a port of EnchantingHelper::createOption.
func createOption(random *utils.Random, inputItem EnchantableItem, requiredXpLevel int) *EnchantingOption {
	enchantingPower := requiredXpLevel

	enchantability := inputItem.GetEnchantability()
	enchantingPower = enchantingPower + random.NextRange(0, enchantability>>2) + random.NextRange(0, enchantability>>2) + 1
	// Random bonus for enchanting power between 0.85 and 1.15
	bonus := 1 + (random.NextFloat()+random.NextFloat()-1)*0.15
	enchantingPower = int(math.Round(float64(enchantingPower) * bonus))

	var resultEnchantments []*EnchantmentInstance
	availableEnchantments := getAvailableEnchantments(enchantingPower, inputItem)

	lastEnchantment := getRandomWeightedEnchantment(random, availableEnchantments)
	if lastEnchantment != nil {
		resultEnchantments = append(resultEnchantments, lastEnchantment)

		// With probability (power + 1) / 50, continue adding enchantments
		for random.NextFloat() <= float64(enchantingPower+1)/50 {
			// Remove from the list of available enchantments anything that conflicts
			// with previously-chosen enchantments
			var filtered []*EnchantmentInstance
			for _, e := range availableEnchantments {
				if e.GetType() != lastEnchantment.GetType() && e.GetType().IsCompatibleWith(lastEnchantment.GetType()) {
					filtered = append(filtered, e)
				}
			}
			availableEnchantments = filtered

			lastEnchantment = getRandomWeightedEnchantment(random, availableEnchantments)
			if lastEnchantment == nil {
				break
			}

			resultEnchantments = append(resultEnchantments, lastEnchantment)
			enchantingPower >>= 1
		}
	}

	return NewEnchantingOption(requiredXpLevel, getRandomOptionName(random), resultEnchantments)
}

// getAvailableEnchantments is a port of EnchantingHelper::getAvailableEnchantments.
func getAvailableEnchantments(enchantingPower int, item EnchantableItem) []*EnchantmentInstance {
	var list []*EnchantmentInstance

	for _, enchantment := range GetAvailableEnchantmentRegistry().GetPrimaryEnchantmentsForItem(item) {
		for lvl := enchantment.GetMaxLevel(); lvl > 0; lvl-- {
			if enchantingPower >= enchantment.GetMinEnchantingPower(lvl) &&
				enchantingPower <= enchantment.GetMaxEnchantingPower(lvl) {
				list = append(list, NewEnchantmentInstance(enchantment, lvl))
				break
			}
		}
	}

	return list
}

// getRandomWeightedEnchantment is a port of EnchantingHelper::getRandomWeightedEnchantment.
func getRandomWeightedEnchantment(random *utils.Random, enchantments []*EnchantmentInstance) *EnchantmentInstance {
	if len(enchantments) == 0 {
		return nil
	}

	totalWeight := 0
	for _, enchantment := range enchantments {
		totalWeight += enchantment.GetType().GetRarity()
	}

	var result *EnchantmentInstance
	randomWeight := random.NextRange(1, totalWeight)

	for _, enchantment := range enchantments {
		randomWeight -= enchantment.GetType().GetRarity()

		if randomWeight <= 0 {
			result = enchantment
			break
		}
	}

	return result
}

// getRandomOptionName is a port of EnchantingHelper::getRandomOptionName.
func getRandomOptionName(random *utils.Random) string {
	name := make([]byte, 0, 15)
	for i := random.NextRange(5, 15); i > 0; i-- {
		name = append(name, byte(random.NextRange('a', 'z')&0xff))
	}

	return string(name)
}
