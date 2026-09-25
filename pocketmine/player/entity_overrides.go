package player

import (
	"fmt"
	stdmath "math"
	"time"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

	"pocketmine-go/pocketmine/entity"
	"pocketmine-go/pocketmine/entity/animation"
	"pocketmine-go/pocketmine/entity/effect"
	"pocketmine-go/pocketmine/event"
	entityevent "pocketmine-go/pocketmine/event/entity"
	playerevent "pocketmine-go/pocketmine/event/player"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/timings"
	"pocketmine-go/pocketmine/world"
	"pocketmine-go/pocketmine/world/sound"
)

// This file ports the parts of pocketmine\player\Player that override Entity/Living/Human, and
// the movement handling.

func init() {
	effect.IsPlayer = func(l effect.Living) bool {
		_, ok := l.(*Player)
		return ok
	}
}

// SpawnTo is a port of Player::spawnTo: players are only shown to living players that can see
// them, and never while in spectator mode.
func (p *Player) SpawnTo(viewer world.EntityViewer) {
	other, isPlayer := viewer.(*Player)
	if p.IsAlive() && (!isPlayer || (other.IsAlive() && other.CanSee(p))) && !p.IsSpectator() {
		p.Human.SpawnTo(viewer)
	}
}

// CanCollideWith is a port of Player::canCollideWith.
func (p *Player) CanCollideWith(other world.Entity) bool { return false }

// CanBeCollidedWith is a port of Player::canBeCollidedWith.
func (p *Player) CanBeCollidedWith() bool {
	return !p.IsSpectator() && p.Human.CanBeCollidedWith()
}

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

// CheckGroundState is a port of Player::checkGroundState.
func (p *Player) CheckGroundState(wantedX, wantedY, wantedZ, dx, dy, dz float64) {
	if !p.blockCollision {
		p.OnGround = false
	} else {
		bb := p.BoundingBox
		y := p.GetLocation().Y
		bb.MinY = y - 0.2
		bb.MaxY = y + 0.2

		//we're already at the new position at this point; check if there are blocks we might have landed on between
		//the old and new positions (running down stairs necessitates this)
		bb = bb.AddCoord(-dx, -dy, -dz)

		p.OnGround = len(p.GetWorld().GetCollisionBlocks(bb, true)) > 0
		p.IsCollided = p.OnGround
	}
}

// CanBeMovedByCurrents is a port of Player::canBeMovedByCurrents: currently has no server-side
// movement.
func (p *Player) CanBeMovedByCurrents() bool { return false }

// checkNearEntities is a port of Player::checkNearEntities.
func (p *Player) checkNearEntities() {
	for _, e := range p.GetWorld().GetNearbyEntitiesExcept(p.BoundingBox.ExpandedCopy(1, 0.5, 1), p) {
		if scheduler, ok := e.(interface{ ScheduleUpdate() }); ok {
			scheduler.ScheduleUpdate()
		}

		if !e.IsAlive() || e.IsFlaggedForDespawn() {
			continue
		}

		if c, ok := e.(interface{ OnCollideWithPlayer(player entity.Player) }); ok {
			c.OnCollideWithPlayer(p)
		}
	}
}

// HandleMovement is a port of Player::handleMovement: attempts to move the player to the given
// coordinates (of the player's feet). This is used for processing movements sent by the player
// over network.
func (p *Player) HandleMovement(newPos math.Vector3) {
	timings.Init()
	timings.PlayerMove.StartTiming()
	defer timings.PlayerMove.StopTiming()
	p.actuallyHandleMovement(newPos)
}

