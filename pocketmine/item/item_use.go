package item

import (
	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/entity/effect"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/world/sound"
)

// This file holds the item side of the player item-use flow (Player::useHeldItem,
// consumeHeldItem, releaseHeldItem): consumables, releasables, and the items that act on
// onClickAir/onReleaseUsing.

// Releasable is a port of pocketmine\item\Releasable: items which are used by holding right
// click, then releasing it (bows, tridents, food, ...).
type Releasable interface {
	CanStartUsingItem(player Player) bool
}

// ConsumableItem is a port of pocketmine\item\ConsumableItem: an item that can be consumed by an
// entity.
type ConsumableItem interface {
	Item
	Releasable
	// GetAdditionalEffects returns effects to be added to the consumer on consumption.
	GetAdditionalEffects() []*effect.EffectInstance
	// OnConsume is called when this item is consumed by an entity, after the standard results.
	OnConsume(consumer effect.Living)
	// GetResidue returns the leftover item after the item is consumed (e.g. a bowl).
	GetResidue() Item
}

// canEater is the part of Player Food::canStartUsingItem needs.
type canEater interface{ CanEat() bool }

// GetResidue is a port of Food::getResidue.
func (f *Food) GetResidue() Item { return VanillaAir() }

// CanStartUsingItem is a port of Food::canStartUsingItem.
func (f *Food) CanStartUsingItem(player Player) bool {
	if requirer, ok := f.self.(interface{ RequiresHunger() bool }); ok && !requirer.RequiresHunger() {
		return true
	}
	e, ok := player.(canEater)
	return ok && e.CanEat()
}

// GetResidue is a port of BeetrootSoup::getResidue.
func (b *BeetrootSoup) GetResidue() Item { return VanillaBowl() }

// GetResidue is a port of MushroomStew::getResidue.
func (m *MushroomStew) GetResidue() Item { return VanillaBowl() }

// GetResidue is a port of RabbitStew::getResidue.
func (r *RabbitStew) GetResidue() Item { return VanillaBowl() }

// GetResidue is a port of SuspiciousStew::getResidue.
func (s *SuspiciousStew) GetResidue() Item { return VanillaBowl() }

// GetResidue is a port of HoneyBottle::getResidue.
func (h *HoneyBottle) GetResidue() Item { return VanillaGlassBottle() }

// GetResidue is a port of Potion::getResidue.
func (p *Potion) GetResidue() Item { return VanillaGlassBottle() }

// CanStartUsingItem is a port of Potion::canStartUsingItem.
func (p *Potion) CanStartUsingItem(player Player) bool { return true }

// GetResidue is a port of Medicine::getResidue.
func (m *Medicine) GetResidue() Item { return VanillaGlassBottle() }

// effectHolder is the part of Player Medicine::canStartUsingItem needs.
type effectHolder interface {
	GetEffects() *effect.EffectManager
}

// CanStartUsingItem is a port of Medicine::canStartUsingItem.
func (m *Medicine) CanStartUsingItem(player Player) bool {
	h, ok := player.(effectHolder)
	return ok && h.GetEffects().Has(m.MedicineTypeValue.GetCuredEffect())
}

// GetResidue is a port of MilkBucket::getResidue.
func (m *MilkBucket) GetResidue() Item { return VanillaBucket() }

// CanStartUsingItem is a port of MilkBucket::canStartUsingItem.
func (m *MilkBucket) CanStartUsingItem(player Player) bool { return true }

// CanStartUsingItem is a port of Spyglass::canStartUsingItem.
func (s *Spyglass) CanStartUsingItem(player Player) bool { return true }

// CanStartUsingItem is a port of GoatHorn::canStartUsingItem. (GoatHorn::onClickAir plays
// GoatHornSound, which the sound package doesn't have yet.)
func (g *GoatHorn) CanStartUsingItem(player Player) bool { return true }

// armorWearer is the part of Player Armor::onClickAir needs.
type armorWearer interface {
	GetArmorItem(slot int) Item
	SetArmorItem(slot int, it Item)
	SetItemInHand(it Item)
	BroadcastSound(s sound.Sound)
}

