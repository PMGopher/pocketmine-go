package player

import (
	stdmath "math"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

	"pocketmine-go/pocketmine/entity"
	"pocketmine-go/pocketmine/entity/animation"
	"pocketmine-go/pocketmine/entity/effect"
	entityevent "pocketmine-go/pocketmine/event/entity"
	playerevent "pocketmine-go/pocketmine/event/player"
	"pocketmine-go/pocketmine/inventory"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
	"pocketmine-go/pocketmine/world"
	"pocketmine-go/pocketmine/world/sound"
)

// This file ports the parts of pocketmine\player\Player that override Entity/Living/Human.

// maxMoveDistanceSquared mirrors actuallyHandleMovement's "moved too fast" limit (15 blocks).
const maxMoveDistanceSquared = 225

// playerStepHeight mirrors Player::$stepHeight.
const playerStepHeight = 0.6

func init() {
	effect.IsPlayer = func(l effect.Living) bool {
		_, ok := l.(*Player)
		return ok
	}
}

// NeverSavedWithChunk marks Player as a NeverSavedWithChunkEntity: players are saved through their
// player data, never with chunks.
func (p *Player) NeverSavedWithChunk() {}

// InitHumanData is a port of Player::initHumanData: the name tag and UUID come from the login.
func (p *Player) InitHumanData(tag *nbt.CompoundTag) {
	p.SetNameTag(p.username)
	p.SetUniqueID(p.playerUUID)
}

// InitEntity is Human::initEntity plus Player's constructor setup (keepMovement, step height).
func (p *Player) InitEntity(tag *nbt.CompoundTag) {
	p.Human.InitEntity(tag)
	p.KeepMovement = true
	p.SetStepHeight(playerStepHeight)
}

// initNetworkHooks mirrors the effect add/remove hooks PHP's Player installs so the client sees
// its own effects (NetworkSession::onEntityEffectAdded/Removed).
func (p *Player) initNetworkHooks() {
	onAdd := effect.EffectAddHook(func(instance *effect.EffectInstance, replacesOldEffect bool) {
		p.SendPacket(entity.EntityEffectAddedPacket(p.GetID(), instance, replacesOldEffect))
	})
	onRemove := effect.EffectRemoveHook(func(instance *effect.EffectInstance) {
		p.SendPacket(entity.EntityEffectRemovedPacket(p.GetID(), instance))
	})
	p.GetEffects().GetEffectAddHooks().Add(&onAdd)
	p.GetEffects().GetEffectRemoveHooks().Add(&onRemove)
}

// SpawnTo is a port of Player::spawnTo: players are only shown to living players, and never while
// in spectator mode. (Player::canSee/hidePlayer isn't ported.)
func (p *Player) SpawnTo(viewer world.EntityViewer) {
	viewerAlive := true
	if other, ok := viewer.(*Player); ok {
		viewerAlive = other.IsAlive()
	}
	if p.IsAlive() && viewerAlive && !p.IsSpectator() {
		p.Human.SpawnTo(viewer)
	}
}

// CanCollideWith is a port of Player::canCollideWith.
func (p *Player) CanCollideWith(other world.Entity) bool { return false }

// CanBeCollidedWith is a port of Player::canBeCollidedWith.
func (p *Player) CanBeCollidedWith() bool {
	return !p.IsSpectator() && p.Human.CanBeCollidedWith()
}

// CanBeMovedByCurrents is a port of Player::canBeMovedByCurrents: currently has no server-side
// movement.
func (p *Player) CanBeMovedByCurrents() bool { return false }

// GetDrops is a port of Player::getDrops: creative players drop nothing.
func (p *Player) GetDrops() []item.Item {
	if p.HasFiniteResources() {
		return p.Human.GetDrops()
	}
	return nil
}

// GetXpDropAmount is a port of Player::getXpDropAmount.
func (p *Player) GetXpDropAmount() int {
	if p.HasFiniteResources() {
		return p.Human.GetXpDropAmount()
	}
	return 0
}

// CalculateFallDamage is a port of Player::calculateFallDamage: flying players take no fall
// damage.
func (p *Player) CalculateFallDamage(fallDistance float64) float64 {
	if p.flying {
		return 0
	}
	return p.Human.CalculateFallDamage(fallDistance)
}

// CheckGroundState is a port of Player::checkGroundState: the player is on the ground if blocks
// collide with a thin box around its feet (widened to cover the whole movement, for running down
// stairs).
func (p *Player) CheckGroundState(wantedX, wantedY, wantedZ, dx, dy, dz float64) {
	if !p.hasBlockCollision {
		p.OnGround = false
	} else {
		bb := p.BoundingBox
		y := p.GetPosition().Y
		bb.MinY = y - 0.2
		bb.MaxY = y + 0.2

		//we're already at the new position at this point; check if there are blocks we might have landed on between
		//the old and new positions (running down stairs necessitates this)
		bb = bb.AddCoord(-dx, -dy, -dz)

		p.OnGround = len(p.GetWorld().GetCollisionBlocks(bb, true)) > 0
		p.IsCollided = p.OnGround
	}
}

