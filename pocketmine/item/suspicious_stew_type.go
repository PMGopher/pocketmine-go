package item

// SuspiciousStewType is a port of pocketmine\item\SuspiciousStewType. GetEffects isn't ported -
// needs EffectInstance (entity/effect package, not ported), same gap documented throughout this
// port wherever an EffectInstance would be constructed.
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
