package enchantment

import (
	"math/rand/v2"

	"pocketmine-go/pocketmine/binaryutils"
)

// GenerateSeed is a port of EnchantingHelper::generateSeed - a random enchantment seed in the
// signed 32-bit range (the rest of EnchantingHelper, the enchanting-table option generator, isn't
// ported yet - see this package's doc comment).
func GenerateSeed() int {
	return int(rand.Int64N(int64(binaryutils.Int32Max)-int64(binaryutils.Int32Min)+1) + int64(binaryutils.Int32Min))
}
