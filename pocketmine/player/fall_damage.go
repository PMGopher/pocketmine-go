package player

import (
	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/world/sound"
)

// entityNetworkTypeIDPlayer mirrors Entity::getNetworkTypeId() for a real Player
// ("minecraft:player") - the only concrete entity type this port has, so it's hardcoded here
// rather than built out as a general Entity-type-registry lookup (a whole separate subsystem this
// port doesn't have anywhere else either).
const entityNetworkTypeIDPlayer = "minecraft:player"

// TrackFallState is a port of the per-tick fall-tracking real PHP drives from inside its own
// physics loop (Entity::move calling Entity::updateFallState with a freshly-computed Y delta every
// tick - named differently here to avoid colliding with the promoted entity.Entity.UpdateFallState
// this itself calls). This port has no server-side physics - Player's position comes from the
// client's own PlayerAuthInput reports instead (see cmd/pocketmine-go's own doc comments on why) -
// so the caller (the PlayerAuthInput handler) supplies the new Y position and the client-reported
// onGround state directly, once per input packet, instead of this being driven internally every
// physics tick.
//
// The long/short-fall/land sound selection on a genuine landing is real (Living::onHitGround's own
// broadcastSound calls) - EntityLandSound's landed-on block is approximated as the block directly
// below the player's feet (real PHP walks up one block if that one has no collision box; skipped
// here since GetBlockAt always returns a real block either way, just possibly non-solid air, which
// is an honest reflection of what's actually below the player).
func (p *Player) TrackFallState(newY float64, onGround bool) {
	deltaY := newY - p.lastY
	p.lastY = newY
	wasOnGround := p.IsOnGround()
	fallDistanceBeforeLanding := p.GetFallDistance()

	p.SetOnGround(onGround)
	p.Human.Living.Entity.UpdateFallState(deltaY, onGround)

	if !onGround || wasOnGround || fallDistanceBeforeLanding <= 0 {
		return
	}

	pos := p.GetPosition()
	damage := p.CalculateFallDamage(fallDistanceBeforeLanding)
	switch {
	case damage > 4:
		p.world.AddSound(pos, sound.EntityLongFallSound{EntityNetworkTypeID: entityNetworkTypeIDPlayer, EntityUniqueID: int64(p.GetID())})
	case damage > 0:
		p.world.AddSound(pos, sound.EntityShortFallSound{EntityNetworkTypeID: entityNetworkTypeIDPlayer})
	default:
		landedOn := p.world.GetBlockAt(pos.FloorX(), pos.FloorY()-1, pos.FloorZ())
		if landedOn.GetTypeId() != block.AIR {
			p.world.AddSound(pos, sound.EntityLandSound{
				BlockStateID:        landedOn.GetStateId(),
				EntityNetworkTypeID: entityNetworkTypeIDPlayer,
				EntityUniqueID:      int64(p.GetID()),
			})
		}
	}
}