// HandleMovement is a port of Player::handleMovement: moves the player to the feet position newPos
// reported by the client. Movements of more than 15 blocks, or into unloaded terrain, are refused
// (reported by returning false, so the caller can resync the client - PHP's revertMovement sends a
// MovePlayerPacket here, which needs the NetworkSession this port doesn't have).
func (p *Player) HandleMovement(newPos math.Vector3) bool {
	oldPos := p.GetPosition()
	distanceSquared := newPos.DistanceSquared(oldPos)

	if distanceSquared > maxMoveDistanceSquared || !p.GetWorld().IsInLoadedTerrain(newPos) {
		return false
	}

	if distanceSquared != 0 {
		p.Move(newPos.X-oldPos.X, newPos.Y-oldPos.Y, newPos.Z-oldPos.Z)
	}
	return true
}

// SetMotion is a port of Player::setMotion: the player's own client is told about its motion too.
func (p *Player) SetMotion(motion math.Vector3) bool {
	if p.Human.SetMotion(motion) {
		p.BroadcastMotion()
		p.SendPacket(&packet.SetActorMotion{EntityRuntimeID: uint64(p.GetID()), Velocity: vec32(motion), Tick: 0})

		return true
	}
	return false
}

// UpdateMovement is a no-op for players (Player::updateMovement): movement is broadcast from
// processMostRecentMovements instead.
func (p *Player) UpdateMovement(teleport bool) {}

// TryChangeMovement is a no-op for players (Player::tryChangeMovement): clients move themselves.
func (p *Player) TryChangeMovement() {}

// OnUpdate is a port of Player::onUpdate.
func (p *Player) OnUpdate(currentTick int64) bool {
	tickDiff := int(currentTick - p.LastUpdate)

	if tickDiff <= 0 {
		return true
	}

	p.LastUpdate = currentTick

	if p.IsJustCreated() {
		p.OnFirstUpdate(currentTick)
	}

	if !p.IsAlive() && p.spawned {
		p.OnDeathUpdate(tickDiff)
		return true
	}

	if p.spawned {
		p.processMostRecentMovements()
		p.SetMotionDirect(math.Vector3Zero()) //TODO: HACK! (Fixes player knockback being messed up)
		if p.OnGround {
			p.inAirTicks = 0
		} else {
			p.inAirTicks += tickDiff
		}

		p.EntityBaseTick(tickDiff)

		if p.IsCreative() && p.GetFireTicks() > 1 {
			p.SetFireTicks(1)
		}

		if !p.IsSpectator() && p.IsAlive() {
			p.checkNearEntities()
		}

		if p.blockBreakHandler != nil && !p.blockBreakHandler.Update(p.GetInventory().GetItemInHand()) {
			p.blockBreakHandler = nil
		}

		p.syncAttributes()
	}

	return true
}

// syncAttributes is NetworkSession::syncAttributes for the player itself: changed attributes
// (health, hunger, experience, movement speed, ...) are sent to the client every tick.
func (p *Player) syncAttributes() {
	needSend := p.GetAttributeMap().NeedSend()
	if pk := entity.SyncAttributesPacket(p.GetID(), needSend); pk != nil {
		p.SendPacket(pk)
		for _, attr := range needSend {
			attr.MarkSynchronized(true)
		}
	}
}

// processMostRecentMovements is a port of Player::processMostRecentMovements, minus the rate limit
// and the cancellable PlayerMoveEvent (not ported).
func (p *Player) processMostRecentMovements() {
	from := p.lastBroadcastLocation
	to := p.GetLocation()

	delta := to.DistanceSquared(from.Vector3)
	deltaAngle := stdmath.Abs(from.Yaw-to.Yaw) + stdmath.Abs(from.Pitch-to.Pitch)

	if delta > 0.0001 || deltaAngle > 1.0 {
		p.lastBroadcastLocation = to
		p.BroadcastMovement(false)

		horizontalDistanceTravelled := stdmath.Sqrt((from.X-to.X)*(from.X-to.X) + (from.Z-to.Z)*(from.Z-to.Z))
		if horizontalDistanceTravelled > 0 {
			//TODO: check for swimming
			if p.IsSprinting() {
				p.GetHungerManager().Exhaust(0.01*horizontalDistanceTravelled, playerevent.ExhaustCauseSprinting)
			} else {
				p.GetHungerManager().Exhaust(0.0, playerevent.ExhaustCauseWalking)
			}
		}
	}
}

func vec32(v math.Vector3) mgl32.Vec3 {
	return mgl32.Vec3{float32(v.X), float32(v.Y), float32(v.Z)}
}

// collidable is the optional surface checkNearEntities calls on colliding entities
// (Entity::onCollideWithPlayer).
type collidable interface {
	OnCollideWithPlayer(player entity.Player)
}

