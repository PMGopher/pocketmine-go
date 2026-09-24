package item

import "pocketmine-go/pocketmine/entity/effect"

// PotionType is a port of pocketmine\item\PotionType. GetEffects (the actual potion effects)
// isn't ported - it needs EffectInstance (entity/effect package, not ported), same gap
// documented throughout this port wherever an EffectInstance would be constructed. Only
// GetDisplayName is ported, since that's plain string data.
type PotionType int

const (
	PotionTypeWater PotionType = iota
	PotionTypeMundane
	PotionTypeLongMundane
	PotionTypeThick
	PotionTypeAwkward
	PotionTypeNightVision
	PotionTypeLongNightVision
	PotionTypeInvisibility
	PotionTypeLongInvisibility
	PotionTypeLeaping
	PotionTypeLongLeaping
	PotionTypeStrongLeaping
	PotionTypeFireResistance
	PotionTypeLongFireResistance
	PotionTypeSwiftness
	PotionTypeLongSwiftness
	PotionTypeStrongSwiftness
	PotionTypeSlowness
	PotionTypeLongSlowness
	PotionTypeWaterBreathing
	PotionTypeLongWaterBreathing
	PotionTypeHealing
	PotionTypeStrongHealing
	PotionTypeHarming
	PotionTypeStrongHarming
	PotionTypePoison
	PotionTypeLongPoison
	PotionTypeStrongPoison
	PotionTypeRegeneration
	PotionTypeLongRegeneration
	PotionTypeStrongRegeneration
	PotionTypeStrength
	PotionTypeLongStrength
	PotionTypeStrongStrength
	PotionTypeWeakness
	PotionTypeLongWeakness
	PotionTypeWither
	PotionTypeTurtleMaster
	PotionTypeLongTurtleMaster
	PotionTypeStrongTurtleMaster
	PotionTypeSlowFalling
	PotionTypeLongSlowFalling
	PotionTypeStrongSlowness
)

var potionTypeDisplayNames = map[PotionType]string{
	PotionTypeWater:              "Water",
	PotionTypeMundane:            "Mundane",
	PotionTypeLongMundane:        "Long Mundane",
	PotionTypeThick:              "Thick",
	PotionTypeAwkward:            "Awkward",
	PotionTypeNightVision:        "Night Vision",
	PotionTypeLongNightVision:    "Long Night Vision",
	PotionTypeInvisibility:       "Invisibility",
	PotionTypeLongInvisibility:   "Long Invisibility",
	PotionTypeLeaping:            "Leaping",
	PotionTypeLongLeaping:        "Long Leaping",
	PotionTypeStrongLeaping:      "Strong Leaping",
	PotionTypeFireResistance:     "Fire Resistance",
	PotionTypeLongFireResistance: "Long Fire Resistance",
	PotionTypeSwiftness:          "Swiftness",
	PotionTypeLongSwiftness:      "Long Swiftness",
	PotionTypeStrongSwiftness:    "Strong Swiftness",
	PotionTypeSlowness:           "Slowness",
	PotionTypeLongSlowness:       "Long Slowness",
	PotionTypeWaterBreathing:     "Water Breathing",
	PotionTypeLongWaterBreathing: "Long Water Breathing",
	PotionTypeHealing:            "Healing",
	PotionTypeStrongHealing:      "Strong Healing",
	PotionTypeHarming:            "Harming",
	PotionTypeStrongHarming:      "Strong Harming",
	PotionTypePoison:             "Poison",
	PotionTypeLongPoison:         "Long Poison",
	PotionTypeStrongPoison:       "Strong Poison",
	PotionTypeRegeneration:       "Regeneration",
	PotionTypeLongRegeneration:   "Long Regeneration",
	PotionTypeStrongRegeneration: "Strong Regeneration",
	PotionTypeStrength:           "Strength",
	PotionTypeLongStrength:       "Long Strength",
	PotionTypeStrongStrength:     "Strong Strength",
	PotionTypeWeakness:           "Weakness",
	PotionTypeLongWeakness:       "Long Weakness",
	PotionTypeWither:             "Wither",
	PotionTypeTurtleMaster:       "Turtle Master",
	PotionTypeLongTurtleMaster:   "Long Turtle Master",
	PotionTypeStrongTurtleMaster: "Strong Turtle Master",
	PotionTypeSlowFalling:        "Slow Falling",
	PotionTypeLongSlowFalling:    "Long Slow Falling",
	PotionTypeStrongSlowness:     "Strong Slowness",
}

func (t PotionType) GetDisplayName() string { return potionTypeDisplayNames[t] }

