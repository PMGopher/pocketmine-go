package bedrock

import (
	"sync"

	"pocketmine-go/pocketmine/entity/effect"
)

// EffectIds constants, a port of pocketmine\data\bedrock\EffectIds.
const (
	EffectIDSpeed          = 1
	EffectIDSlowness       = 2
	EffectIDHaste          = 3
	EffectIDMiningFatigue  = 4
	EffectIDStrength       = 5
	EffectIDInstantHealth  = 6
	EffectIDInstantDamage  = 7
	EffectIDJumpBoost      = 8
	EffectIDNausea         = 9
	EffectIDRegeneration   = 10
	EffectIDResistance     = 11
	EffectIDFireResistance = 12
	EffectIDWaterBreathing = 13
	EffectIDInvisibility   = 14
	EffectIDBlindness      = 15
	EffectIDNightVision    = 16
	EffectIDHunger         = 17
	EffectIDWeakness       = 18
	EffectIDPoison         = 19
	EffectIDWither         = 20
	EffectIDHealthBoost    = 21
	EffectIDAbsorption     = 22
	EffectIDSaturation     = 23
	EffectIDLevitation     = 24
	EffectIDFatalPoison    = 25
	EffectIDConduitPower   = 26
	EffectIDSlowFalling    = 27
	EffectIDBadOmen        = 28
	EffectIDVillageHero    = 29
	EffectIDDarkness       = 30
)

var (
	effectIDMapOnce sync.Once
	effectIDMap     *IntSaveIdMap[effect.Effect]
)

// EffectIdMap is the port of EffectIdMap::getInstance().
func EffectIdMap() *IntSaveIdMap[effect.Effect] {
	effectIDMapOnce.Do(func() {
		m := newIntSaveIdMap[effect.Effect]()
		m.Register(EffectIDSpeed, effect.VanillaSpeed())
		m.Register(EffectIDSlowness, effect.VanillaSlowness())
		m.Register(EffectIDHaste, effect.VanillaHaste())
		m.Register(EffectIDMiningFatigue, effect.VanillaMiningFatigue())
		m.Register(EffectIDStrength, effect.VanillaStrength())
		m.Register(EffectIDInstantHealth, effect.VanillaInstantHealth())
		m.Register(EffectIDInstantDamage, effect.VanillaInstantDamage())
		m.Register(EffectIDJumpBoost, effect.VanillaJumpBoost())
		m.Register(EffectIDNausea, effect.VanillaNausea())
		m.Register(EffectIDRegeneration, effect.VanillaRegeneration())
		m.Register(EffectIDResistance, effect.VanillaResistance())
		m.Register(EffectIDFireResistance, effect.VanillaFireResistance())
		m.Register(EffectIDWaterBreathing, effect.VanillaWaterBreathing())
		m.Register(EffectIDInvisibility, effect.VanillaInvisibility())
		m.Register(EffectIDBlindness, effect.VanillaBlindness())
		m.Register(EffectIDNightVision, effect.VanillaNightVision())
		m.Register(EffectIDHunger, effect.VanillaHunger())
		m.Register(EffectIDWeakness, effect.VanillaWeakness())
		m.Register(EffectIDPoison, effect.VanillaPoison())
		m.Register(EffectIDWither, effect.VanillaWither())
		m.Register(EffectIDHealthBoost, effect.VanillaHealthBoost())
		m.Register(EffectIDAbsorption, effect.VanillaAbsorption())
		m.Register(EffectIDSaturation, effect.VanillaSaturation())
		m.Register(EffectIDLevitation, effect.VanillaLevitation())
		m.Register(EffectIDFatalPoison, effect.VanillaFatalPoison())
		m.Register(EffectIDConduitPower, effect.VanillaConduitPower())
		//TODO: SLOW_FALLING
		//TODO: BAD_OMEN
		//TODO: VILLAGE_HERO
		m.Register(EffectIDDarkness, effect.VanillaDarkness())
		effectIDMap = m
	})
	return effectIDMap
}
