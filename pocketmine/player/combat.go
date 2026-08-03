package player

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/entity"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/world/sound"
)

// actorEventArmSwing/actorEventHurt mirror pocketmine\network\mcpe\protocol\types\ActorEvent::
// ARM_SWING (4)/HURT_ANIMATION (2) - verified against github.com/pmmp/BedrockProtocol, since
// gophertunnel's own same-numbered constants (ActorEventStartAttacking/ActorEventHurt) use
// different names for the same wire values.
const (
	actorEventArmSwing = packet.ActorEventStartAttacking
	actorEventHurt     = packet.ActorEventHurt
)

// maxReachDistanceEntityInteraction mirrors Player::MAX_REACH_DISTANCE_ENTITY_INTERACTION.
const maxReachDistanceEntityInteraction = 8

// attackable is the local surface AttackEntity needs from its target - declared locally (matching
// this port's established forward-compatible-local-interface convention) rather than requiring a
// concrete *Player, so any future concrete mob type satisfies it automatically without this
// package needing to import one. *Player itself already satisfies this (EntityLike/IsAlive/Attack
// promoted from entity.Entity/entity.Living, GetID defined directly - see player.go).
type attackable interface {
	entity.EntityLike
	IsAlive() bool
	GetID() int
	GetBoundingBox() math.AxisAlignedBB
	Attack(source entity.DamageSource)
}

// attackPointsGetter is the local surface AttackEntity needs from heldItem to compute base damage
// - block.Item itself doesn't expose GetAttackPoints (see item.Item's own richer interface, not
// used here to avoid pulling in the whole item package), so this defaults to 1 (matching real
// PHP's own base Item::getAttackPoints() default) for any block.Item that doesn't implement it.
type attackPointsGetter interface{ GetAttackPoints() int }

func attackPoints(heldItem block.Item) float64 {
	if g, ok := heldItem.(attackPointsGetter); ok {
		return float64(g.GetAttackPoints())
	}
	return 1
}

// AttackEntity is a port of a slice of Player::attackEntity - the ArmSwingAnimation/HurtAnimation
// broadcasts are real (see World.BroadcastPacketToViewers), sent as ActorEventPacket exactly like
// real PHP's own ArmSwingAnimation/HurtAnimation classes (not AnimatePacket - that's only used by
// CriticalHitAnimation/MagicHitAnimation, neither of which fire here, see below).
//
// Not ported: the ItemEntity/Arrow non-attackable-type exclusion (neither type exists in this
// port yet), melee weapon enchantment damage bonuses (no enchantment system exists) and the
// resulting MagicHitAnimation, the critical-hit sprint/fall-distance damage modifier (this port
// tracks fall distance but not sprinting state) and the resulting CriticalHitAnimation, and the
// PVP server-property gate (no ServerProperties/config type exists). heldItem is the item the
// player is holding right now - see SurvivalBlockBreakHandler's own doc comment on why this port's
// Player has no "selected hotbar slot" concept yet to read it from directly.
func (p *Player) AttackEntity(target attackable, heldItem block.Item) bool {
	if !target.IsAlive() {
		return false
	}

	ev := entity.NewEntityDamageByEntityEvent(p, target, entity.EntityDamageCauseEntityAttack, attackPoints(heldItem), nil)
	if !p.CanInteract(target.GetPosition(), maxReachDistanceEntityInteraction) {
		ev.Cancel()
	} else if p.IsSpectator() {
		ev.Cancel()
	}

	target.Attack(ev)
	p.world.BroadcastPacketToViewers(p.GetPosition(), &packet.ActorEvent{EntityRuntimeID: uint64(p.GetID()), EventType: actorEventArmSwing})

	bb := target.GetBoundingBox()
	soundPos := target.GetPosition().Add(0, (bb.MaxY-bb.MinY)/2, 0)
	if ev.IsCancelled() {
		p.world.AddSound(soundPos, sound.EntityAttackNoDamageSound{})
		return false
	}
	p.world.AddSound(soundPos, sound.EntityAttackSound{})
	p.world.BroadcastPacketToViewers(target.GetPosition(), &packet.ActorEvent{EntityRuntimeID: uint64(target.GetID()), EventType: actorEventHurt})
	return true
}
