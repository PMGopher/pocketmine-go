package effect

import (
	"sync"

	"pocketmine-go/pocketmine/color"
	"pocketmine-go/pocketmine/lang"
)

// VanillaEffects is a port of pocketmine\entity\effect\VanillaEffects (generated in PHP from
// VanillaEffectsInputs.php). Every getter returns the same singleton instance: effect types are
// compared by identity (see Effect's doc comment), unlike blocks/items which hand out clones.
//
// Every name, color, and flag below is copied from VanillaEffectsInputs.php. The TODO entries there
// (bad_omen, slow_falling, village_hero) aren't registered upstream either.

func tr(key string) *lang.Translatable { return lang.NewTranslatable(key, nil) }

var (
	vanillaEffectsOnce    sync.Once
	vanillaEffects        map[string]Effect
	vanillaEffectOrder    []string
	vanillaAbsorption     Effect
	vanillaBlindness      Effect
	vanillaConduitPower   Effect
	vanillaDarkness       Effect
	vanillaFatalPoison    Effect
	vanillaFireResistance Effect
	vanillaHaste          Effect
	vanillaHealthBoost    Effect
	vanillaHunger         Effect
	vanillaInstantDamage  Effect
	vanillaInstantHealth  Effect
	vanillaInvisibility   Effect
	vanillaJumpBoost      Effect
	vanillaLevitation     Effect
	vanillaMiningFatigue  Effect
	vanillaNausea         Effect
	vanillaNightVision    Effect
	vanillaPoison         Effect
	vanillaRegeneration   Effect
	vanillaResistance     Effect
	vanillaSaturation     Effect
	vanillaSlowness       Effect
	vanillaSpeed          Effect
	vanillaStrength       Effect
	vanillaWaterBreathing Effect
	vanillaWeakness       Effect
	vanillaWither         Effect
)

func setupVanillaEffects() {
	vanillaEffectsOnce.Do(func() {
		vanillaEffects = map[string]Effect{}
		register := func(name string, e Effect) Effect {
			vanillaEffects[name] = e
			vanillaEffectOrder = append(vanillaEffectOrder, name)
			return e
		}
		vanillaAbsorption = register("absorption", NewAbsorptionEffect(tr("potion.absorption"), color.NewColor(0x25, 0x52, 0xa5)))
		vanillaBlindness = register("blindness", NewEffect(tr("potion.blindness"), color.NewColor(0x1f, 0x1f, 0x23), true, 600, true))
		vanillaConduitPower = register("conduit_power", NewEffect(tr("potion.conduitPower"), color.NewColor(0x1d, 0xc2, 0xd1), false, 600, true))
		vanillaDarkness = register("darkness", NewEffect(tr("effect.darkness"), color.NewColor(0x29, 0x27, 0x21), true, 600, false))
		vanillaFatalPoison = register("fatal_poison", NewPoisonEffect(tr("potion.poison"), color.NewColor(0x4e, 0x93, 0x31), true, 600, true, true))
		vanillaFireResistance = register("fire_resistance", NewEffect(tr("potion.fireResistance"), color.NewColor(0xe4, 0x9a, 0x3a), false, 600, true))
		vanillaHaste = register("haste", NewEffect(tr("potion.digSpeed"), color.NewColor(0xd9, 0xc0, 0x43), false, 600, true))
		vanillaHealthBoost = register("health_boost", NewHealthBoostEffect(tr("potion.healthBoost"), color.NewColor(0xf8, 0x7d, 0x23)))
		vanillaHunger = register("hunger", NewHungerEffect(tr("potion.hunger"), color.NewColor(0x58, 0x76, 0x53), true))
		vanillaInstantDamage = register("instant_damage", NewInstantDamageEffect(tr("potion.harm"), color.NewColor(0x43, 0x0a, 0x09), true, false))
		vanillaInstantHealth = register("instant_health", NewInstantHealthEffect(tr("potion.heal"), color.NewColor(0xf8, 0x24, 0x23), false, false))
		vanillaInvisibility = register("invisibility", NewInvisibilityEffect(tr("potion.invisibility"), color.NewColor(0x7f, 0x83, 0x92)))
		vanillaJumpBoost = register("jump_boost", NewEffect(tr("potion.jump"), color.NewColor(0x22, 0xff, 0x4c), false, 600, true))
		vanillaLevitation = register("levitation", NewLevitationEffect(tr("potion.levitation"), color.NewColor(0xce, 0xff, 0xff)))
		vanillaMiningFatigue = register("mining_fatigue", NewEffect(tr("potion.digSlowDown"), color.NewColor(0x4a, 0x42, 0x17), true, 600, true))
		vanillaNausea = register("nausea", NewEffect(tr("potion.confusion"), color.NewColor(0x55, 0x1d, 0x4a), true, 600, true))
		vanillaNightVision = register("night_vision", NewEffect(tr("potion.nightVision"), color.NewColor(0x1f, 0x1f, 0xa1), false, 600, true))
		vanillaPoison = register("poison", NewPoisonEffect(tr("potion.poison"), color.NewColor(0x4e, 0x93, 0x31), true, 600, true, false))
		vanillaRegeneration = register("regeneration", NewRegenerationEffect(tr("potion.regeneration"), color.NewColor(0xcd, 0x5c, 0xab)))
		vanillaResistance = register("resistance", NewEffect(tr("potion.resistance"), color.NewColor(0x99, 0x45, 0x3a), false, 600, true))
		vanillaSaturation = register("saturation", NewSaturationEffect(tr("potion.saturation"), color.NewColor(0xf8, 0x24, 0x23)))
		vanillaSlowness = register("slowness", NewSlownessEffect(tr("potion.moveSlowdown"), color.NewColor(0x5a, 0x6c, 0x81), true))
		vanillaSpeed = register("speed", NewSpeedEffect(tr("potion.moveSpeed"), color.NewColor(0x7c, 0xaf, 0xc6)))
		vanillaStrength = register("strength", NewEffect(tr("potion.damageBoost"), color.NewColor(0x93, 0x24, 0x23), false, 600, true))
		vanillaWaterBreathing = register("water_breathing", NewEffect(tr("potion.waterBreathing"), color.NewColor(0x2e, 0x52, 0x99), false, 600, true))
		vanillaWeakness = register("weakness", NewEffect(tr("potion.weakness"), color.NewColor(0x48, 0x4d, 0x48), true, 600, true))
		vanillaWither = register("wither", NewWitherEffect(tr("potion.wither"), color.NewColor(0x35, 0x2a, 0x27), true))
	})
}

