package player

import (
	"pocketmine-go/pocketmine/event"
	playerevent "pocketmine-go/pocketmine/event/player"
)

// ToggleSprint is a port of Player::toggleSprint.
func (p *Player) ToggleSprint(sprint bool) bool {
	if sprint == p.IsSprinting() {
		return true
	}
	ev := playerevent.NewPlayerToggleSprintEvent(p, sprint)
	event.Call(ev)
	if ev.IsCancelled() {
		return false
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

	ev := playerevent.NewPlayerToggleSneakEvent(p, sneak, sneakPressed)
	if sneak == p.IsSneaking() {
		ev.Cancel()
	}
	event.Call(ev)

	if ev.IsCancelled() {
		return false
	}
	p.SetSneaking(sneak)
	return true
}

// ToggleFlight is a port of Player::toggleFlight.
func (p *Player) ToggleFlight(fly bool) bool {
	if fly == p.flying {
		return true
	}
	ev := playerevent.NewPlayerToggleFlightEvent(p, fly)
	if !p.allowFlight {
		ev.Cancel()
	}
	event.Call(ev)
	if ev.IsCancelled() {
		return false
	}
	p.SetFlying(fly)
	return true
}

// ToggleGlide is a port of Player::toggleGlide.
func (p *Player) ToggleGlide(glide bool) bool {
	if glide == p.IsGliding() {
		return true
	}
	ev := playerevent.NewPlayerToggleGlideEvent(p, glide)
	event.Call(ev)
	if ev.IsCancelled() {
		return false
	}
	p.SetGliding(glide)
	return true
}

// ToggleSwim is a port of Player::toggleSwim.
func (p *Player) ToggleSwim(swim bool) bool {
	if swim == p.IsSwimming() {
		return true
	}
	ev := playerevent.NewPlayerToggleSwimEvent(p, swim)
	event.Call(ev)
	if ev.IsCancelled() {
		return false
	}
	p.SetSwimming(swim)
	return true
}
