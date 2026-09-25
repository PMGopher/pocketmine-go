package player

import (
	"pocketmine-go/pocketmine/event"
	entityevent "pocketmine-go/pocketmine/event/entity"
	"pocketmine-go/pocketmine/lang"
)

// PlayerDeathEvent is a port of pocketmine\event\player\PlayerDeathEvent. It extends
// EntityDeathEvent, so EntityDeathEvent handlers receive it too.
type PlayerDeathEvent struct {
	entityevent.EntityDeathEvent

	player             Player
	deathMessage       any
	deathScreenMessage any
	keepInventory      bool
	keepXp             bool
}

// NewPlayerDeathEvent creates the event. deathMessage may be nil, in which case it's derived from
// the player's last damage cause (DeriveMessage); lastDamageCause is Entity::getLastDamageCause.
func NewPlayerDeathEvent(entity Player, drops []Item, xp int, deathMessage any, lastDamageCause entityevent.DamageSource) *PlayerDeathEvent {
	e := &PlayerDeathEvent{EntityDeathEvent: *entityevent.NewEntityDeathEvent(entity, drops, xp), player: entity}
	if deathMessage == nil {
		deathMessage = DeriveMessage(entity.GetDisplayName(), lastDamageCause)
	}
	e.deathMessage = deathMessage
	e.deathScreenMessage = deathMessage
	return e
}

func (e *PlayerDeathEvent) GetPlayer() Player { return e.player }

func (e *PlayerDeathEvent) GetDeathMessage() any { return e.deathMessage }

func (e *PlayerDeathEvent) SetDeathMessage(deathMessage any) { e.deathMessage = deathMessage }

func (e *PlayerDeathEvent) GetDeathScreenMessage() any { return e.deathScreenMessage }

func (e *PlayerDeathEvent) SetDeathScreenMessage(deathScreenMessage any) {
	e.deathScreenMessage = deathScreenMessage
}

func (e *PlayerDeathEvent) GetKeepInventory() bool { return e.keepInventory }

func (e *PlayerDeathEvent) SetKeepInventory(keepInventory bool) { e.keepInventory = keepInventory }

func (e *PlayerDeathEvent) GetKeepXp() bool { return e.keepXp }

func (e *PlayerDeathEvent) SetKeepXp(keepXp bool) { e.keepXp = keepXp }

// DeathMessageClassifier answers the instanceof checks PlayerDeathEvent::deriveMessage makes on
// damagers, which live in packages above this one. The player package installs the real one in
// its init(); the zero value treats every entity as a plain non-living entity.
var DeathMessageClassifier struct {
	IsPlayer         func(e Entity) bool
	IsLiving         func(e Entity) bool
	IsTrident        func(e Entity) bool
	IsFireworkRocket func(e Entity) bool
	// FallingBlockKind reports whether e is a FallingBlock, and whether its block is an anvil.
	FallingBlockKind func(e Entity) (isFallingBlock, isAnvil bool)
	// IsCactus reports whether b is a cactus (BlockTypeIds::CACTUS).
	IsCactus func(b entityevent.Block) bool
}

// byEntity is PHP's `instanceof EntityDamageByEntityEvent`, which also matches its subclass
// EntityDamageByChildEntityEvent.
type byEntity interface {
	GetDamager() entityevent.Entity
}

func classify(f func(e Entity) bool, e Entity) bool { return f != nil && e != nil && f(e) }

// displayName is Living::getDisplayName.
func displayName(e Entity) string {
	if n, ok := e.(interface{ GetDisplayName() string }); ok {
		return n.GetDisplayName()
	}
	return ""
}

// DeriveMessage is a port of PlayerDeathEvent::deriveMessage: the vanilla death message for a
// player named name killed by deathCause (nil for an unknown cause).
func DeriveMessage(name string, deathCause entityevent.DamageSource) *lang.Translatable {
	f := lang.KnownTranslationFactory
	c := &DeathMessageClassifier
	cause := entityevent.CauseCustom
	if deathCause != nil {
		cause = deathCause.GetCause()
	}
	switch cause {
	case entityevent.CauseEntityAttack:
		if ev, ok := deathCause.(byEntity); ok {
			e := ev.GetDamager()
			if classify(c.IsPlayer, e) {
				return f.DeathAttackPlayer(name, displayName(e))
			} else if classify(c.IsLiving, e) {
				return f.DeathAttackMob(name, displayName(e))
			}
		}
	case entityevent.CauseProjectile:
		if ev, ok := deathCause.(*entityevent.EntityDamageByChildEntityEvent); ok {
			e := ev.GetDamager()
			if classify(c.IsLiving, e) {
				if classify(c.IsTrident, ev.GetChild()) {
					return f.DeathAttackTrident(name, displayName(e))
				}
				return f.DeathAttackArrow(name, displayName(e))
			}
		}
	case entityevent.CauseSuicide:
		return f.DeathAttackGeneric(name)
	case entityevent.CauseVoid:
		return f.DeathAttackOutOfWorld(name)
	case entityevent.CauseFall:
		if deathCause != nil && deathCause.GetFinalDamage() > 2 {
			return f.DeathFellAccidentGeneric(name)
		}
		return f.DeathAttackFall(name)
	case entityevent.CauseSuffocation:
		return f.DeathAttackInWall(name)
	case entityevent.CauseLava:
		return f.DeathAttackLava(name)
	case entityevent.CauseFire:
		return f.DeathAttackOnFire(name)
	case entityevent.CauseFireTick:
		return f.DeathAttackInFire(name)
	case entityevent.CauseDrowning:
		return f.DeathAttackDrown(name)
	case entityevent.CauseContact:
		if ev, ok := deathCause.(*entityevent.EntityDamageByBlockEvent); ok {
			if c.IsCactus != nil && c.IsCactus(ev.GetDamager()) {
				return f.DeathAttackCactus(name)
			}
		}
	case entityevent.CauseBlockExplosion, entityevent.CauseEntityExplosion:
		if ev, ok := deathCause.(byEntity); ok {
			e := ev.GetDamager()
			if classify(c.IsFireworkRocket, e) {
				return f.DeathAttackFireworks(name)
			} else if classify(c.IsLiving, e) {
				return f.DeathAttackExplosionPlayer(name, displayName(e))
			}
		}
		return f.DeathAttackExplosion(name)
	case entityevent.CauseMagic:
		return f.DeathAttackMagic(name)
	case entityevent.CauseFallingBlock:
		if ev, ok := deathCause.(byEntity); ok {
			if e := ev.GetDamager(); e != nil && c.FallingBlockKind != nil {
				if isFallingBlock, isAnvil := c.FallingBlockKind(e); isFallingBlock {
					if isAnvil {
						return f.DeathAttackAnvil(name)
					}
					return f.DeathAttackFallingBlock(name)
				}
			}
		}
	}
	return f.DeathAttackGeneric(name)
}

func init() {
	event.DeclareParent[PlayerDeathEvent, entityevent.EntityDeathEvent]()
}