// GetEffects is a port of PotionType::getEffects: fresh EffectInstances every call (PHP rebuilds
// them from a closure each time, so callers can freely modify the result).
func (t PotionType) GetEffects() []*effect.EffectInstance {
	with := effect.NewEffectInstanceWith
	switch t {
	case PotionTypeNightVision:
		return []*effect.EffectInstance{with(effect.VanillaNightVision(), 3600, 0)}
	case PotionTypeLongNightVision:
		return []*effect.EffectInstance{with(effect.VanillaNightVision(), 9600, 0)}
	case PotionTypeInvisibility:
		return []*effect.EffectInstance{with(effect.VanillaInvisibility(), 3600, 0)}
	case PotionTypeLongInvisibility:
		return []*effect.EffectInstance{with(effect.VanillaInvisibility(), 9600, 0)}
	case PotionTypeLeaping:
		return []*effect.EffectInstance{with(effect.VanillaJumpBoost(), 3600, 0)}
	case PotionTypeLongLeaping:
		return []*effect.EffectInstance{with(effect.VanillaJumpBoost(), 9600, 0)}
	case PotionTypeStrongLeaping:
		return []*effect.EffectInstance{with(effect.VanillaJumpBoost(), 1800, 1)}
	case PotionTypeFireResistance:
		return []*effect.EffectInstance{with(effect.VanillaFireResistance(), 3600, 0)}
	case PotionTypeLongFireResistance:
		return []*effect.EffectInstance{with(effect.VanillaFireResistance(), 9600, 0)}
	case PotionTypeSwiftness:
		return []*effect.EffectInstance{with(effect.VanillaSpeed(), 3600, 0)}
	case PotionTypeLongSwiftness:
		return []*effect.EffectInstance{with(effect.VanillaSpeed(), 9600, 0)}
	case PotionTypeStrongSwiftness:
		return []*effect.EffectInstance{with(effect.VanillaSpeed(), 1800, 1)}
	case PotionTypeSlowness:
		return []*effect.EffectInstance{with(effect.VanillaSlowness(), 1800, 0)}
	case PotionTypeLongSlowness:
		return []*effect.EffectInstance{with(effect.VanillaSlowness(), 4800, 0)}
	case PotionTypeWaterBreathing:
		return []*effect.EffectInstance{with(effect.VanillaWaterBreathing(), 3600, 0)}
	case PotionTypeLongWaterBreathing:
		return []*effect.EffectInstance{with(effect.VanillaWaterBreathing(), 9600, 0)}
	case PotionTypeHealing:
		return []*effect.EffectInstance{effect.NewEffectInstance(effect.VanillaInstantHealth())}
	case PotionTypeStrongHealing:
		return []*effect.EffectInstance{effect.NewEffectInstanceFull(effect.VanillaInstantHealth(), nil, 1, true, false, nil, false)}
	case PotionTypeHarming:
		return []*effect.EffectInstance{effect.NewEffectInstance(effect.VanillaInstantDamage())}
	case PotionTypeStrongHarming:
		return []*effect.EffectInstance{effect.NewEffectInstanceFull(effect.VanillaInstantDamage(), nil, 1, true, false, nil, false)}
	case PotionTypePoison:
		return []*effect.EffectInstance{with(effect.VanillaPoison(), 900, 0)}
	case PotionTypeLongPoison:
		return []*effect.EffectInstance{with(effect.VanillaPoison(), 2400, 0)}
	case PotionTypeStrongPoison:
		return []*effect.EffectInstance{with(effect.VanillaPoison(), 440, 1)}
	case PotionTypeRegeneration:
		return []*effect.EffectInstance{with(effect.VanillaRegeneration(), 900, 0)}
	case PotionTypeLongRegeneration:
		return []*effect.EffectInstance{with(effect.VanillaRegeneration(), 2400, 0)}
	case PotionTypeStrongRegeneration:
		return []*effect.EffectInstance{with(effect.VanillaRegeneration(), 440, 1)}
	case PotionTypeStrength:
		return []*effect.EffectInstance{with(effect.VanillaStrength(), 3600, 0)}
	case PotionTypeLongStrength:
		return []*effect.EffectInstance{with(effect.VanillaStrength(), 9600, 0)}
	case PotionTypeStrongStrength:
		return []*effect.EffectInstance{with(effect.VanillaStrength(), 1800, 1)}
	case PotionTypeWeakness:
		return []*effect.EffectInstance{with(effect.VanillaWeakness(), 1800, 0)}
	case PotionTypeLongWeakness:
		return []*effect.EffectInstance{with(effect.VanillaWeakness(), 4800, 0)}
	case PotionTypeWither:
		return []*effect.EffectInstance{with(effect.VanillaWither(), 800, 1)}
	case PotionTypeTurtleMaster:
		return []*effect.EffectInstance{with(effect.VanillaSlowness(), 20*20, 3), with(effect.VanillaResistance(), 20*20, 2)}
	case PotionTypeLongTurtleMaster:
		return []*effect.EffectInstance{with(effect.VanillaSlowness(), 40*20, 3), with(effect.VanillaResistance(), 40*20, 2)}
	case PotionTypeStrongTurtleMaster:
		return []*effect.EffectInstance{with(effect.VanillaSlowness(), 20*20, 5), with(effect.VanillaResistance(), 20*20, 3)}
	case PotionTypeStrongSlowness:
		return []*effect.EffectInstance{with(effect.VanillaSlowness(), 20*20, 3)}
	}
	// Water, Mundane, Long Mundane, Thick, Awkward have no effects; Slow Falling / Long Slow
	// Falling are TODO upstream (the slow_falling effect isn't registered).
	return nil
}

// PotionTypeIdMap is a port of pocketmine\data\bedrock\PotionTypeIdMap. It lives here rather than
// in data/bedrock because this package already imports data/bedrock (for EnchantmentIdMap), so the
// reverse import would be a cycle. PotionType's Go values are declared in exactly
// PotionTypeIds' order (WATER = 0 ... STRONG_SLOWNESS = 42), so the saved ID is the enum value.
type potionTypeIdMap struct{}

// PotionTypeIdMapInstance is the port of PotionTypeIdMap::getInstance().
var PotionTypeIdMapInstance potionTypeIdMap

// FromID returns the potion type saved as id (false if unknown).
func (potionTypeIdMap) FromID(id int) (PotionType, bool) {
	if id < int(PotionTypeWater) || id > int(PotionTypeStrongSlowness) {
		return 0, false
	}
	return PotionType(id), true
}

// ToID returns the save ID of t.
func (potionTypeIdMap) ToID(t PotionType) int { return int(t) }
