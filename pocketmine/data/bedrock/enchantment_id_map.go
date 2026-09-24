package bedrock

import (
	"sync"

	"pocketmine-go/pocketmine/item/enchantment"
)

// EnchantmentIds constants, a port of pocketmine\data\bedrock\EnchantmentIds.
const (
	EnchantmentIDProtection           = 0
	EnchantmentIDFireProtection       = 1
	EnchantmentIDFeatherFalling       = 2
	EnchantmentIDBlastProtection      = 3
	EnchantmentIDProjectileProtection = 4
	EnchantmentIDThorns               = 5
	EnchantmentIDRespiration          = 6
	EnchantmentIDDepthStrider         = 7
	EnchantmentIDAquaAffinity         = 8
	EnchantmentIDSharpness            = 9
	EnchantmentIDSmite                = 10
	EnchantmentIDBaneOfArthropods     = 11
	EnchantmentIDKnockback            = 12
	EnchantmentIDFireAspect           = 13
	EnchantmentIDLooting              = 14
	EnchantmentIDEfficiency           = 15
	EnchantmentIDSilkTouch            = 16
	EnchantmentIDUnbreaking           = 17
	EnchantmentIDFortune              = 18
	EnchantmentIDPower                = 19
	EnchantmentIDPunch                = 20
	EnchantmentIDFlame                = 21
	EnchantmentIDInfinity             = 22
	EnchantmentIDLuckOfTheSea         = 23
	EnchantmentIDLure                 = 24
	EnchantmentIDFrostWalker          = 25
	EnchantmentIDMending              = 26
	EnchantmentIDBinding              = 27
	EnchantmentIDVanishing            = 28
	EnchantmentIDImpaling             = 29
	EnchantmentIDRiptide              = 30
	EnchantmentIDLoyalty              = 31
	EnchantmentIDChanneling           = 32
	EnchantmentIDMultishot            = 33
	EnchantmentIDPiercing             = 34
	EnchantmentIDQuickCharge          = 35
	EnchantmentIDSoulSpeed            = 36
	EnchantmentIDSwiftSneak           = 37
)

var (
	enchantmentIDMapOnce sync.Once
	enchantmentIDMap     *IntSaveIdMap[enchantment.Enchantment]
)

// EnchantmentIdMap is the port of EnchantmentIdMap::getInstance().
func EnchantmentIdMap() *IntSaveIdMap[enchantment.Enchantment] {
	enchantmentIDMapOnce.Do(func() {
		m := newIntSaveIdMap[enchantment.Enchantment]()
		m.Register(EnchantmentIDProtection, enchantment.VanillaProtection())
		m.Register(EnchantmentIDFireProtection, enchantment.VanillaFireProtection())
		m.Register(EnchantmentIDFeatherFalling, enchantment.VanillaFeatherFalling())
		m.Register(EnchantmentIDBlastProtection, enchantment.VanillaBlastProtection())
		m.Register(EnchantmentIDProjectileProtection, enchantment.VanillaProjectileProtection())
		m.Register(EnchantmentIDThorns, enchantment.VanillaThorns())
		m.Register(EnchantmentIDRespiration, enchantment.VanillaRespiration())
		m.Register(EnchantmentIDAquaAffinity, enchantment.VanillaAquaAffinity())
		m.Register(EnchantmentIDSharpness, enchantment.VanillaSharpness())
		//TODO: smite, bane of arthropods (these don't make sense now because their applicable mobs don't exist yet)
		m.Register(EnchantmentIDKnockback, enchantment.VanillaKnockback())
		m.Register(EnchantmentIDFireAspect, enchantment.VanillaFireAspect())
		m.Register(EnchantmentIDEfficiency, enchantment.VanillaEfficiency())
		m.Register(EnchantmentIDFortune, enchantment.VanillaFortune())
		m.Register(EnchantmentIDSilkTouch, enchantment.VanillaSilkTouch())
		m.Register(EnchantmentIDUnbreaking, enchantment.VanillaUnbreaking())
		m.Register(EnchantmentIDPower, enchantment.VanillaPower())
		m.Register(EnchantmentIDPunch, enchantment.VanillaPunch())
		m.Register(EnchantmentIDFlame, enchantment.VanillaFlame())
		m.Register(EnchantmentIDInfinity, enchantment.VanillaInfinity())
		m.Register(EnchantmentIDMending, enchantment.VanillaMending())
		m.Register(EnchantmentIDVanishing, enchantment.VanillaVanishing())
		m.Register(EnchantmentIDSwiftSneak, enchantment.VanillaSwiftSneak())
		m.Register(EnchantmentIDFrostWalker, enchantment.VanillaFrostWalker())
		enchantmentIDMap = m
	})
	return enchantmentIDMap
}
