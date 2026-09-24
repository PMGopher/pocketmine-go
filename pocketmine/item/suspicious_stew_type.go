package item

import (
	"pocketmine-go/pocketmine/entity/effect"
)

// SuspiciousStewType is a port of pocketmine\item\SuspiciousStewType.
type SuspiciousStewType int

const (
	SuspiciousStewTypePoppy SuspiciousStewType = iota
	SuspiciousStewTypeCornflower
	SuspiciousStewTypeTulip
	SuspiciousStewTypeAzureBluet
	SuspiciousStewTypeLilyOfTheValley
	SuspiciousStewTypeDandelion
	SuspiciousStewTypeBlueOrchid
	SuspiciousStewTypeAllium
	SuspiciousStewTypeOxeyeDaisy
	SuspiciousStewTypeWitherRose
)

// GetEffects is a port of SuspiciousStewType::getEffects.
func (t SuspiciousStewType) GetEffects() []*effect.EffectInstance {
	with := effect.NewEffectInstanceWith
	switch t {
	case SuspiciousStewTypePoppy:
		return []*effect.EffectInstance{with(effect.VanillaNightVision(), 80, 0)}
	case SuspiciousStewTypeCornflower:
		return []*effect.EffectInstance{with(effect.VanillaJumpBoost(), 80, 0)}
	case SuspiciousStewTypeTulip:
		return []*effect.EffectInstance{with(effect.VanillaWeakness(), 140, 0)}
	case SuspiciousStewTypeAzureBluet:
		return []*effect.EffectInstance{with(effect.VanillaBlindness(), 120, 0)}
	case SuspiciousStewTypeLilyOfTheValley:
		return []*effect.EffectInstance{with(effect.VanillaPoison(), 200, 0)}
	case SuspiciousStewTypeDandelion, SuspiciousStewTypeBlueOrchid:
		return []*effect.EffectInstance{with(effect.VanillaSaturation(), 6, 0)}
	case SuspiciousStewTypeAllium:
		return []*effect.EffectInstance{with(effect.VanillaFireResistance(), 40, 0)}
	case SuspiciousStewTypeOxeyeDaisy:
		return []*effect.EffectInstance{with(effect.VanillaRegeneration(), 120, 0)}
	case SuspiciousStewTypeWitherRose:
		return []*effect.EffectInstance{with(effect.VanillaWither(), 120, 0)}
	}
	return nil
}