func (p *Player) actuallyHandleMovement(newPos math.Vector3) {
	p.moveRateLimit--
	if p.moveRateLimit < 0 {
		return
	}

	oldPos := p.GetLocation()
	distanceSquared := newPos.DistanceSquared(oldPos.Vector3)

	revert := false

	if distanceSquared > 225 { //15 blocks
		//TODO: this is probably too big if we process every movement
		/* !!! BEWARE YE WHO ENTER HERE !!!
		 *
		 * This is NOT an anti-cheat check. It is a safety check.
		 * Without it hackers can teleport with freedom on their own and cause lots of undesirable behaviour, like
		 * freezes, lag spikes and memory exhaustion due to sync chunk loading and collision checks across large distances.
		 * Not only that, but high-latency players can trigger such behaviour innocently.
		 *
		 * If you must tamper with this code, be aware that this can cause very nasty results. Do not waste our time
		 * asking for help if you suffer the consequences of messing with this.
		 */
		p.logger.Debug(fmt.Sprintf("Moved too fast (%g blocks in 1 movement), reverting movement", stdmath.Sqrt(distanceSquared)))
		p.logger.Debug(fmt.Sprintf("Old position: %v, new position: %v", oldPos.Vector3, newPos))
		revert = true
	} else if !p.GetWorld().IsInLoadedTerrain(newPos) {
		revert = true
		p.nextChunkOrderRun = 0
	}

	if !revert && distanceSquared != 0 {
		dx := newPos.X - oldPos.X
		dy := newPos.Y - oldPos.Y
		dz := newPos.Z - oldPos.Z

		p.Move(dx, dy, dz)
	}

	if revert {
		p.revertMovement(oldPos)
	}
}

// toEventLocation converts a location for the move event.
func toEventLocation(l entity.Location) playerevent.Location {
	pos := entityevent.Position{Vector3: l.Vector3}
	if l.IsValid() {
		pos.World = l.GetWorld()
	}
	return playerevent.Location{Position: pos, Yaw: l.Yaw, Pitch: l.Pitch}
}

// processMostRecentMovements is a port of Player::processMostRecentMovements: fires movement
// events and synchronizes player movement, every tick.
func (p *Player) processMostRecentMovements() {
	now := time.Now()
	multiplier := 1.0
	if p.lastMovementProcess != nil {
		multiplier = now.Sub(*p.lastMovementProcess).Seconds() * 20
	}
	exceededRateLimit := p.moveRateLimit < 0
	p.moveRateLimit = stdmath.Min(moveBacklogSize, stdmath.Max(0, p.moveRateLimit)+movesPerTick*multiplier)
	p.lastMovementProcess = &now

	from := p.lastBroadcastLocation
	to := p.GetLocation()

	delta := to.DistanceSquared(from.Vector3)
	deltaAngle := stdmath.Abs(from.Yaw-to.Yaw) + stdmath.Abs(from.Pitch-to.Pitch)

	if delta > 0.0001 || deltaAngle > 1.0 {
		if event.HasHandlers[playerevent.PlayerMoveEvent]() {
			ev := playerevent.NewPlayerMoveEvent(p, toEventLocation(from), toEventLocation(to))

			event.Call(ev)

			if ev.IsCancelled() {
				p.revertMovement(from)
				return
			}

			evTo := ev.GetTo()
			if to.DistanceSquared(evTo.Vector3) > 0.01 { //If plugins modify the destination
				w := p.GetWorld()
				if tw, ok := evTo.World.(*world.World); ok && tw != nil {
					w = tw
				}
				yaw, pitch := evTo.Yaw, evTo.Pitch
				p.TeleportTo(evTo.Vector3, w, &yaw, &pitch)
				return
			}
		}

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

			if p.nextChunkOrderRun > 20 {
				p.nextChunkOrderRun = 20
			}
		}
	}

	if exceededRateLimit { //client and server positions will be out of sync if this happens
		p.logger.Debug("Exceeded movement rate limit, forcing to last accepted position")
		loc := p.GetLocation()
		yaw, pitch := loc.Yaw, loc.Pitch
		p.sendPosition(loc.Vector3, &yaw, &pitch, packet.MoveModeReset)
	}
}

// revertMovement is a port of Player::revertMovement.
func (p *Player) revertMovement(from entity.Location) {
	p.SetPositionInWorld(from.Vector3, from.GetWorld())
	yaw, pitch := from.Yaw, from.Pitch
	p.sendPosition(from.Vector3, &yaw, &pitch, packet.MoveModeReset)
}

