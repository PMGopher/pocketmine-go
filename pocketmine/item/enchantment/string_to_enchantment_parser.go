package enchantment

import (
	"sync"

	"pocketmine-go/pocketmine/utils"
)

// StringToEnchantmentParser is a port of pocketmine\item\enchantment\StringToEnchantmentParser -
// handles parsing enchantments from strings. This is used to interpret names in the /enchant
// command.
type StringToEnchantmentParser struct {
	*utils.StringToTParser[Enchantment]
}

var (
	stringToEnchantmentParserOnce     sync.Once
	stringToEnchantmentParserInstance *StringToEnchantmentParser
)

// GetStringToEnchantmentParser is the port of StringToEnchantmentParser::getInstance().
func GetStringToEnchantmentParser() *StringToEnchantmentParser {
	stringToEnchantmentParserOnce.Do(func() {
		p := &StringToEnchantmentParser{StringToTParser: utils.NewStringToTParser[Enchantment]()}
		_ = p.Register("blast_protection", func(string) Enchantment { return VanillaBlastProtection() })
		_ = p.Register("efficiency", func(string) Enchantment { return VanillaEfficiency() })
		_ = p.Register("feather_falling", func(string) Enchantment { return VanillaFeatherFalling() })
		_ = p.Register("fire_aspect", func(string) Enchantment { return VanillaFireAspect() })
		_ = p.Register("fire_protection", func(string) Enchantment { return VanillaFireProtection() })
		_ = p.Register("flame", func(string) Enchantment { return VanillaFlame() })
		_ = p.Register("fortune", func(string) Enchantment { return VanillaFortune() })
		_ = p.Register("frost_walker", func(string) Enchantment { return VanillaFrostWalker() })
		_ = p.Register("infinity", func(string) Enchantment { return VanillaInfinity() })
		_ = p.Register("knockback", func(string) Enchantment { return VanillaKnockback() })
		_ = p.Register("mending", func(string) Enchantment { return VanillaMending() })
		_ = p.Register("power", func(string) Enchantment { return VanillaPower() })
		_ = p.Register("projectile_protection", func(string) Enchantment { return VanillaProjectileProtection() })
		_ = p.Register("protection", func(string) Enchantment { return VanillaProtection() })
		_ = p.Register("punch", func(string) Enchantment { return VanillaPunch() })
		_ = p.Register("respiration", func(string) Enchantment { return VanillaRespiration() })
		_ = p.Register("aqua_affinity", func(string) Enchantment { return VanillaAquaAffinity() })
		_ = p.Register("sharpness", func(string) Enchantment { return VanillaSharpness() })
		_ = p.Register("silk_touch", func(string) Enchantment { return VanillaSilkTouch() })
		_ = p.Register("swift_sneak", func(string) Enchantment { return VanillaSwiftSneak() })
		_ = p.Register("thorns", func(string) Enchantment { return VanillaThorns() })
		_ = p.Register("unbreaking", func(string) Enchantment { return VanillaUnbreaking() })
		_ = p.Register("vanishing", func(string) Enchantment { return VanillaVanishing() })
		stringToEnchantmentParserInstance = p
	})
	return stringToEnchantmentParserInstance
}