// GetAllVanillaEffects is a port of VanillaEffects::getAll(), keyed by upper-case registry name
// like PHP (e.g. "FIRE_RESISTANCE").
func GetAllVanillaEffects() map[string]Effect {
	setupVanillaEffects()
	result := make(map[string]Effect, len(vanillaEffects))
	for name, e := range vanillaEffects {
		result[upperName(name)] = e
	}
	return result
}

func upperName(name string) string {
	b := []byte(name)
	for i, c := range b {
		if c >= 'a' && c <= 'z' {
			b[i] = c - 'a' + 'A'
		}
	}
	return string(b)
}

func VanillaAbsorption() Effect {
	setupVanillaEffects()
	return vanillaAbsorption
}

func VanillaBlindness() Effect {
	setupVanillaEffects()
	return vanillaBlindness
}

func VanillaConduitPower() Effect {
	setupVanillaEffects()
	return vanillaConduitPower
}

func VanillaDarkness() Effect {
	setupVanillaEffects()
	return vanillaDarkness
}

func VanillaFatalPoison() Effect {
	setupVanillaEffects()
	return vanillaFatalPoison
}

func VanillaFireResistance() Effect {
	setupVanillaEffects()
	return vanillaFireResistance
}

func VanillaHaste() Effect {
	setupVanillaEffects()
	return vanillaHaste
}

func VanillaHealthBoost() Effect {
	setupVanillaEffects()
	return vanillaHealthBoost
}

func VanillaHunger() Effect {
	setupVanillaEffects()
	return vanillaHunger
}

func VanillaInstantDamage() Effect {
	setupVanillaEffects()
	return vanillaInstantDamage
}

func VanillaInstantHealth() Effect {
	setupVanillaEffects()
	return vanillaInstantHealth
}

func VanillaInvisibility() Effect {
	setupVanillaEffects()
	return vanillaInvisibility
}

func VanillaJumpBoost() Effect {
	setupVanillaEffects()
	return vanillaJumpBoost
}

func VanillaLevitation() Effect {
	setupVanillaEffects()
	return vanillaLevitation
}

func VanillaMiningFatigue() Effect {
	setupVanillaEffects()
	return vanillaMiningFatigue
}

func VanillaNausea() Effect {
	setupVanillaEffects()
	return vanillaNausea
}

func VanillaNightVision() Effect {
	setupVanillaEffects()
	return vanillaNightVision
}

func VanillaPoison() Effect {
	setupVanillaEffects()
	return vanillaPoison
}

func VanillaRegeneration() Effect {
	setupVanillaEffects()
	return vanillaRegeneration
}

func VanillaResistance() Effect {
	setupVanillaEffects()
	return vanillaResistance
}

func VanillaSaturation() Effect {
	setupVanillaEffects()
	return vanillaSaturation
}

func VanillaSlowness() Effect {
	setupVanillaEffects()
	return vanillaSlowness
}

func VanillaSpeed() Effect {
	setupVanillaEffects()
	return vanillaSpeed
}

func VanillaStrength() Effect {
	setupVanillaEffects()
	return vanillaStrength
}

func VanillaWaterBreathing() Effect {
	setupVanillaEffects()
	return vanillaWaterBreathing
}

func VanillaWeakness() Effect {
	setupVanillaEffects()
	return vanillaWeakness
}

func VanillaWither() Effect {
	setupVanillaEffects()
	return vanillaWither
}
