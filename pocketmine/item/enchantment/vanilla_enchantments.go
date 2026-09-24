package enchantment

import (
	"sync"

	entityevent "pocketmine-go/pocketmine/event/entity"
	"pocketmine-go/pocketmine/lang"
)

// VanillaEnchantments is a port of pocketmine\item\enchantment\VanillaEnchantments (generated in
// PHP from VanillaEnchantmentsInputs.php). Every getter returns the same singleton instance, since
// enchantments are compared by identity. Every value below is copied from
// VanillaEnchantmentsInputs.php (smite, bane of arthropods and looting are TODOs upstream too).

func tr(key string) *lang.Translatable { return lang.NewTranslatable(key, nil) }

var (
	vanillaEnchantmentsOnce     sync.Once
	vanillaEnchantments         map[string]Enchantment
	vanillaProtection           Enchantment
	vanillaFireProtection       Enchantment
	vanillaFeatherFalling       Enchantment
	vanillaBlastProtection      Enchantment
	vanillaProjectileProtection Enchantment
	vanillaThorns               Enchantment
	vanillaRespiration          Enchantment
	vanillaFrostWalker          Enchantment
	vanillaAquaAffinity         Enchantment
	vanillaSharpness            Enchantment
	vanillaKnockback            Enchantment
	vanillaFireAspect           Enchantment
	vanillaEfficiency           Enchantment
	vanillaFortune              Enchantment
	vanillaSilkTouch            Enchantment
	vanillaUnbreaking           Enchantment
	vanillaPower                Enchantment
	vanillaPunch                Enchantment
	vanillaFlame                Enchantment
	vanillaInfinity             Enchantment
	vanillaMending              Enchantment
	vanillaVanishing            Enchantment
	vanillaSwiftSneak           Enchantment
)

func setupVanillaEnchantments() {
	vanillaEnchantmentsOnce.Do(func() {
		vanillaEnchantments = map[string]Enchantment{}
		register := func(name string, e Enchantment) Enchantment {
			vanillaEnchantments[name] = e
			return e
		}
		vanillaProtection = register("PROTECTION", NewProtectionEnchantment(tr("enchantment.protect.all"), RarityCommon, 0, 0, 4, 0.75, nil, func(level int) int { return 11*(level-1) + 1 }, 20))
		vanillaFireProtection = register("FIRE_PROTECTION", NewProtectionEnchantment(tr("enchantment.protect.fire"), RarityUncommon, 0, 0, 4, 1.25, []int{entityevent.CauseFire, entityevent.CauseFireTick, entityevent.CauseLava}, func(level int) int { return 8*(level-1) + 10 }, 12))
		vanillaFeatherFalling = register("FEATHER_FALLING", NewProtectionEnchantment(tr("enchantment.protect.fall"), RarityUncommon, 0, 0, 4, 2.5, []int{entityevent.CauseFall}, func(level int) int { return 6*(level-1) + 5 }, 10))
		vanillaBlastProtection = register("BLAST_PROTECTION", NewProtectionEnchantment(tr("enchantment.protect.explosion"), RarityRare, 0, 0, 4, 1.5, []int{entityevent.CauseBlockExplosion, entityevent.CauseEntityExplosion}, func(level int) int { return 8*(level-1) + 5 }, 12))
		vanillaProjectileProtection = register("PROJECTILE_PROTECTION", NewProtectionEnchantment(tr("enchantment.protect.projectile"), RarityUncommon, 0, 0, 4, 1.5, []int{entityevent.CauseProjectile}, func(level int) int { return 6*(level-1) + 3 }, 15))
		vanillaThorns = register("THORNS", NewEnchantment(tr("enchantment.thorns"), RarityMythic, 0, 0, 3, func(level int) int { return 20*(level-1) + 10 }, 50))
		vanillaRespiration = register("RESPIRATION", NewEnchantment(tr("enchantment.oxygen"), RarityRare, 0, 0, 3, func(level int) int { return 10 * level }, 30))
		vanillaFrostWalker = register("FROST_WALKER", NewEnchantment(tr("enchantment.frostwalker"), RarityRare, 0, 0, 2, func(level int) int { return 10 * level }, 15))
		vanillaAquaAffinity = register("AQUA_AFFINITY", NewEnchantment(tr("enchantment.waterWorker"), RarityRare, 0, 0, 1, nil, 40))
		vanillaSharpness = register("SHARPNESS", NewSharpnessEnchantment(tr("enchantment.damage.all"), RarityCommon, 0, 0, 5, func(level int) int { return 11*(level-1) + 1 }, 20))
		vanillaKnockback = register("KNOCKBACK", NewKnockbackEnchantment(tr("enchantment.knockback"), RarityUncommon, 0, 0, 2, func(level int) int { return 20*(level-1) + 5 }, 50))
		vanillaFireAspect = register("FIRE_ASPECT", NewFireAspectEnchantment(tr("enchantment.fire"), RarityRare, 0, 0, 2, func(level int) int { return 20*(level-1) + 10 }, 50))
		vanillaEfficiency = register("EFFICIENCY", NewEnchantment(tr("enchantment.digging"), RarityCommon, 0, 0, 5, func(level int) int { return 10*(level-1) + 1 }, 50))
		vanillaFortune = register("FORTUNE", NewEnchantment(tr("enchantment.lootBonusDigger"), RarityRare, 0, 0, 3, func(level int) int { return 9*(level-1) + 15 }, 50))
		vanillaSilkTouch = register("SILK_TOUCH", NewEnchantment(tr("enchantment.untouching"), RarityMythic, 0, 0, 1, func(level int) int { return 15 }, 50))
		vanillaUnbreaking = register("UNBREAKING", NewEnchantment(tr("enchantment.durability"), RarityUncommon, 0, 0, 3, func(level int) int { return 8*(level-1) + 5 }, 50))
		vanillaPower = register("POWER", NewEnchantment(tr("enchantment.arrowDamage"), RarityCommon, 0, 0, 5, func(level int) int { return 10*(level-1) + 1 }, 15))
		vanillaPunch = register("PUNCH", NewEnchantment(tr("enchantment.arrowKnockback"), RarityRare, 0, 0, 2, func(level int) int { return 20*(level-1) + 12 }, 25))
		vanillaFlame = register("FLAME", NewEnchantment(tr("enchantment.arrowFire"), RarityRare, 0, 0, 1, func(level int) int { return 20 }, 30))
		vanillaInfinity = register("INFINITY", NewEnchantment(tr("enchantment.arrowInfinite"), RarityMythic, 0, 0, 1, func(level int) int { return 20 }, 30))
		vanillaMending = register("MENDING", NewEnchantment(tr("enchantment.mending"), RarityRare, 0, 0, 1, func(level int) int { return 25 }, 50))
		vanillaVanishing = register("VANISHING", NewEnchantment(tr("enchantment.curse.vanishing"), RarityMythic, 0, 0, 1, func(level int) int { return 25 }, 25))
		vanillaSwiftSneak = register("SWIFT_SNEAK", NewEnchantment(tr("enchantment.swift_sneak"), RarityMythic, 0, 0, 3, func(level int) int { return 10 * level }, 5))
	})
}

