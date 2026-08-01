package item

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
