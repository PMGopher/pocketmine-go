package block

import (
	"math/rand"

	"pocketmine-go/pocketmine/item/enchantment"
)

// This file is a port of pocketmine\block\utils\FortuneDropHelper (in the block package, like
// BlockEventHelper, since its callers are blocks and it needs the block Item interface).

// fortuneLevel is $usedItem->getEnchantmentLevel(VanillaEnchantments::FORTUNE()).
func fortuneLevel(usedItem Item) int {
	if leveled, ok := usedItem.(interface {
		GetEnchantmentLevel(e enchantment.Enchantment) int
	}); ok {
		return leveled.GetEnchantmentLevel(enchantment.VanillaFortune())
	}
	return 0
}

// mtRand is PHP's mt_rand($min, $max) (inclusive).
func mtRand(min, max int) int { return min + rand.Intn(max-min+1) }

// FortuneWeighted is a port of FortuneDropHelper::weighted: if a random number between 0-1 is
// greater than 2/(level+2), the max drop amount is multiplied by level+1; a random amount between
// min and the (possibly multiplied) max is picked. Panics if maxBase < min.
func FortuneWeighted(usedItem Item, min, maxBase int) int {
	if maxBase < min {
		panic("Maximum drop amount must be greater than or equal to minimum drop amount")
	}
	level := fortuneLevel(usedItem)
	max := maxBase
	if level > 0 && rand.Float64() > 2/float64(level+2) {
		max = maxBase * (level + 1)
	}
	return mtRand(min, max)
}

// FortuneBinomial is a port of FortuneDropHelper::binomial: maxBase+level rolls, each adding 1
// with the given chance (PHP's defaults are minRolls 3 and chance 4/7).
func FortuneBinomial(usedItem Item, min, minRolls int, chance float64) int {
	count := min
	rolls := minRolls + fortuneLevel(usedItem)
	for i := 0; i < rolls; i++ {
		if rand.Float64() < chance {
			count++
		}
	}
	return count
}

// FortuneDiscrete is a port of FortuneDropHelper::discrete: a random amount between min and
// maxBase + the fortune level. Panics if maxBase < min.
func FortuneDiscrete(usedItem Item, min, maxBase int) int {
	if maxBase < min {
		panic("Minimum base drop amount must be less than or equal to maximum base drop amount")
	}
	return mtRand(min, maxBase+fortuneLevel(usedItem))
}

// FortuneBonusChanceDivisor is a port of FortuneDropHelper::bonusChanceDivisor: a 1 in
// (divisorBase - level * divisorSubtractPerLevel) chance of a bonus drop.
func FortuneBonusChanceDivisor(usedItem Item, divisorBase, divisorSubtractPerLevel int) bool {
	return mtRand(1, max(1, divisorBase-fortuneLevel(usedItem)*divisorSubtractPerLevel)) == 1
}

// FortuneBonusChanceFixed is a port of FortuneDropHelper::bonusChanceFixed.
func FortuneBonusChanceFixed(usedItem Item, chanceBase, addedChancePerLevel float64) bool {
	chance := min(1, chanceBase+float64(fortuneLevel(usedItem))*addedChancePerLevel)
	return rand.Float64() < chance
}

// noFortuneItem is VanillaItems::AIR() as the item passed to FortuneDropHelper when a drop mustn't
// be affected by Fortune (Stem).
type noFortuneItem struct{ Item }