// GetAllVanillaEnchantments is a port of VanillaEnchantments::getAll(), keyed by registry name
// (e.g. "FIRE_PROTECTION").
func GetAllVanillaEnchantments() map[string]Enchantment {
	setupVanillaEnchantments()
	result := make(map[string]Enchantment, len(vanillaEnchantments))
	for k, v := range vanillaEnchantments {
		result[k] = v
	}
	return result
}

func VanillaProtection() Enchantment {
	setupVanillaEnchantments()
	return vanillaProtection
}

func VanillaFireProtection() Enchantment {
	setupVanillaEnchantments()
	return vanillaFireProtection
}

func VanillaFeatherFalling() Enchantment {
	setupVanillaEnchantments()
	return vanillaFeatherFalling
}

func VanillaBlastProtection() Enchantment {
	setupVanillaEnchantments()
	return vanillaBlastProtection
}

func VanillaProjectileProtection() Enchantment {
	setupVanillaEnchantments()
	return vanillaProjectileProtection
}

func VanillaThorns() Enchantment {
	setupVanillaEnchantments()
	return vanillaThorns
}

func VanillaRespiration() Enchantment {
	setupVanillaEnchantments()
	return vanillaRespiration
}

func VanillaFrostWalker() Enchantment {
	setupVanillaEnchantments()
	return vanillaFrostWalker
}

func VanillaAquaAffinity() Enchantment {
	setupVanillaEnchantments()
	return vanillaAquaAffinity
}

func VanillaSharpness() Enchantment {
	setupVanillaEnchantments()
	return vanillaSharpness
}

func VanillaKnockback() Enchantment {
	setupVanillaEnchantments()
	return vanillaKnockback
}

func VanillaFireAspect() Enchantment {
	setupVanillaEnchantments()
	return vanillaFireAspect
}

func VanillaEfficiency() Enchantment {
	setupVanillaEnchantments()
	return vanillaEfficiency
}

func VanillaFortune() Enchantment {
	setupVanillaEnchantments()
	return vanillaFortune
}

func VanillaSilkTouch() Enchantment {
	setupVanillaEnchantments()
	return vanillaSilkTouch
}

func VanillaUnbreaking() Enchantment {
	setupVanillaEnchantments()
	return vanillaUnbreaking
}

func VanillaPower() Enchantment {
	setupVanillaEnchantments()
	return vanillaPower
}

func VanillaPunch() Enchantment {
	setupVanillaEnchantments()
	return vanillaPunch
}

func VanillaFlame() Enchantment {
	setupVanillaEnchantments()
	return vanillaFlame
}

func VanillaInfinity() Enchantment {
	setupVanillaEnchantments()
	return vanillaInfinity
}

func VanillaMending() Enchantment {
	setupVanillaEnchantments()
	return vanillaMending
}

func VanillaVanishing() Enchantment {
	setupVanillaEnchantments()
	return vanillaVanishing
}

func VanillaSwiftSneak() Enchantment {
	setupVanillaEnchantments()
	return vanillaSwiftSneak
}
