package entity

import "fmt"

// Attribute ID constants, a port of Attribute's constants.
const (
	AttributeMCPrefix                  = "minecraft:"
	AttributeAbsorption                = AttributeMCPrefix + "absorption"
	AttributeSaturation                = AttributeMCPrefix + "player.saturation"
	AttributeExhaustion                = AttributeMCPrefix + "player.exhaustion"
	AttributeKnockbackResistance       = AttributeMCPrefix + "knockback_resistance"
	AttributeHealth                    = AttributeMCPrefix + "health"
	AttributeMovementSpeed             = AttributeMCPrefix + "movement"
	AttributeFollowRange               = AttributeMCPrefix + "follow_range"
	AttributeHunger                    = AttributeMCPrefix + "player.hunger"
	AttributeFood                      = AttributeHunger
	AttributeAttackDamage              = AttributeMCPrefix + "attack_damage"
	AttributeExperienceLevel           = AttributeMCPrefix + "player.level"
	AttributeExperience                = AttributeMCPrefix + "player.experience"
	AttributeUnderwaterMovement        = AttributeMCPrefix + "underwater_movement"
	AttributeLuck                      = AttributeMCPrefix + "luck"
	AttributeFallDamage                = AttributeMCPrefix + "fall_damage"
	AttributeHorseJumpStrength         = AttributeMCPrefix + "horse.jump_strength"
	AttributeZombieSpawnReinforcements = AttributeMCPrefix + "zombie.spawn_reinforcements"
	AttributeLavaMovement              = AttributeMCPrefix + "lava_movement"
)

// Attribute is a port of pocketmine\entity\Attribute. The setters panic on out-of-range values,
// like PHP's InvalidArgumentException.
type Attribute struct {
	id             string
	minValue       float64
	maxValue       float64
	defaultValue   float64
	currentValue   float64
	shouldSend     bool
	desynchronized bool
}

// NewAttribute is a port of Attribute::__construct.
func NewAttribute(id string, minValue, maxValue, defaultValue float64, shouldSend bool) *Attribute {
	if minValue > maxValue || defaultValue > maxValue || defaultValue < minValue {
		panic(fmt.Sprintf("Invalid ranges: min value: %v, max value: %v, %v: %v", minValue, maxValue, defaultValue, defaultValue))
	}
	return &Attribute{
		id:             id,
		minValue:       minValue,
		maxValue:       maxValue,
		defaultValue:   defaultValue,
		currentValue:   defaultValue,
		shouldSend:     shouldSend,
		desynchronized: true,
	}
}

// Clone is PHP's `clone $attribute`.
func (a *Attribute) Clone() *Attribute {
	c := *a
	return &c
}

func (a *Attribute) GetMinValue() float64 { return a.minValue }

func (a *Attribute) SetMinValue(minValue float64) *Attribute {
	if maxValue := a.GetMaxValue(); minValue > maxValue {
		panic(fmt.Sprintf("Minimum %v is greater than the maximum %v", minValue, maxValue))
	}
	if a.minValue != minValue {
		a.desynchronized = true
		a.minValue = minValue
	}
	return a
}

func (a *Attribute) GetMaxValue() float64 { return a.maxValue }

func (a *Attribute) SetMaxValue(maxValue float64) *Attribute {
	if minValue := a.GetMinValue(); maxValue < minValue {
		panic(fmt.Sprintf("Maximum %v is less than the minimum %v", maxValue, minValue))
	}
	if a.maxValue != maxValue {
		a.desynchronized = true
		a.maxValue = maxValue
	}
	return a
}

func (a *Attribute) GetDefaultValue() float64 { return a.defaultValue }

func (a *Attribute) SetDefaultValue(defaultValue float64) *Attribute {
	if defaultValue > a.GetMaxValue() || defaultValue < a.GetMinValue() {
		panic(fmt.Sprintf("Default %v is outside the range %v - %v", defaultValue, a.GetMinValue(), a.GetMaxValue()))
	}
	if a.defaultValue != defaultValue {
		a.desynchronized = true
		a.defaultValue = defaultValue
	}
	return a
}

func (a *Attribute) ResetToDefault() { a.SetValue(a.GetDefaultValue(), true, false) }

func (a *Attribute) GetValue() float64 { return a.currentValue }

// SetValue is a port of Attribute::setValue: out-of-range values panic unless fit is true, in
// which case they're clamped. forceSend marks the attribute for sending even if unchanged.
func (a *Attribute) SetValue(value float64, fit, forceSend bool) *Attribute {
	if value > a.GetMaxValue() || value < a.GetMinValue() {
		if !fit {
			panic(fmt.Sprintf("Value %v is outside the range %v - %v", value, a.GetMinValue(), a.GetMaxValue()))
		}
		value = min(max(value, a.GetMinValue()), a.GetMaxValue())
	}

	if a.currentValue != value {
		a.desynchronized = true
		a.currentValue = value
	} else if forceSend {
		a.desynchronized = true
	}
	return a
}

func (a *Attribute) GetID() string { return a.id }

func (a *Attribute) IsSyncable() bool { return a.shouldSend }

func (a *Attribute) IsDesynchronized() bool { return a.shouldSend && a.desynchronized }

func (a *Attribute) MarkSynchronized(synced bool) { a.desynchronized = !synced }
