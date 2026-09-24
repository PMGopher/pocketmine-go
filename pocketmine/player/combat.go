package player

import (
	"pocketmine-go/pocketmine/entity/animation"
	"pocketmine-go/pocketmine/entity/effect"
	"pocketmine-go/pocketmine/entity/object"
	"pocketmine-go/pocketmine/entity/projectile"
	entityevent "pocketmine-go/pocketmine/event/entity"
	playerevent "pocketmine-go/pocketmine/event/player"
	"pocketmine-go/pocketmine/item/enchantment"
	"pocketmine-go/pocketmine/world"
	"pocketmine-go/pocketmine/world/sound"
)

// maxReachDistanceEntityInteraction mirrors Player::MAX_REACH_DISTANCE_ENTITY_INTERACTION.
const maxReachDistanceEntityInteraction = 8

// livingTarget is the surface AttackEntity needs to recognise a Living target (PHP's instanceof).
type livingTarget interface {
	world.Entity
	IsLiving() bool
	BroadcastAnimation(anim animation.Animation, targets []world.EntityViewer)
}

// AttackEntity is a port of Player::attackEntity: the player attacks target with the held item.
//
// Not ported: the PVP server-property gate (no ServerProperties), and the held item's own
// onAttackEntity (durability loss/returned items - item-use actions aren't ported).
func (p *Player) AttackEntity(target world.Entity) bool {
	if !target.IsAlive() {
		return false
	}
	switch target.(type) {
	case *object.ItemEntity, *projectile.Arrow:
		return false
	}

	heldItem := p.GetInventory().GetItemInHand()

	ev := entityevent.NewEntityDamageByEntityEvent(p, target, entityevent.CauseEntityAttack, float64(heldItem.GetAttackPoints()), nil)
	if !p.CanInteract(target.GetPosition(), maxReachDistanceEntityInteraction) {
		ev.Cancel()
	} else if p.IsSpectator() {
		ev.Cancel()
	}

	meleeEnchantmentDamage := 0.0
	var meleeEnchantments []*enchantment.EnchantmentInstance
	for _, instance := range heldItem.GetEnchantments() {
		if melee, ok := instance.GetType().(enchantment.MeleeWeaponEnchantment); ok {
			if victim, ok := target.(enchantment.Entity); ok && melee.IsApplicableTo(victim) {
				meleeEnchantmentDamage += melee.GetDamageBonus(instance.GetLevel())
				meleeEnchantments = append(meleeEnchantments, instance)
			}
		}
	}
	ev.SetModifier(meleeEnchantmentDamage, entityevent.ModifierWeaponEnchantments)

	if !p.IsSprinting() && !p.IsFlying() && p.FallDistance > 0 && !p.GetEffects().Has(effect.VanillaBlindness()) && !p.IsUnderwater() {
		ev.SetModifier(ev.GetFinalDamage()/2, entityevent.ModifierCritical)
	}

	target.Attack(ev)
	p.BroadcastAnimation(animation.ArmSwingAnimation{Entity: p}, p.GetViewers())

	bb := target.GetBoundingBox()
	soundPos := target.GetPosition().Add(0, bb.GetYLength()/2, 0)
	if ev.IsCancelled() {
		p.GetWorld().AddSound(soundPos, sound.EntityAttackNoDamageSound{})
		return false
	}
	p.GetWorld().AddSound(soundPos, sound.EntityAttackSound{})

	if living, ok := target.(livingTarget); ok {
		if ev.GetModifier(entityevent.ModifierCritical) > 0 {
			living.BroadcastAnimation(animation.CriticalHitAnimation{Entity: living, ParticleCount: animation.DefaultCriticalHitParticleCount}, nil)
		}
		if ev.GetModifier(entityevent.ModifierWeaponEnchantments) > 0 {
			living.BroadcastAnimation(animation.MagicHitAnimation{Entity: living, ParticleCount: animation.DefaultMagicHitParticleCount}, nil)
		}
	}

	if victim, ok := target.(enchantment.Entity); ok {
		for _, instance := range meleeEnchantments {
			instance.GetType().(enchantment.MeleeWeaponEnchantment).OnPostAttack(p, victim, instance.GetLevel())
		}
	}

	if p.IsAlive() {
		//reactive damage like thorns might cause us to be killed by attacking another mob, which
		//would mean we'd already have dropped the inventory by the time we reached here
		p.GetHungerManager().Exhaust(0.1, playerevent.ExhaustCauseAttack)
	}

	return true
}
