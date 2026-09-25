package player

import (
	"fmt"
	"strconv"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/entity"
	"pocketmine-go/pocketmine/entity/animation"
	"pocketmine-go/pocketmine/entity/effect"
	"pocketmine-go/pocketmine/entity/object"
	"pocketmine-go/pocketmine/entity/projectile"
	"pocketmine-go/pocketmine/event"
	entityevent "pocketmine-go/pocketmine/event/entity"
	playerevent "pocketmine-go/pocketmine/event/player"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/item/enchantment"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/world"
	"pocketmine-go/pocketmine/world/particle"
	"pocketmine-go/pocketmine/world/sound"
)

// cooldownKey is `$item->getCooldownTag() ?? $item->getStateId()`.
func cooldownKey(it item.Item) string {
	if tag, ok := it.GetCooldownTag(); ok {
		return tag
	}
	return "#" + strconv.Itoa(it.GetStateId())
}

// GetItemCooldownExpiry returns the server tick on which the player's cooldown period expires for
// the given item.
func (p *Player) GetItemCooldownExpiry(it item.Item) int64 {
	p.checkItemCooldowns()
	return p.usedItemsCooldown[cooldownKey(it)]
}

// HasItemCooldown returns whether the player has a cooldown period left before it can use the
// given item again.
func (p *Player) HasItemCooldown(it item.Item) bool {
	p.checkItemCooldowns()
	_, ok := p.usedItemsCooldown[cooldownKey(it)]
	return ok
}

// ResetItemCooldown resets the player's cooldown time for the given item back to the maximum;
// ticks nil uses the item's own cooldown.
func (p *Player) ResetItemCooldown(it item.Item, ticks *int) {
	t := it.GetCooldownTicks()
	if ticks != nil {
		t = *ticks
	}
	if t > 0 {
		p.usedItemsCooldown[cooldownKey(it)] = p.server.GetTick() + int64(t)
		p.GetNetworkSession().OnItemCooldownChanged(it, t)
	}
}

// checkItemCooldowns is a port of Player::checkItemCooldowns.
func (p *Player) checkItemCooldowns() {
	serverTick := p.server.GetTick()
	for key, cooldownUntil := range p.usedItemsCooldown {
		if cooldownUntil <= serverTick {
			delete(p.usedItemsCooldown, key)
		}
	}
}

// Chat is in chat.go.

// SelectHotbarSlot is a port of Player::selectHotbarSlot.
func (p *Player) SelectHotbarSlot(hotbarSlot int) bool {
	inventory := p.GetInventory()
	if !inventory.IsHotbarSlot(hotbarSlot) { //TODO: exception here?
		return false
	}
	if hotbarSlot == inventory.GetHeldItemIndex() {
		return true
	}

	ev := playerevent.NewPlayerItemHeldEvent(p, inventory.GetItem(hotbarSlot), hotbarSlot)
	event.Call(ev)
	if ev.IsCancelled() {
		return false
	}

	inventory.SetHeldItemIndex(hotbarSlot)
	p.SetUsingItem(false)

	return true
}

// durable is item.Durable as returnItemsFromAction needs it.
type durable interface {
	item.Item
	GetDamage() int
	SetDamage(damage int)
	GetMaxDurability() int
	IsBroken() bool
}

// returnItemsFromAction is a port of Player::returnItemsFromAction.
func (p *Player) returnItemsFromAction(oldHeldItem, newHeldItem item.Item, extraReturnedItems []item.Item) {
	heldItemChanged := false

	if !newHeldItem.EqualsExact(oldHeldItem) && oldHeldItem.EqualsExact(p.GetInventory().GetItemInHand()) {
		//determine if the item was changed in some meaningful way, or just damaged/changed count
		//if it was really changed we always need to set it, whether we have finite resources or not
		newReplica := oldHeldItem.Clone()
		newReplica.SetCount(newHeldItem.GetCount())
		if replica, ok := newReplica.(durable); ok {
			if newDurable, ok := newHeldItem.(durable); ok {
				newDamage := newDurable.GetDamage()
				if newDamage >= 0 && newDamage <= replica.GetMaxDurability() {
					replica.SetDamage(newDamage)
				}
			}
		}
		damagedOrDeducted := newReplica.EqualsExact(newHeldItem)

		if !damagedOrDeducted || p.HasFiniteResources() {
			if d, ok := newHeldItem.(durable); ok && d.IsBroken() {
				p.BroadcastSound(sound.ItemBreakSound{})
			}
			p.GetInventory().SetItemInHand(newHeldItem)
			heldItemChanged = true
		}
	}

	if !heldItemChanged {
		newHeldItem = oldHeldItem
	}

	if heldItemChanged && len(extraReturnedItems) > 0 && newHeldItem.IsNull() {
		p.GetInventory().SetItemInHand(extraReturnedItems[0])
		extraReturnedItems = extraReturnedItems[1:]
	}
	for _, drop := range p.GetInventory().AddItem(extraReturnedItems...) {
		//TODO: we can't generate a transaction for this since the items aren't coming from an inventory :(
		ev := playerevent.NewPlayerDropItemEvent(p, drop)
		if p.IsSpectator() {
			ev.Cancel()
		}
		event.Call(ev)
		if !ev.IsCancelled() {
			p.DropItem(drop)
		}
	}
}