// OnClickAir is a port of Armor::onClickAir: the armor is equipped, swapping out what was worn.
func (a *Armor) OnClickAir(player Player, directionVector math.Vector3, returnedItems *[]Item) ItemUseResult {
	w, ok := player.(armorWearer)
	if !ok {
		return ItemUseResultNone
	}
	existing := w.GetArmorItem(a.GetArmorSlot())
	thisCopy := a.Clone().(*Armor)
	newItem := thisCopy.PopCount(1).(*Armor)
	w.SetArmorItem(a.GetArmorSlot(), newItem)
	w.SetItemInHand(existing)
	if s := newItem.GetMaterial().GetEquipSound(); s != nil {
		w.BroadcastSound(s)
	}
	if !thisCopy.IsNull() {
		//if the stack size was bigger than 1 (usually won't happen, but might be caused by plugins)
		*returnedItems = append(*returnedItems, thisCopy)
	}
	return ItemUseResultSuccess
}

// ThrowProjectileFunc is ProjectileItem::onClickAir minus the final pop: it creates the item's
// projectile entity at the player's eyes (createEntity), launches it and plays the throw sound.
// The entity/projectile package installs it, since the entities can't be created here.
var ThrowProjectileFunc func(it ProjectileItem, player Player, directionVector math.Vector3) ItemUseResult

func throwProjectile(it ProjectileItem, player Player, directionVector math.Vector3) ItemUseResult {
	if ThrowProjectileFunc == nil {
		return ItemUseResultNone
	}
	result := ThrowProjectileFunc(it, player, directionVector)
	if result == ItemUseResultSuccess {
		it.Pop()
	}
	return result
}

// OnClickAir is a port of ProjectileItem::onClickAir.
func (t *throwableItem) OnClickAir(player Player, directionVector math.Vector3, returnedItems *[]Item) ItemUseResult {
	return throwProjectile(t.self.(ProjectileItem), player, directionVector)
}

// ReleaseBowFunc is Bow::onReleaseUsing's body (creating and shooting the arrow entity), installed
// by the entity/projectile package.
var ReleaseBowFunc func(bow *Bow, player Player, returnedItems *[]Item) ItemUseResult

// ReleaseTridentFunc is Trident::onReleaseUsing's body (throwing the trident entity), installed by
// the entity/projectile package.
var ReleaseTridentFunc func(trident *Trident, player Player, returnedItems *[]Item) ItemUseResult

// TridentDamageOnThrow is Trident::DAMAGE_ON_THROW.
const TridentDamageOnThrow = 1

// CanStartUsingItem is a port of Trident::canStartUsingItem.
func (t *Trident) CanStartUsingItem(player Player) bool {
	return t.Damage < t.GetMaxDurability()-TridentDamageOnThrow
}

// OnReleaseUsing is a port of Trident::onReleaseUsing.
func (t *Trident) OnReleaseUsing(player Player, returnedItems *[]Item) ItemUseResult {
	if ReleaseTridentFunc == nil {
		return ItemUseResultNone
	}
	return ReleaseTridentFunc(t, player, returnedItems)
}

// OnAttackEntity is a port of Trident::onAttackEntity.
func (t *Trident) OnAttackEntity(victim Entity, returnedItems *[]Item) bool {
	return t.ApplyDamage(1)
}

// OnDestroyBlock is a port of Trident::onDestroyBlock.
func (t *Trident) OnDestroyBlock(blk block.Behavior, returnedItems *[]Item) bool {
	return damageUnlessInstant(blk, t.ApplyDamage, 2)
}

// Bow is a port of pocketmine\item\Bow.
type Bow struct {
	Tool
}

func NewBow(identifier ItemIdentifier, name string, enchantmentTags ...string) *Bow {
	b := &Bow{}
	b.Init(b, identifier, name)
	b.enchantmentTags = enchantmentTags
	return b
}

func (b *Bow) Clone() Item {
	c := *b
	c.rebind(&c)
	return &c
}

func (b *Bow) GetFuelTime() int { return 200 }

func (b *Bow) GetMaxDurability() int { return 385 }

// OnReleaseUsing is a port of Bow::onReleaseUsing.
func (b *Bow) OnReleaseUsing(player Player, returnedItems *[]Item) ItemUseResult {
	if ReleaseBowFunc == nil {
		return ItemUseResultNone
	}
	return ReleaseBowFunc(b, player, returnedItems)
}

// arrowHolder is the part of Player Bow::canStartUsingItem needs.
type arrowHolder interface {
	HasFiniteResources() bool
	HasArrow() bool
}

// CanStartUsingItem is a port of Bow::canStartUsingItem: the player must have an arrow in the
// off-hand or inventory, unless in creative.
func (b *Bow) CanStartUsingItem(player Player) bool {
	h, ok := player.(arrowHolder)
	return ok && (!h.HasFiniteResources() || h.HasArrow())
}
