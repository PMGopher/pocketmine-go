package effect

import (
	"fmt"

	"pocketmine-go/pocketmine/binaryutils"
	"pocketmine-go/pocketmine/color"
)

// EffectInstance is a port of pocketmine\entity\effect\EffectInstance.
type EffectInstance struct {
	effectType Effect
	duration   int
	amplifier  int
	visible    bool
	ambient    bool
	color      color.Color
	infinite   bool
}

// NewEffectInstance is a port of EffectInstance::__construct with PHP's defaults (duration = the
// type's default, amplifier 0, visible, not ambient, the type's color, not infinite). Use
// NewEffectInstanceFull for the rest.
func NewEffectInstance(effectType Effect) *EffectInstance {
	return NewEffectInstanceFull(effectType, nil, 0, true, false, nil, false)
}

// NewEffectInstanceWith is NewEffectInstance with an explicit duration and amplifier - the most
// common PHP call shape (`new EffectInstance($type, $duration, $amplifier)`).
func NewEffectInstanceWith(effectType Effect, duration, amplifier int) *EffectInstance {
	return NewEffectInstanceFull(effectType, &duration, amplifier, true, false, nil, false)
}

// NewEffectInstanceFull is the full port of EffectInstance::__construct. A nil duration means the
// type's default duration; a nil overrideColor means the type's color. Panics on the same invalid
// arguments as the PHP InvalidArgumentException.
func NewEffectInstanceFull(effectType Effect, duration *int, amplifier int, visible, ambient bool, overrideColor *color.Color, infinite bool) *EffectInstance {
	e := &EffectInstance{effectType: effectType, visible: visible, ambient: ambient, infinite: infinite}
	if duration != nil {
		e.SetDuration(*duration)
	} else {
		e.SetDuration(effectType.GetDefaultDuration())
	}
	e.SetAmplifier(amplifier)
	if overrideColor != nil {
		e.color = *overrideColor
	} else {
		e.color = effectType.GetColor()
	}
	return e
}

// Clone is PHP's `clone $effectInstance` - a shallow copy sharing the same Effect type.
func (e *EffectInstance) Clone() *EffectInstance {
	c := *e
	return &c
}

func (e *EffectInstance) GetType() Effect { return e.effectType }

// GetDuration returns the number of ticks remaining until the effect expires.
func (e *EffectInstance) GetDuration() int { return e.duration }

// SetDuration sets the number of ticks remaining until the effect expires.
func (e *EffectInstance) SetDuration(duration int) *EffectInstance {
	if duration < 0 || duration > binaryutils.Int32Max {
		panic(fmt.Sprintf("Effect duration must be in range 0 - %d, got %d", binaryutils.Int32Max, duration))
	}
	e.duration = duration
	return e
}

// DecreaseDuration is a port of EffectInstance::decreaseDuration: infinite effects wrap around
// instead of expiring.
func (e *EffectInstance) DecreaseDuration(ticks int) *EffectInstance {
	newDuration := e.duration - ticks
	if newDuration <= 0 {
		if e.infinite {
			newDuration += binaryutils.Int32Max
		} else {
			newDuration = 0
		}
	}
	e.duration = newDuration
	return e
}

// HasExpired returns whether the duration has run out.
func (e *EffectInstance) HasExpired() bool { return e.duration <= 0 }

func (e *EffectInstance) GetAmplifier() int { return e.amplifier }

// GetEffectLevel returns the level of this effect, which is always one higher than the amplifier.
func (e *EffectInstance) GetEffectLevel() int { return e.amplifier + 1 }

// SetAmplifier is a port of EffectInstance::setAmplifier (range 0-255).
func (e *EffectInstance) SetAmplifier(amplifier int) *EffectInstance {
	if amplifier < 0 || amplifier > 255 {
		panic(fmt.Sprintf("Amplifier must be in range 0 - 255, got %d", amplifier))
	}
	e.amplifier = amplifier
	return e
}

// IsVisible returns whether this effect will produce some visible effect, such as bubbles or
// particles.
func (e *EffectInstance) IsVisible() bool { return e.visible }

func (e *EffectInstance) SetVisible(visible bool) *EffectInstance {
	e.visible = visible
	return e
}

// IsAmbient returns whether the effect originated from the ambient environment. Ambient effects can
// originate from things such as a Beacon's area of effect radius. If this flag is set, the amount
// of visible particles will be reduced by a factor of 5.
func (e *EffectInstance) IsAmbient() bool { return e.ambient }

func (e *EffectInstance) SetAmbient(ambient bool) *EffectInstance {
	e.ambient = ambient
	return e
}

// GetColor returns the particle colour of this effect instance. This can be overridden on a
// per-EffectInstance basis, so it is not reflective of the default colour of the effect.
func (e *EffectInstance) GetColor() color.Color { return e.color }

func (e *EffectInstance) SetColor(c color.Color) *EffectInstance {
	e.color = c
	return e
}

// ResetColor resets the colour of this EffectInstance to the default specified by its type.
func (e *EffectInstance) ResetColor() *EffectInstance {
	e.color = e.effectType.GetColor()
	return e
}

func (e *EffectInstance) IsInfinite() bool { return e.infinite }