// UseHeldItem is a port of Player::useHeldItem: activates the item in hand, for example throwing
// a projectile. Returns whether it did something.
func (p *Player) UseHeldItem() bool {
	directionVector := p.GetDirectionVector()
	it := p.GetInventory().GetItemInHand()
	oldItem := it.Clone()

	ev := playerevent.NewPlayerItemUseEvent(p, it, directionVector)
	if p.HasItemCooldown(it) || p.IsSpectator() {
		ev.Cancel()
	}

	event.Call(ev)

	if ev.IsCancelled() {
		return false
	}

	var returnedItems []item.Item
	result := it.OnClickAir(p, directionVector, &returnedItems)
	if result == item.ItemUseResultFail {
		return false
	}

	p.ResetItemCooldown(oldItem, nil)
	p.returnItemsFromAction(oldItem, it, returnedItems)

	releasable, isReleasable := it.(item.Releasable)
	p.SetUsingItem(isReleasable && releasable.CanStartUsingItem(p))

	return true
}

// ConsumeHeldItem is a port of Player::consumeHeldItem: consumes the currently-held item.
// Returns whether the consumption succeeded.
func (p *Player) ConsumeHeldItem() bool {
	slot, ok := p.GetInventory().GetItemInHand().(item.ConsumableItem)
	if !ok {
		return false
	}
	oldItem := slot.Clone()

	residue := slot.GetResidue()
	var residues []playerevent.Item
	if !residue.IsNull() {
		residues = []playerevent.Item{residue}
	}
	ev := playerevent.NewPlayerItemConsumeEvent(p, slot, residues)
	if p.HasItemCooldown(slot) {
		ev.Cancel()
	}
	event.Call(ev)

	if ev.IsCancelled() || !p.ConsumeObject(slot) {
		return false
	}

	p.SetUsingItem(false)
	p.ResetItemCooldown(oldItem, nil)

	slot.Pop()
	var returned []item.Item
	for _, r := range ev.GetResidue() {
		if ri, ok := r.(item.Item); ok {
			returned = append(returned, ri)
		}
	}
	p.returnItemsFromAction(oldItem, slot, returned)

	return true
}

// ReleaseHeldItem is a port of Player::releaseHeldItem: releases the held item, for example to
// fire a bow. This should be preceded by a call to UseHeldItem. Returns whether it did something.
func (p *Player) ReleaseHeldItem() bool {
	defer p.SetUsingItem(false)

	it := p.GetInventory().GetItemInHand()
	if !p.IsUsingItem() || p.HasItemCooldown(it) {
		return false
	}

	oldItem := it.Clone()

	var returnedItems []item.Item
	result := it.OnReleaseUsing(p, &returnedItems)
	if result == item.ItemUseResultSuccess {
		p.ResetItemCooldown(oldItem, nil)
		p.returnItemsFromAction(oldItem, it, returnedItems)
		return true
	}

	return false
}

// PickBlock is a port of Player::pickBlock.
func (p *Player) PickBlock(pos math.Vector3, addTileNBT bool) bool {
	blk := p.GetWorld().GetBlock(pos)
	if _, unknown := blk.(*block.UnknownBlock); unknown {
		return true
	}

	picked, ok := blk.GetPickedItem(addTileNBT).(item.Item)
	if !ok || picked == nil {
		return true
	}

	ev := playerevent.NewPlayerBlockPickEvent(p, blk, picked)
	existingSlot := p.GetInventory().First(picked, false)
	if existingSlot == -1 && p.HasFiniteResources() {
		ev.Cancel()
	}
	event.Call(ev)

	if !ev.IsCancelled() {
		p.equipOrAddPickedItem(existingSlot, picked)
	}

	return true
}