// CalculateFallDamage is a port of Player::calculateFallDamage: flying players take no fall
// damage.
func (p *Player) CalculateFallDamage(fallDistance float64) float64 {
	if p.flying {
		return 0
	}
	return p.Human.CalculateFallDamage(fallDistance)
}

// Jump is a port of Player::jump.
func (p *Player) Jump() {
	event.Call(playerevent.NewPlayerJumpEvent(p))
	p.Human.Jump()
}

// SetMotion is a port of Player::setMotion: the player's own client is told about its motion
// too.
func (p *Player) SetMotion(motion math.Vector3) bool {
	if p.Human.SetMotion(motion) {
		p.BroadcastMotion()
		p.GetNetworkSession().SendDataPacket(&packet.SetActorMotion{EntityRuntimeID: uint64(p.GetID()), Velocity: vec32(motion), Tick: 0})
		return true
	}
	return false
}

// UpdateMovement is a no-op for players (Player::updateMovement): movement is broadcast from
// processMostRecentMovements instead.
func (p *Player) UpdateMovement(teleport bool) {}

// TryChangeMovement is a no-op for players (Player::tryChangeMovement).
func (p *Player) TryChangeMovement() {}

// OnUpdate is a port of Player::onUpdate.
func (p *Player) OnUpdate(currentTick int64) bool {
	tickDiff := int(currentTick - p.LastUpdate)

	if tickDiff <= 0 {
		return true
	}

	p.messageCounter = 2

	p.LastUpdate = currentTick

	if p.IsJustCreated() {
		p.OnFirstUpdate(currentTick)
	}

	if !p.IsAlive() && p.spawned {
		p.OnDeathUpdate(tickDiff)
		return true
	}

	if p.spawned {
		timings.Init()
		timings.PlayerMove.StartTiming()
		p.processMostRecentMovements()
		p.SetMotionDirect(math.Vector3Zero()) //TODO: HACK! (Fixes player knockback being messed up)
		if p.OnGround {
			p.inAirTicks = 0
		} else {
			p.inAirTicks += tickDiff
		}
		timings.PlayerMove.StopTiming()

		timings.EntityBaseTick.StartTiming()
		p.EntityBaseTick(tickDiff)
		timings.EntityBaseTick.StopTiming()

		if p.IsCreative() && p.GetFireTicks() > 1 {
			p.SetFireTicks(1)
		}

		if !p.IsSpectator() && p.IsAlive() {
			timings.PlayerCheckNearEntities.StartTiming()
			p.checkNearEntities()
			timings.PlayerCheckNearEntities.StopTiming()
		}

		if p.blockBreakHandler != nil && !p.blockBreakHandler.Update() {
			p.setBlockBreakHandler(nil)
		}

		if p.IsUsingItem() && p.GetItemUseDuration()%4 == 0 {
			if it, ok := p.GetInventory().GetItemInHand().(item.ConsumableItem); ok {
				p.BroadcastAnimation(animation.ConsumingItemAnimation{Entity: p, Item: it}, nil)
			}
		}
	}

	return true
}

// CanEat is a port of Player::canEat.
func (p *Player) CanEat() bool { return p.IsCreative() || p.Human.CanEat() }

// CanBreathe is a port of Player::canBreathe.
func (p *Player) CanBreathe() bool { return p.IsCreative() || p.Human.CanBreathe() }

// CanInteract is a port of Player::canInteract (with the default maxDiff: half of the 3D
// diagonal width of a block).
func (p *Player) CanInteract(pos math.Vector3, maxDistance float64) bool {
	return p.CanInteractWithin(pos, maxDistance, stdmath.Sqrt(3)/2)
}

// CanInteractWithin is Player::canInteract with an explicit maxDiff: whether the player can
// interact with the specified position. This checks distance and direction.
func (p *Player) CanInteractWithin(pos math.Vector3, maxDistance, maxDiff float64) bool {
	eyePos := p.GetEyePos()
	if eyePos.DistanceSquared(pos) > maxDistance*maxDistance {
		return false
	}

	dV := p.GetDirectionVector()
	eyeDot := dV.Dot(eyePos)
	targetDot := dV.Dot(pos)
	return (targetDot - eyeDot) >= -maxDiff
}

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