// checkNearEntities is a port of Player::checkNearEntities: item entities, arrows and orbs next to
// the player are picked up.
func (p *Player) checkNearEntities() {
	for _, e := range p.GetWorld().GetNearbyEntitiesExcept(p.BoundingBox.ExpandedCopy(1, 0.5, 1), p) {
		if !e.IsClosed() {
			if scheduler, ok := e.(interface{ ScheduleUpdate() }); ok {
				scheduler.ScheduleUpdate()
			}
		}

		if !e.IsAlive() || e.IsFlaggedForDespawn() {
			continue
		}

		if c, ok := e.(collidable); ok {
			c.OnCollideWithPlayer(p)
		}
	}
}

// CanEat is a port of Player::canEat.
func (p *Player) CanEat() bool { return p.IsCreative() || p.Human.CanEat() }

// CanBreathe is a port of Player::canBreathe.
func (p *Player) CanBreathe() bool { return p.IsCreative() || p.Human.CanBreathe() }

// ApplyPostDamageEffects is a port of Player::applyPostDamageEffects.
func (p *Player) ApplyPostDamageEffects(source entityevent.DamageSource) {
	p.Human.ApplyPostDamageEffects(source)

	p.GetHungerManager().Exhaust(0.1, playerevent.ExhaustCauseDamage)
}

// Attack is a port of Player::attack: creative players take no damage (except /kill), and players
// allowed to fly take no fall damage.
func (p *Player) Attack(source entityevent.DamageSource) {
	if !p.IsAlive() {
		return
	}

	if p.IsCreative() && source.GetCause() != entityevent.CauseSuicide {
		source.Cancel()
	} else if p.allowFlight && source.GetCause() == entityevent.CauseFall {
		source.Cancel()
	}

	p.Human.Attack(source)
}

// SyncNetworkData is a port of Player::syncNetworkData, minus the death position (players don't
// track one yet) and sleeping bed position (sleeping isn't ported).
func (p *Player) SyncNetworkData(properties *entity.MetadataCollection) {
	p.Human.SyncNetworkData(properties)

	properties.SetGenericFlag(entity.FlagAction, false) // item use (startAction) isn't ported
	properties.SetGenericFlag(entity.FlagHasCollision, p.HasBlockCollision())

	properties.SetPlayerFlag(playerFlagSleep, false)
}

// playerFlagSleep mirrors PlayerMetadataFlags::SLEEP.
const playerFlagSleep = 1

// selfAndViewers is PHP's `$targets = $this->getViewers(); $targets[] = $this;`.
func (p *Player) selfAndViewers() []world.EntityViewer {
	return append(p.GetViewers(), p)
}

// SendData is a port of Player::sendData: the player's own client gets its metadata too.
func (p *Player) SendData(targets []world.EntityViewer, data protocol.EntityMetadata) {
	if targets == nil {
		targets = p.selfAndViewers()
	}
	p.Human.SendData(targets, data)
}

// BroadcastAnimation is a port of Player::broadcastAnimation.
func (p *Player) BroadcastAnimation(anim animation.Animation, targets []world.EntityViewer) {
	if p.spawned && targets == nil {
		targets = p.selfAndViewers()
	}
	p.Human.BroadcastAnimation(anim, targets)
}

// BroadcastSound is a port of Player::broadcastSound (to viewers and the player itself).
func (p *Player) BroadcastSound(s sound.Sound) { p.BroadcastSoundTo(s, nil) }

// BroadcastSoundTo is Player::broadcastSound with explicit targets.
func (p *Player) BroadcastSoundTo(s sound.Sound, targets []world.EntityViewer) {
	if p.spawned && targets == nil {
		targets = p.selfAndViewers()
	}
	p.Human.BroadcastSoundTo(s, targets)
}

// Kill is a port of Player::kill: players that haven't spawned yet can't die.
func (p *Player) Kill() {
	if !p.spawned {
		return
	}
	p.Human.Kill()
}

// OnDeath is a port of Player::onDeath, minus PlayerDeathEvent (keep-inventory/keep-xp/death
// message - event/player isn't ported, so inventory and XP are always dropped) and the client
// death screen (NetworkSession::onServerDeath).
func (p *Player) OnDeath() {
	w := p.GetWorld()
	for _, it := range p.GetDrops() {
		w.DropItem(p.GetPosition(), it, nil, 10)
	}

	clearInventory := func(inv inventory.Inventory) {
		kept := map[int]item.Item{}
		for slot, it := range inv.GetContents(false) {
			if it.KeepOnDeath() {
				kept[slot] = it
			}
		}
		inv.SetContents(kept)
	}
	p.GetInventory().SetHeldItemIndex(0)
	clearInventory(p.GetInventory())
	clearInventory(p.GetArmorInventory())
	clearInventory(p.GetOffHandInventory())

	w.DropExperience(p.GetPosition(), p.GetXpDropAmount())
	level, progress := 0, 0.0
	p.GetXpManager().SetXpAndProgress(&level, &progress)

	p.StartDeathAnimation()
}

// OnDeathUpdate is a port of Player::onDeathUpdate: players are never flagged for despawn.
func (p *Player) OnDeathUpdate(tickDiff int) bool {
	p.Human.OnDeathUpdate(tickDiff)
	return false //never flag players for despawn
}