// PickEntity is a port of Player::pickEntity.
func (p *Player) PickEntity(entityID int) bool {
	e, ok := p.GetWorld().GetEntity(entityID)
	//TODO: HACK! We really shouldn't be keeping disconnected players (and generally flagged-for-despawn entities)
	//in the world's entity table, but changing that is too risky for a hotfix. This workaround will do for now.
	if !ok || e.IsFlaggedForDespawn() {
		return true
	}

	pickable, ok := e.(interface{ GetPickedItem() item.Item })
	if !ok {
		return true
	}
	picked := pickable.GetPickedItem()
	if picked == nil {
		return true
	}

	ev := playerevent.NewPlayerEntityPickEvent(p, e, picked)
	existingSlot := p.GetInventory().First(picked, false)
	if existingSlot == -1 && (p.HasFiniteResources() || p.IsSpectator()) {
		ev.Cancel()
	}
	event.Call(ev)

	if !ev.IsCancelled() {
		p.equipOrAddPickedItem(existingSlot, picked)
	}

	return true
}

// equipOrAddPickedItem is a port of Player::equipOrAddPickedItem.
func (p *Player) equipOrAddPickedItem(existingSlot int, it item.Item) {
	inv := p.GetInventory()
	if existingSlot != -1 {
		if existingSlot < inv.GetHotbarSize() {
			inv.SetHeldItemIndex(existingSlot)
		} else {
			inv.Swap(inv.GetHeldItemIndex(), existingSlot)
		}
	} else {
		firstEmpty := inv.FirstEmpty()
		if firstEmpty == -1 { //full inventory
			inv.SetItemInHand(it)
		} else if firstEmpty < inv.GetHotbarSize() {
			inv.SetItem(firstEmpty, it)
			inv.SetHeldItemIndex(firstEmpty)
		} else {
			inv.Swap(inv.GetHeldItemIndex(), firstEmpty)
			inv.SetItemInHand(it)
		}
	}
}

// blockTypeTagsFire is BlockTypeTags::FIRE.
const blockTypeTagsFire = "pocketmine:fire"

// AttackBlock is a port of Player::attackBlock: performs a left-click (attack) action on the
// block. Returns whether an action took place successfully.
func (p *Player) AttackBlock(pos math.Vector3, face math.Facing) bool {
	if pos.DistanceSquared(p.GetLocation().Vector3) > 10000 {
		return false //TODO: maybe this should throw an exception instead?
	}

	target := p.GetWorld().GetBlock(pos)

	ev := playerevent.NewPlayerInteractEvent(p, p.GetInventory().GetItemInHand(), target, nil, int(face), playerevent.InteractLeftClickBlock)
	if p.IsSpectator() {
		ev.Cancel()
	}
	event.Call(ev)
	if ev.IsCancelled() {
		return false
	}
	p.BroadcastAnimation(animation.ArmSwingAnimation{Entity: p}, p.GetViewers())
	if target.OnAttack(p.GetInventory().GetItemInHand(), face, p) {
		return true
	}

	side := target.(interface {
		GetSide(side math.Facing, step int) block.Behavior
	}).GetSide(face, 1)
	if tagged, ok := side.(interface{ HasTypeTag(tag string) bool }); ok && tagged.HasTypeTag(blockTypeTagsFire) {
		sidePos := side.GetPosition()
		_ = p.GetWorld().SetBlock(sidePos, block.VanillaAir())
		p.GetWorld().AddSound(sidePos.Add(0.5, 0.5, 0.5), sound.FireExtinguishSound{})
		return true
	}

	if !p.IsCreative() && !target.GetBreakInfo().BreaksInstantly() {
		p.setBlockBreakHandler(NewSurvivalBlockBreakHandler(p, pos, target, face, 16, DefaultFxIntervalTicks))
	}

	return true
}

// setBlockBreakHandler replaces the block break handler; the old one's BLOCK_STOP_BREAK is sent
// (SurvivalBlockBreakHandler::__destruct runs when PHP drops the reference).
func (p *Player) setBlockBreakHandler(h *SurvivalBlockBreakHandler) {
	if p.blockBreakHandler != nil && p.blockBreakHandler != h {
		p.blockBreakHandler.Close()
	}
	p.blockBreakHandler = h
}

