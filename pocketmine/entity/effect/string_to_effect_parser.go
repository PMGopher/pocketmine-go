package effect

import (
	"sync"

	"pocketmine-go/pocketmine/utils"
)

// StringToEffectParser is a port of pocketmine\entity\effect\StringToEffectParser - handles
// parsing effect types from strings. This is used to interpret names in the /effect command.
type StringToEffectParser struct {
	*utils.StringToTParser[Effect]
}

var (
	stringToEffectParserOnce     sync.Once
	stringToEffectParserInstance *StringToEffectParser
)

// GetStringToEffectParser is the port of StringToEffectParser::getInstance().
func GetStringToEffectParser() *StringToEffectParser {
	stringToEffectParserOnce.Do(func() {
		p := &StringToEffectParser{StringToTParser: utils.NewStringToTParser[Effect]()}
		_ = p.Register("absorption", func(string) Effect { return VanillaAbsorption() })
		_ = p.Register("blindness", func(string) Effect { return VanillaBlindness() })
		_ = p.Register("conduit_power", func(string) Effect { return VanillaConduitPower() })
		_ = p.Register("darkness", func(string) Effect { return VanillaDarkness() })
		_ = p.Register("fatal_poison", func(string) Effect { return VanillaFatalPoison() })
		_ = p.Register("fire_resistance", func(string) Effect { return VanillaFireResistance() })
		_ = p.Register("haste", func(string) Effect { return VanillaHaste() })
		_ = p.Register("health_boost", func(string) Effect { return VanillaHealthBoost() })
		_ = p.Register("hunger", func(string) Effect { return VanillaHunger() })
		_ = p.Register("instant_damage", func(string) Effect { return VanillaInstantDamage() })
		_ = p.Register("instant_health", func(string) Effect { return VanillaInstantHealth() })
		_ = p.Register("invisibility", func(string) Effect { return VanillaInvisibility() })
		_ = p.Register("jump_boost", func(string) Effect { return VanillaJumpBoost() })
		_ = p.Register("levitation", func(string) Effect { return VanillaLevitation() })
		_ = p.Register("mining_fatigue", func(string) Effect { return VanillaMiningFatigue() })
		_ = p.Register("nausea", func(string) Effect { return VanillaNausea() })
		_ = p.Register("night_vision", func(string) Effect { return VanillaNightVision() })
		_ = p.Register("poison", func(string) Effect { return VanillaPoison() })
		_ = p.Register("regeneration", func(string) Effect { return VanillaRegeneration() })
		_ = p.Register("resistance", func(string) Effect { return VanillaResistance() })
		_ = p.Register("saturation", func(string) Effect { return VanillaSaturation() })
		_ = p.Register("slowness", func(string) Effect { return VanillaSlowness() })
		_ = p.Register("speed", func(string) Effect { return VanillaSpeed() })
		_ = p.Register("strength", func(string) Effect { return VanillaStrength() })
		_ = p.Register("water_breathing", func(string) Effect { return VanillaWaterBreathing() })
		_ = p.Register("weakness", func(string) Effect { return VanillaWeakness() })
		_ = p.Register("wither", func(string) Effect { return VanillaWither() })
		stringToEffectParserInstance = p
	})
	return stringToEffectParserInstance
}