// playerFlagSleep is PlayerMetadataFlags::SLEEP.
const playerFlagSleep = 1

// SyncNetworkData is a port of Player::syncNetworkData.
func (p *Player) SyncNetworkData(properties *entity.MetadataCollection) {
	p.Human.SyncNetworkData(properties)

	properties.SetGenericFlag(entity.FlagAction, p.startAction > -1)
	properties.SetGenericFlag(entity.FlagHasCollision, p.HasBlockCollision())

	properties.SetPlayerFlag(playerFlagSleep, p.sleeping != nil)
	if p.sleeping != nil {
		//this should only be sent when the player enters the bed, as of 1.26.??
		//previously we were setting this to 0,0,0 if the player wasn't sleeping, but that now causes the player to
		//teleport to that position temporarily when leaving the bed. Bugrock moment...
		properties.SetBlockPos(protocol.EntityDataKeyBedPosition, blockPos(*p.sleeping))
	}

	if p.deathPosition != nil && p.deathPosition.World == p.GetWorld() {
		properties.SetBlockPos(protocol.EntityDataKeyPlayerLastDeathPosition, blockPos(p.deathPosition.Vector3))
		//TODO: this should be updated when dimensions are implemented
		properties.SetInt(protocol.EntityDataKeyPlayerLastDeathDimension, packet.DimensionOverworld)
		properties.SetByte(protocol.EntityDataKeyPlayerHasDied, 1)
	} else {
		properties.SetBlockPos(protocol.EntityDataKeyPlayerLastDeathPosition, protocol.BlockPos{})
		properties.SetInt(protocol.EntityDataKeyPlayerLastDeathDimension, packet.DimensionOverworld)
		properties.SetByte(protocol.EntityDataKeyPlayerHasDied, 0)
	}
}

func blockPos(v math.Vector3) protocol.BlockPos {
	return protocol.BlockPos{int32(v.FloorX()), int32(v.FloorY()), int32(v.FloorZ())}
}

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

// sendPosition is a port of Player::sendPosition.
func (p *Player) sendPosition(pos math.Vector3, yaw, pitch *float64, mode byte) {
	if p.networkSession != nil {
		p.networkSession.SyncMovement(pos, yaw, pitch, mode)
	}
	p.SetYSize(0)
}

// Teleport is Player::teleport with the player's own rotation.
func (p *Player) Teleport(pos math.Vector3) bool { return p.TeleportTo(pos, nil, nil, nil) }

// TeleportTo is a port of Player::teleport: w nil means the current world, nil yaw/pitch keep the
// current rotation.
func (p *Player) TeleportTo(pos math.Vector3, w *world.World, yaw, pitch *float64) bool {
	if p.Human.TeleportTo(pos, w, yaw, pitch) {
		p.RemoveCurrentWindow()
		p.StopSleep()

		loc := p.GetLocation()
		locYaw, locPitch := loc.Yaw, loc.Pitch
		p.sendPosition(loc.Vector3, &locYaw, &locPitch, packet.MoveModeTeleport)
		p.BroadcastMovement(true)

		p.SpawnToAll()

		p.ResetFallDistance()
		p.nextChunkOrderRun = 0
		if p.spawnChunkLoadCount != -1 {
			p.spawnChunkLoadCount = 0
		}
		p.setBlockBreakHandler(nil)

		//TODO: workaround for player last pos not getting updated
		//Entity::updateMovement() normally handles this, but it's overridden with an empty function in Player
		p.ResetLastMovements()
		p.lastBroadcastLocation = p.GetLocation()

		return true
	}
	return false
}

func vec32(v math.Vector3) mgl32.Vec3 {
	return mgl32.Vec3{float32(v.X), float32(v.Y), float32(v.Z)}
}