// GetBlockBreakHandler returns the handler of the block the player is breaking, or nil.
func (p *Player) GetBlockBreakHandler() *SurvivalBlockBreakHandler { return p.blockBreakHandler }

// ContinueBreakBlock is a port of Player::continueBreakBlock.
func (p *Player) ContinueBreakBlock(pos math.Vector3, face math.Facing) {
	if p.blockBreakHandler != nil && p.blockBreakHandler.GetBlockPos().DistanceSquared(pos) < 0.0001 {
		p.blockBreakHandler.SetTargetedFace(face)
	}
}

// StopBreakBlock is a port of Player::stopBreakBlock.
func (p *Player) StopBreakBlock(pos math.Vector3) {
	if p.blockBreakHandler != nil && p.blockBreakHandler.GetBlockPos().DistanceSquared(pos) < 0.0001 {
		p.setBlockBreakHandler(nil)
	}
}

// BreakBlock is a port of Player::breakBlock: breaks the block at the given position using the
// currently-held item. Returns whether the block was broken; false means a rollback needs to
// take place.
func (p *Player) BreakBlock(pos math.Vector3) bool {
	p.RemoveCurrentWindow()

	maxDistance := float64(maxReachDistanceSurvival)
	if p.IsCreative() {
		maxDistance = maxReachDistanceCreative
	}
	if p.CanInteract(pos.Add(0.5, 0.5, 0.5), maxDistance) {
		p.BroadcastAnimation(animation.ArmSwingAnimation{Entity: p}, p.GetViewers())
		p.StopBreakBlock(pos)
		it := p.GetInventory().GetItemInHand()
		oldItem := it.Clone()
		var returnedItems []item.Item
		if p.GetWorld().UseBreakOnWith(pos, it, p, true, &returnedItems) {
			p.returnItemsFromAction(oldItem, it, returnedItems)
			p.GetHungerManager().Exhaust(0.005, playerevent.ExhaustCauseMining)
			return true
		}
	} else {
		p.logger.Debug(fmt.Sprintf("Cancelled block break at %v due to not currently being interactable", pos))
	}

	return false
}

// InteractBlock is a port of Player::interactBlock: touches the block at the given position with
// the currently-held item. Returns whether it did something.
func (p *Player) InteractBlock(pos math.Vector3, face math.Facing, clickOffset math.Vector3) bool {
	p.SetUsingItem(false)

	maxDistance := float64(maxReachDistanceSurvival)
	if p.IsCreative() {
		maxDistance = maxReachDistanceCreative
	}
	if p.CanInteract(pos.Add(0.5, 0.5, 0.5), maxDistance) {
		p.BroadcastAnimation(animation.ArmSwingAnimation{Entity: p}, p.GetViewers())
		it := p.GetInventory().GetItemInHand() //this is a copy of the real item
		oldItem := it.Clone()
		var returnedItems []item.Item
		if p.GetWorld().UseItemOn(pos, it, face, &clickOffset, p, true, &returnedItems) {
			p.returnItemsFromAction(oldItem, it, returnedItems)
			return true
		}
	} else {
		p.logger.Debug(fmt.Sprintf("Cancelled interaction of block at %v due to not currently being interactable", pos))
	}

	return false
}

// livingTarget is the part of Living AttackEntity needs.
type livingTarget interface {
	world.Entity
	IsLiving() bool
	BroadcastAnimation(anim animation.Animation, targets []world.EntityViewer)
}

