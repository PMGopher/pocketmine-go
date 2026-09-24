package player

import (
	"pocketmine-go/pocketmine/entity/animation"
	"pocketmine-go/pocketmine/world/sound"
)

// The Player::toggleX methods the client's input flags go through. Player toggle events aren't
// ported, so only the event's built-in cancellations remain (toggleSneak's "nothing changed",
// toggleFlight's "flight not allowed").

// ToggleSprint is a port of Player::toggleSprint.
func (p *Player) ToggleSprint(sprint bool) bool {
	if sprint == p.IsSprinting() {
		return true
	}
	p.SetSprinting(sprint)
	return true
}

// ToggleSneak is a port of Player::toggleSneak.
func (p *Player) ToggleSneak(sneak, sneakPressed bool) bool {
	if sneak == p.IsSneaking() && sneakPressed == p.sneakPressed {
		return true
	}
	p.SetSneakPressed(sneakPressed)

	if sneak == p.IsSneaking() { // PHP cancels the PlayerToggleSneakEvent here
		return false
	}
	p.SetSneaking(sneak)
	return true
}

// ToggleGlide is a port of Player::toggleGlide.
func (p *Player) ToggleGlide(glide bool) bool {
	if glide == p.IsGliding() {
		return true
	}
	p.SetGliding(glide)
	return true
}

// ToggleSwim is a port of Player::toggleSwim.
func (p *Player) ToggleSwim(swim bool) bool {
	if swim == p.IsSwimming() {
		return true
	}
	p.SetSwimming(swim)
	return true
}

// ToggleFlight is a port of Player::toggleFlight.
func (p *Player) ToggleFlight(fly bool) bool {
	if fly == p.flying {
		return true
	}
	if !p.allowFlight { // PHP cancels the PlayerToggleFlightEvent here
		return false
	}
	p.SetFlying(fly)
	return true
}

// MissSwing is a port of Player::missSwing (minus the cancellable PlayerMissSwingEvent).
func (p *Player) MissSwing() {
	p.BroadcastSound(sound.EntityAttackNoDamageSound{})
	p.BroadcastAnimation(animation.ArmSwingAnimation{Entity: p}, p.GetViewers())
}