// AttackEntity is a port of Player::attackEntity: attacks the given entity with the
// currently-held item. Returns whether the entity was dealt damage.
func (p *Player) AttackEntity(target world.Entity) bool {
	if !target.IsAlive() {
		return false
	}
	switch target.(type) {
	case *object.ItemEntity, *projectile.Arrow:
		p.logger.Debug(fmt.Sprintf("Attempted to attack non-attackable entity %T", target))
		return false
	}

	heldItem := p.GetInventory().GetItemInHand()
	oldItem := heldItem.Clone()

	ev := entityevent.NewEntityDamageByEntityEvent(p, target, entityevent.CauseEntityAttack, float64(heldItem.GetAttackPoints()), nil)
	if !p.CanInteract(target.GetPosition(), maxReachDistanceEntityInteraction) {
		p.logger.Debug(fmt.Sprintf("Cancelled attack of entity %d due to not currently being interactable", target.GetID()))
		ev.Cancel()
	} else if _, isPlayer := target.(*Player); p.IsSpectator() || (isPlayer && !p.server.GetConfigBool("pvp", true)) {
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

	if living, ok := target.(livingTarget); ok && living.IsLiving() {
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
		var returnedItems []item.Item
		heldItem.OnAttackEntity(target, &returnedItems)
		p.returnItemsFromAction(oldItem, heldItem, returnedItems)

		p.GetHungerManager().Exhaust(0.1, playerevent.ExhaustCauseAttack)
	}

	return true
}

// MissSwing is a port of Player::missSwing: performs actions associated with the attack action
// (left-click) without a target entity.
func (p *Player) MissSwing() {
	ev := playerevent.NewPlayerMissSwingEvent(p)
	event.Call(ev)
	if !ev.IsCancelled() {
		p.BroadcastSound(sound.EntityAttackNoDamageSound{})
		p.BroadcastAnimation(animation.ArmSwingAnimation{Entity: p}, p.GetViewers())
	}
}

// interactable is the part of Entity InteractEntity needs.
type interactable interface {
	world.Entity
	OnInteract(player entity.Player, clickPos math.Vector3) bool
}

// InteractEntity is a port of Player::interactEntity: interacts with the given entity using the
// currently-held item.
func (p *Player) InteractEntity(target world.Entity, clickPos math.Vector3) bool {
	ev := playerevent.NewPlayerEntityInteractEvent(p, target, clickPos)

	if !p.CanInteract(target.GetPosition(), maxReachDistanceEntityInteraction) {
		p.logger.Debug(fmt.Sprintf("Cancelled interaction with entity %d due to not currently being interactable", target.GetID()))
		ev.Cancel()
	}

	event.Call(ev)

	it := p.GetInventory().GetItemInHand()
	oldItem := it.Clone()
	if ev.IsCancelled() {
		return false
	}
	if it.OnInteractEntity(p, target, clickPos) {
		if p.HasFiniteResources() && !it.EqualsExact(oldItem) && oldItem.EqualsExact(p.GetInventory().GetItemInHand()) {
			if d, ok := it.(durable); ok && d.IsBroken() {
				p.BroadcastSound(sound.ItemBreakSound{})
			}
			p.GetInventory().SetItemInHand(it)
		}
	}
	if i, ok := target.(interactable); ok {
		return i.OnInteract(p, clickPos)
	}
	return false
}

// Emote is a port of Player::emote.
func (p *Player) Emote(emoteID string) {
	currentTick := p.server.GetTick()
	if currentTick-p.lastEmoteTick > 5 {
		p.lastEmoteTick = currentTick
		ev := playerevent.NewPlayerEmoteEvent(p, emoteID)
		event.Call(ev)
		if !ev.IsCancelled() {
			p.Human.Emote(ev.GetEmoteId())
		}
	}
}

// DropItem is a port of Player::dropItem: drops an item on the ground in front of the player.
func (p *Player) DropItem(it item.Item) {
	p.BroadcastAnimation(animation.ArmSwingAnimation{Entity: p}, p.GetViewers())
	motion := p.GetDirectionVector().Multiply(0.4)
	p.GetWorld().DropItem(p.GetLocation().Add(0, 1.3, 0), it, &motion, 40)
}

// HasArrow and TakeArrow are what Bow::onReleaseUsing reads arrows through: the off-hand is
// checked before the inventory.
func (p *Player) HasArrow() bool {
	arrow := item.VanillaArrow()
	return p.GetOffHandInventory().Contains(arrow) || p.GetInventory().Contains(arrow)
}

func (p *Player) TakeArrow() bool {
	arrow := item.VanillaArrow()
	if p.GetOffHandInventory().Contains(arrow) {
		p.GetOffHandInventory().RemoveItem(arrow)
		return true
	}
	if p.GetInventory().Contains(arrow) {
		p.GetInventory().RemoveItem(arrow)
		return true
	}
	return false
}

// GetArmorItem/SetArmorItem/SetItemInHand are what Armor::onClickAir needs.
func (p *Player) GetArmorItem(slot int) item.Item     { return p.GetArmorInventory().GetItem(slot) }
func (p *Player) SetArmorItem(slot int, it item.Item) { p.GetArmorInventory().SetItem(slot, it) }
func (p *Player) SetItemInHand(it item.Item)          { p.GetInventory().SetItemInHand(it) }

var _ = particle.BlockBreakParticle{}
