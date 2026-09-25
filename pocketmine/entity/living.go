package entity

import (
	"iter"
	stdmath "math"
	"math/rand/v2"
	"sort"

	"pocketmine-go/pocketmine/binaryutils"
	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/data/bedrock"
	"pocketmine-go/pocketmine/entity/animation"
	"pocketmine-go/pocketmine/entity/effect"
	entityevent "pocketmine-go/pocketmine/event/entity"
	"pocketmine-go/pocketmine/inventory"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/item/enchantment"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
	"pocketmine-go/pocketmine/utils"
	"pocketmine-go/pocketmine/world"
	"pocketmine-go/pocketmine/world/sound"
)

// defaultBreathTicks mirrors Living::DEFAULT_BREATH_TICKS.
const defaultBreathTicks = 300

// DefaultKnockbackForce/DefaultKnockbackVerticalLimit mirror Living::DEFAULT_KNOCKBACK_FORCE/
// DEFAULT_KNOCKBACK_VERTICAL_LIMIT (declared in event/entity, which needs them for
// EntityDamageByEntityEvent's defaults).
const (
	DefaultKnockbackForce         = entityevent.DefaultKnockbackForce
	DefaultKnockbackVerticalLimit = entityevent.DefaultKnockbackVerticalLimit
)

// Living NBT keys, a port of Living's TAG_* constants.
const (
	tagLegacyHealth        = "HealF"         //TAG_Float
	tagHealth              = "Health"        //TAG_Float
	tagBreathTicks         = "Air"           //TAG_Short
	tagActiveEffects       = "ActiveEffects" //TAG_List<TAG_Compound>
	tagEffectID            = "Id"            //TAG_Byte
	tagEffectDuration      = "Duration"      //TAG_Int
	tagEffectAmplifier     = "Amplifier"     //TAG_Byte
	tagEffectShowParticles = "ShowParticles" //TAG_Byte
	tagEffectAmbient       = "Ambient"       //TAG_Byte
)

// LivingHooks extends Hooks with the Living methods Living's own bodies call on $this.
type LivingHooks interface {
	Hooks
	effect.Living

	GetName() string
	GetSneakOffset() float64
	ApplyDamageModifiers(source entityevent.DamageSource)
	ApplyPostDamageEffects(source entityevent.DamageSource)
	DoHitAnimation()
	KnockBack(x, z, force, verticalLimit float64)
	OnAirExpired()
	CanBreathe() bool
	GetDrops() []item.Item
	GetXpDropAmount() int
	StartDeathAnimation()
	EndDeathAnimation()
	CalculateFallDamage(fallDistance float64) float64
	ApplyConsumptionResults(consumable Consumable)
	SetSprinting(value bool)
	IsSprinting() bool
	IsUnderwater() bool
}

// Living is a port of pocketmine\entity\Living.
type Living struct {
	Entity

	lself LivingHooks

	attackTime int

	DeadTicks    int
	maxDeadTicks int

	jumpVelocity float64

	effectManager *effect.EffectManager

	armorInventory *inventory.ArmorInventory

	breathing      bool
	breathTicks    int
	maxBreathTicks int

	healthAttr              *Attribute
	absorptionAttr          *Attribute
	knockbackResistanceAttr *Attribute
	moveSpeedAttr           *Attribute

	sprinting bool
	sneaking  bool
	gliding   bool
	swimming  bool

	frostWalkerLevel *int
}

// ConstructLiving runs Living's field defaults and then Entity.Construct (see its doc comment).
// Living subtypes call this instead of Construct.
func (l *Living) ConstructLiving(self LivingHooks, location Location, tag *nbt.CompoundTag) {
	l.lself = self
	l.maxDeadTicks = 25
	l.jumpVelocity = 0.42
	l.breathing = true
	l.breathTicks = defaultBreathTicks
	l.maxBreathTicks = defaultBreathTicks
	l.Construct(self, location, tag)
}

func (l *Living) GetInitialDragMultiplier() float64 { return 0.02 }

func (l *Living) GetInitialGravity() float64 { return 0.08 }

func (l *Living) CanBeRenamed() bool { return true }

// InitEntity is a port of Living::initEntity.
func (l *Living) InitEntity(tag *nbt.CompoundTag) {
	l.Entity.InitEntity(tag)

	l.effectManager = effect.NewEffectManager(l.lself)
	markDirtyOnAdd := effect.EffectAddHook(func(*effect.EffectInstance, bool) { l.networkPropertiesDirty = true })
	markDirtyOnRemove := effect.EffectRemoveHook(func(*effect.EffectInstance) { l.networkPropertiesDirty = true })
	l.effectManager.GetEffectAddHooks().Add(&markDirtyOnAdd)
	l.effectManager.GetEffectRemoveHooks().Add(&markDirtyOnRemove)

	l.armorInventory = inventory.NewArmorInventory(l.lself)
	//TODO: load/save armor inventory contents
	l.armorInventory.GetListeners().Add(inventory.OnAnyChange(func(inventory.Inventory) {
		BroadcastPackets(l.GetViewers(), mobArmorChangePacket(l))
	}))
	l.armorInventory.GetListeners().Add(inventory.NewCallbackInventoryListener(
		func(_ inventory.Inventory, slot int, _ item.Item) {
			if slot == inventory.ArmorSlotFeet {
				l.frostWalkerLevel = nil
			}
		},
		func(inventory.Inventory, map[int]item.Item) { l.frostWalkerLevel = nil },
	))

	health := float64(l.lself.GetMaxHealth())

	if healF, ok := tag.GetTag(tagLegacyHealth); ok && isFloatTag(healF) {
		health = float64(healF.(nbt.FloatTag))
	} else if healthTag, ok := tag.GetTag(tagHealth); ok {
		switch v := healthTag.(type) {
		case nbt.ShortTag:
			health = float64(v) //Older versions of PocketMine-MP incorrectly saved this as a short instead of a float
		case nbt.FloatTag:
			health = float64(v)
		}
	}

	l.lself.SetHealth(health)

	l.SetAirSupplyTicks(int(tag.GetShortOr(tagBreathTicks, defaultBreathTicks)))

	if activeEffects, ok, _ := tag.GetListTag(tagActiveEffects); ok {
		for _, t := range activeEffects.Values() {
			e, ok := t.(*nbt.CompoundTag)
			if !ok {
				continue
			}
			effectType, ok := bedrock.EffectIdMap().FromID(int(e.GetByteOr(tagEffectID, 0)))
			if !ok {
				continue
			}

			duration := int(e.GetIntOr(tagEffectDuration, 0))
			infinite := duration == -1
			if infinite {
				duration = binaryutils.Int32Max
			}
			l.effectManager.Add(effect.NewEffectInstanceFull(
				effectType,
				&duration,
				int(uint8(e.GetByteOr(tagEffectAmplifier, 0))),
				e.GetByteOr(tagEffectShowParticles, 1) != 0,
				e.GetByteOr(tagEffectAmbient, 0) != 0,
				nil,
				infinite,
			))
		}
	}
}

func isFloatTag(t nbt.Tag) bool {
	_, ok := t.(nbt.FloatTag)
	return ok
}

// AddAttributes is a port of Living::addAttributes.
func (l *Living) AddAttributes() {
	factory := GetAttributeFactory()
	l.healthAttr = factory.MustGet(AttributeHealth)
	l.attributeMap.Add(l.healthAttr)
	l.attributeMap.Add(factory.MustGet(AttributeFollowRange))
	l.knockbackResistanceAttr = factory.MustGet(AttributeKnockbackResistance)
	l.attributeMap.Add(l.knockbackResistanceAttr)
	l.moveSpeedAttr = factory.MustGet(AttributeMovementSpeed)
	l.attributeMap.Add(l.moveSpeedAttr)
	l.attributeMap.Add(factory.MustGet(AttributeAttackDamage))
	l.absorptionAttr = factory.MustGet(AttributeAbsorption)
	l.attributeMap.Add(l.absorptionAttr)
}

// GetDisplayName returns the name used to describe this entity in chat and command outputs.
func (l *Living) GetDisplayName() string {
	if l.nameTag != "" {
		return l.nameTag
	}
	return l.lself.GetName()
}

// SetHealth is a port of Living::setHealth.
func (l *Living) SetHealth(amount float64) {
	wasAlive := l.IsAlive()
	l.Entity.SetHealth(amount)
	l.healthAttr.SetValue(stdmath.Ceil(l.GetHealth()), true, false)
	if l.IsAlive() && !wasAlive {
		l.lself.BroadcastAnimation(animation.RespawnAnimation{Entity: l}, nil)
	}
}

func (l *Living) GetMaxHealth() int { return int(l.healthAttr.GetMaxValue()) }

func (l *Living) SetMaxHealth(amount int) {
	l.healthAttr.SetMaxValue(float64(amount)).SetDefaultValue(float64(amount))
}

func (l *Living) GetAbsorption() float64 { return l.absorptionAttr.GetValue() }

func (l *Living) SetAbsorption(absorption float64) {
	l.absorptionAttr.SetValue(absorption, false, false)
}

// GetSneakOffset is Living::getSneakOffset's default.
func (l *Living) GetSneakOffset() float64 { return 0.0 }

func (l *Living) IsSneaking() bool { return l.sneaking }

// IsLiving marks a Living for the block package's local Living interface.
func (l *Living) IsLiving() bool { return true }

func (l *Living) SetSneaking(value bool) {
	l.sneaking = value
	l.networkPropertiesDirty = true
	l.recalculateSize()
}

func (l *Living) IsSprinting() bool { return l.sprinting }

// SetSprinting is a port of Living::setSprinting.
func (l *Living) SetSprinting(value bool) {
	if value != l.lself.IsSprinting() {
		l.sprinting = value
		l.networkPropertiesDirty = true
		moveSpeed := l.GetMovementSpeed()
		if value {
			l.SetMovementSpeed(moveSpeed*1.3, false)
		} else {
			l.SetMovementSpeed(moveSpeed/1.3, false)
		}
		l.moveSpeedAttr.MarkSynchronized(false) //TODO: reevaluate this hack
	}
}

func (l *Living) IsGliding() bool { return l.gliding }

func (l *Living) SetGliding(value bool) {
	l.gliding = value
	l.networkPropertiesDirty = true
	l.recalculateSize()
}

func (l *Living) IsSwimming() bool { return l.swimming }

func (l *Living) SetSwimming(value bool) {
	l.swimming = value
	l.networkPropertiesDirty = true
	l.recalculateSize()
}

// recalculateSize is a port of Living::recalculateSize.
func (l *Living) recalculateSize() {
	size := l.self.GetInitialSizeInfo()
	if l.IsSwimming() || l.IsGliding() {
		width := size.GetWidth()
		l.SetSize(NewEntitySizeInfoWithEyeHeight(width, width, width*0.9).Scale(l.GetScale()))
	} else if l.IsSneaking() {
		offset := l.lself.GetSneakOffset()
		l.SetSize(NewEntitySizeInfoWithEyeHeight(size.GetHeight()-offset, size.GetWidth(), size.GetEyeHeight()-offset).Scale(l.GetScale()))
	} else {
		l.SetSize(size.Scale(l.GetScale()))
	}
}

func (l *Living) GetMovementSpeed() float64 { return l.moveSpeedAttr.GetValue() }

func (l *Living) SetMovementSpeed(v float64, fit bool) {
	l.moveSpeedAttr.SetValue(v, fit, false)
}

// SaveNBT is a port of Living::saveNBT.
func (l *Living) SaveNBT() *nbt.CompoundTag {
	tag := l.Entity.SaveNBT()
	tag.SetFloat(tagHealth, nbt.FloatTag(l.GetHealth()))

	tag.SetShort(tagBreathTicks, nbt.ShortTag(l.GetAirSupplyTicks()))

	if effects := l.effectManager.All(); len(effects) > 0 {
		values := make([]nbt.Tag, 0, len(effects))
		for _, e := range effects {
			duration := e.GetDuration()
			if e.IsInfinite() {
				duration = -1
			}
			values = append(values, nbt.NewCompoundTag().
				SetByte(tagEffectID, nbt.ByteTag(bedrock.EffectIdMap().ToID(e.GetType()))).
				SetByte(tagEffectAmplifier, nbt.ByteTag(int8(e.GetAmplifier()))).
				SetInt(tagEffectDuration, nbt.IntTag(duration)).
				SetByte(tagEffectAmbient, boolByte(e.IsAmbient())).
				SetByte(tagEffectShowParticles, boolByte(e.IsVisible())))
		}
		list, err := nbt.NewListTag(values, nbt.TagCompound)
		if err != nil {
			panic(err)
		}
		tag.SetTag(tagActiveEffects, list)
	}

	return tag
}

func (l *Living) GetEffects() *effect.EffectManager { return l.effectManager }

// ConsumeObject is a port of Living::consumeObject: causes the mob to consume the given Consumable
// object, applying applicable effects, health bonuses, food bonuses, etc.
func (l *Living) ConsumeObject(consumable Consumable) bool {
	l.lself.ApplyConsumptionResults(consumable)
	return true
}

// ApplyConsumptionResults applies effects from consuming the object. This shouldn't do any
// can-consume checks (those are expected to be handled by the caller).
func (l *Living) ApplyConsumptionResults(consumable Consumable) {
	for _, e := range consumable.GetAdditionalEffects() {
		l.effectManager.Add(e)
	}
	if _, ok := consumable.(FoodSource); ok {
		l.lself.BroadcastSound(sound.BurpSound{})
	}

	consumable.OnConsume(l.lself)
}

func (l *Living) jumpBoostLevel() int {
	if jumpBoost := l.effectManager.Get(effect.VanillaJumpBoost()); jumpBoost != nil {
		return jumpBoost.GetEffectLevel()
	}
	return 0
}

// GetJumpVelocity returns the initial upwards velocity of a jumping entity in blocks/tick,
// including additional velocity due to effects.
func (l *Living) GetJumpVelocity() float64 {
	return l.jumpVelocity + float64(l.jumpBoostLevel())/10
}

// Jump is called when the entity jumps from the ground. This method adds upwards velocity to the
// entity.
func (l *Living) Jump() {
	if l.OnGround {
		l.motion.Y = l.GetJumpVelocity() //Y motion should already be 0 if we're jumping from the ground.
	}
}

// CalculateFallDamage is a port of Living::calculateFallDamage.
func (l *Living) CalculateFallDamage(fallDistance float64) float64 {
	return stdmath.Ceil(fallDistance - 3 - float64(l.jumpBoostLevel()))
}

// collisionBoxes is the promoted-from-*block.Block surface OnHitGround needs.
type collisionBoxes interface {
	GetCollisionBoxes() []math.AxisAlignedBB
}

// OnHitGround is a port of Living::onHitGround.
func (l *Living) OnHitGround() *float64 {
	w := l.GetWorld()
	fallBlockPos := l.location.Floor()
	fallBlock := w.GetBlockAtIfLoaded(fallBlockPos.FloorX(), fallBlockPos.FloorY(), fallBlockPos.FloorZ())
	if boxes, ok := fallBlock.(collisionBoxes); ok && len(boxes.GetCollisionBoxes()) == 0 {
		fallBlockPos = fallBlockPos.Down(1)
		fallBlock = w.GetBlockAtIfLoaded(fallBlockPos.FloorX(), fallBlockPos.FloorY(), fallBlockPos.FloorZ())
	}
	var newVerticalVelocity *float64
	if v, ok := fallBlock.OnEntityLand(l.lself); ok {
		newVerticalVelocity = &v
	}

	damage := l.lself.CalculateFallDamage(l.FallDistance)
	if damage > 0 {
		ev := entityevent.NewEntityDamageEvent(l.lself, entityevent.CauseFall, damage, nil)
		l.lself.Attack(ev)

		if damage > 4 {
			l.lself.BroadcastSound(sound.EntityLongFallSound{EntityNetworkTypeID: l.self.GetNetworkTypeID(), EntityUniqueID: int64(l.GetID())})
		} else {
			l.lself.BroadcastSound(sound.EntityShortFallSound{EntityNetworkTypeID: l.self.GetNetworkTypeID()})
		}
	} else if fallBlock.GetTypeId() != block.AIR {
		l.lself.BroadcastSound(sound.EntityLandSound{BlockStateID: fallBlock.GetStateId(), EntityNetworkTypeID: l.self.GetNetworkTypeID(), EntityUniqueID: int64(l.GetID())})
	}
	return newVerticalVelocity
}

// GetArmorPoints returns how many armour points this mob has. Armour points provide a percentage
// reduction to damage.
func (l *Living) GetArmorPoints() int {
	total := 0
	for _, it := range l.armorInventory.GetContents(false) {
		total += it.GetDefensePoints()
	}
	return total
}

// GetHighestArmorEnchantmentLevel returns the highest level of the specified enchantment on any
// armour piece that the entity is currently wearing.
func (l *Living) GetHighestArmorEnchantmentLevel(e enchantment.Enchantment) int {
	result := 0
	for _, it := range l.armorInventory.GetContents(false) {
		result = max(result, it.GetEnchantmentLevel(e))
	}
	return result
}

func (l *Living) GetArmorInventory() *inventory.ArmorInventory { return l.armorInventory }

// SetOnFire is a port of Living::setOnFire: Fire Protection shortens the burn.
func (l *Living) SetOnFire(seconds int) {
	reduction := int(min(float64(seconds), float64(seconds)*float64(l.GetHighestArmorEnchantmentLevel(enchantment.VanillaFireProtection()))*0.15))
	l.Entity.SetOnFire(seconds - reduction)
}

// armorItem is the surface of item.Armor (and TurtleHelmet, which embeds it) Living needs - PHP's
// `instanceof Armor`.
type armorItem interface {
	item.Item
	GetEnchantmentProtectionFactor(event entityevent.DamageSource) int
	GetArmorSlot() int
}

// durableItem is the surface of item.Durable Living needs - PHP's `instanceof Durable`.
type durableItem interface {
	item.Item
	ApplyDamage(amount int) bool
	IsBroken() bool
}

// sortedArmorContents returns the armor inventory's non-empty contents in slot order (PHP arrays
// keep the order the slots were filled in, which for a 4-slot inventory is slot order).
func (l *Living) sortedArmorContents() iter.Seq2[int, item.Item] {
	contents := l.armorInventory.GetContents(false)
	slots := make([]int, 0, len(contents))
	for slot := range contents {
		slots = append(slots, slot)
	}
	sort.Ints(slots)
	return func(yield func(int, item.Item) bool) {
		for _, slot := range slots {
			if !yield(slot, contents[slot]) {
				return
			}
		}
	}
}

// ApplyDamageModifiers is a port of Living::applyDamageModifiers: called prior to EntityDamageEvent
// execution to apply modifications to the event's damage, such as reduction due to effects or
// armour.
func (l *Living) ApplyDamageModifiers(source entityevent.DamageSource) {
	if l.lastDamageCause != nil && l.attackTime > 0 {
		if l.lastDamageCause.GetBaseDamage() >= source.GetBaseDamage() {
			source.Cancel()
		}
		source.SetModifier(-l.lastDamageCause.GetBaseDamage(), entityevent.ModifierPreviousDamageCooldown)
	}
	if source.CanBeReducedByArmor() {
		//MCPE uses the same system as PC did pre-1.9
		source.SetModifier(-source.GetFinalDamage()*float64(l.GetArmorPoints())*0.04, entityevent.ModifierArmor)
	}

	cause := source.GetCause()
	if resistance := l.effectManager.Get(effect.VanillaResistance()); resistance != nil && cause != entityevent.CauseVoid && cause != entityevent.CauseSuicide {
		source.SetModifier(-source.GetFinalDamage()*min(1, 0.2*float64(resistance.GetEffectLevel())), entityevent.ModifierResistance)
	}

	totalEpf := 0
	for _, it := range l.sortedArmorContents() {
		if armor, ok := it.(armorItem); ok {
			totalEpf += armor.GetEnchantmentProtectionFactor(source)
		}
	}
	randomFactor := float64(50+rand.IntN(51)) / 100
	source.SetModifier(-source.GetFinalDamage()*min(stdmath.Ceil(float64(min(totalEpf, 25))*randomFactor), 20)*0.04, entityevent.ModifierArmorEnchantments)

	source.SetModifier(-min(l.GetAbsorption(), source.GetFinalDamage()), entityevent.ModifierAbsorption)

	if _, isArmor := l.armorInventory.GetHelmet().(armorItem); cause == entityevent.CauseFallingBlock && isArmor {
		source.SetModifier(-(source.GetFinalDamage() / 4), entityevent.ModifierArmorHelmet)
	}
}

// ApplyPostDamageEffects is a port of Living::applyPostDamageEffects: called after
// EntityDamageEvent execution to apply post-hurt effects, such as reducing absorption or modifying
// armour durability. This will not be called by damage sources causing death.
func (l *Living) ApplyPostDamageEffects(source entityevent.DamageSource) {
	l.SetAbsorption(max(0, l.GetAbsorption()+source.GetModifier(entityevent.ModifierAbsorption)))
	if source.CanBeReducedByArmor() {
		l.DamageArmor(source.GetBaseDamage())
	}

	if byEntity, ok := AsDamageByEntity(source); ok {
		if attacker := byEntity.GetDamager(); attacker != nil {
			damage := 0
			for k, it := range l.sortedArmorContents() {
				armor, ok := it.(armorItem)
				if !ok {
					continue
				}
				if thornsLevel := armor.GetEnchantmentLevel(enchantment.VanillaThorns()); thornsLevel > 0 {
					if durable, ok := it.(durableItem); ok {
						if rand.IntN(100) < thornsLevel*15 {
							l.damageItem(durable, 3)
							if thornsLevel > 10 {
								damage += thornsLevel - 10
							} else {
								damage += 1 + rand.IntN(4)
							}
						} else {
							l.damageItem(durable, 1) //thorns causes an extra +1 durability loss even if it didn't activate
						}
					}

					l.armorInventory.SetItem(k, it)
				}
			}

			if damage > 0 {
				if victim, ok := attacker.(world.Entity); ok {
					victim.Attack(entityevent.NewEntityDamageByEntityEvent(l.lself, attacker, entityevent.CauseMagic, float64(damage), nil))
				}
			}

			if source.GetModifier(entityevent.ModifierArmorHelmet) < 0 {
				helmet := l.armorInventory.GetHelmet()
				if durable, ok := helmet.(durableItem); ok {
					if _, isArmor := helmet.(armorItem); isArmor {
						finalDamage := source.GetFinalDamage()
						l.damageItem(durable, int(stdmath.Round(finalDamage*4+utils.GetRandomFloat()*finalDamage*2)))
						l.armorInventory.SetHelmet(helmet)
					}
				}
			}
		}
	}
}

// AsDamageByEntity narrows a damage source to an EntityDamageByEntityEvent (including the
// ByChildEntity subclass) - PHP's `instanceof EntityDamageByEntityEvent`.
func AsDamageByEntity(source entityevent.DamageSource) (*entityevent.EntityDamageByEntityEvent, bool) {
	switch s := source.(type) {
	case *entityevent.EntityDamageByEntityEvent:
		return s, true
	case *entityevent.EntityDamageByChildEntityEvent:
		return &s.EntityDamageByEntityEvent, true
	}
	return nil, false
}

// DamageArmor damages the worn armour according to the amount of damage given. Each 4 points
// (rounded down) deals 1 damage point to each armour piece, but never less than 1 total.
func (l *Living) DamageArmor(damage float64) {
	durabilityRemoved := int(max(stdmath.Floor(damage/4), 1))

	for slotID, it := range l.sortedArmorContents() {
		if _, isArmor := it.(armorItem); !isArmor {
			continue
		}
		durable, ok := it.(durableItem)
		if !ok {
			continue
		}
		oldItem := it.Clone()
		l.damageItem(durable, durabilityRemoved)
		if !it.EqualsExact(oldItem) {
			l.armorInventory.SetItem(slotID, it)
		}
	}
}

func (l *Living) damageItem(it durableItem, durabilityRemoved int) {
	it.ApplyDamage(durabilityRemoved)
	if it.IsBroken() {
		l.lself.BroadcastSound(sound.ItemBreakSound{})
	}
}

// Attack is a port of Living::attack.
func (l *Living) Attack(source entityevent.DamageSource) {
	if l.NoDamageTicks > 0 && source.GetCause() != entityevent.CauseSuicide {
		source.Cancel()
	}

	if l.effectManager.Has(effect.VanillaFireResistance()) {
		switch source.GetCause() {
		case entityevent.CauseFire, entityevent.CauseFireTick, entityevent.CauseLava:
			source.Cancel()
		}
	}

	if source.GetCause() != entityevent.CauseSuicide {
		l.lself.ApplyDamageModifiers(source)
	}

	if byEntity, ok := AsDamageByEntity(source); ok &&
		(source.GetCause() == entityevent.CauseBlockExplosion || source.GetCause() == entityevent.CauseEntityExplosion) {
		//TODO: knockback should not just apply for entity damage sources
		//this doesn't matter for TNT right now because the PrimedTNT entity is considered the source, not the block.
		base := byEntity.GetKnockBack()
		byEntity.SetKnockBack(base - min(base, base*float64(l.GetHighestArmorEnchantmentLevel(enchantment.VanillaBlastProtection()))*0.15))
	}

	l.Entity.Attack(source)

	if source.IsCancelled() {
		return
	}

	if l.attackTime <= 0 {
		//this logic only applies if the entity was cold attacked

		l.attackTime = source.GetAttackCooldown()

		switch s := source.(type) {
		case *entityevent.EntityDamageByChildEntityEvent:
			if child := s.GetChild(); child != nil {
				if mover, ok := child.(interface{ GetMotion() math.Vector3 }); ok {
					motion := mover.GetMotion()
					l.lself.KnockBack(motion.X, motion.Z, s.GetKnockBack(), s.GetVerticalKnockBackLimit())
				}
			}
		case *entityevent.EntityDamageByEntityEvent:
			if e := s.GetDamager(); e != nil {
				damagerPos := e.GetPosition()
				deltaX := l.location.X - damagerPos.X
				deltaZ := l.location.Z - damagerPos.Z
				l.lself.KnockBack(deltaX, deltaZ, s.GetKnockBack(), s.GetVerticalKnockBackLimit())
			}
		}

		if l.IsAlive() {
			l.lself.DoHitAnimation()
		}
	}

	if l.IsAlive() {
		l.lself.ApplyPostDamageEffects(source)
	}
}

// DoHitAnimation is a port of Living::doHitAnimation.
func (l *Living) DoHitAnimation() {
	l.lself.BroadcastAnimation(animation.HurtAnimation{Entity: l}, nil)
}

// KnockBack is a port of Living::knockBack. PHP's nullable $verticalLimit (null meaning "use
// force") has no Go equivalent; pass force explicitly for that behaviour.
func (l *Living) KnockBack(x, z, force, verticalLimit float64) {
	f := stdmath.Sqrt(x*x + z*z)
	if f <= 0 {
		return
	}
	if rand.Float64() > l.knockbackResistanceAttr.GetValue() {
		f = 1 / f

		motionX := l.motion.X / 2
		motionY := l.motion.Y / 2
		motionZ := l.motion.Z / 2
		motionX += x * f * force
		motionY += force
		motionZ += z * f * force

		if motionY > verticalLimit {
			motionY = verticalLimit
		}

		l.lself.SetMotion(math.NewVector3(motionX, motionY, motionZ))
	}
}

// OnDeath is a port of Living::onDeath: fires EntityDeathEvent, drops the (possibly modified) drops
// and experience, and starts the death animation.
func (l *Living) OnDeath() {
	drops := l.lself.GetDrops()
	eventDrops := make([]entityevent.Item, len(drops))
	for i, d := range drops {
		eventDrops[i] = d
	}
	ev := entityevent.NewEntityDeathEvent(l.lself, eventDrops, l.lself.GetXpDropAmount())
	ev.Call()
	for _, d := range ev.GetDrops() {
		if it, ok := d.(item.Item); ok {
			l.GetWorld().DropItem(l.location.Vector3, it, nil, 10)
		}
	}

	//TODO: check death conditions (must have been damaged by player < 5 seconds from death)
	l.GetWorld().DropExperience(l.location.Vector3, ev.GetXpDropAmount())

	l.lself.StartDeathAnimation()
}

// OnDeathUpdate is a port of Living::onDeathUpdate.
func (l *Living) OnDeathUpdate(tickDiff int) bool {
	if l.DeadTicks < l.maxDeadTicks {
		l.DeadTicks += tickDiff
		if l.DeadTicks >= l.maxDeadTicks {
			l.lself.EndDeathAnimation()
		}
	}

	return l.DeadTicks >= l.maxDeadTicks
}

func (l *Living) StartDeathAnimation() {
	l.lself.BroadcastAnimation(animation.DeathAnimation{Entity: l}, nil)
}

func (l *Living) EndDeathAnimation() { l.DespawnFromAll() }

// EntityBaseTick is a port of Living::entityBaseTick.
func (l *Living) EntityBaseTick(tickDiff int) bool {
	hasUpdate := l.Entity.EntityBaseTick(tickDiff)

	if l.IsAlive() {
		if l.effectManager.Tick(tickDiff) {
			hasUpdate = true
		}

		if l.IsInsideOfSolid() {
			hasUpdate = true
			ev := entityevent.NewEntityDamageEvent(l.lself, entityevent.CauseSuffocation, 1, nil)
			l.lself.Attack(ev)
		}

		if l.DoAirSupplyTick(tickDiff) {
			hasUpdate = true
		}

		for index, it := range l.sortedArmorContents() {
			oldItem := it.Clone()
			if it.OnTickWorn(l.lself) {
				hasUpdate = true
				if !it.EqualsExact(oldItem) {
					l.armorInventory.SetItem(index, it)
				}
			}
		}
	}

	if l.attackTime > 0 {
		l.attackTime -= tickDiff
	}

	return hasUpdate
}

// Move is a port of Living::move (Frost Walker on top of Entity.Move).
func (l *Living) Move(dx, dy, dz float64) {
	oldX := l.location.X
	oldZ := l.location.Z

	l.Entity.Move(dx, dy, dz)

	frostWalkerLevel := l.GetFrostWalkerLevel()
	if frostWalkerLevel > 0 && (stdmath.Abs(l.location.X-oldX) > MotionThreshold || stdmath.Abs(l.location.Z-oldZ) > MotionThreshold) {
		l.ApplyFrostWalker(frostWalkerLevel)
	}
}

// ApplyFrostWalker is a port of Living::applyFrostWalker.
func (l *Living) ApplyFrostWalker(level int) {
	radius := level + 2
	w := l.GetWorld()

	baseX := l.location.FloorX()
	y := l.location.FloorY() - 1
	baseZ := l.location.FloorZ()

	liquid := block.VanillaWater()
	targetBlock := block.VanillaFrostedIce()
	if entityevent.HasHandlers[entityevent.EntityFrostWalkerEvent]() {
		ev := entityevent.NewEntityFrostWalkerEvent(l.lself, radius, liquid, targetBlock)
		ev.Call()
		if ev.IsCancelled() {
			return
		}
		radius = ev.GetRadius()
		if b, ok := ev.GetLiquid().(block.Behavior); ok {
			liquid = b
		}
		if b, ok := ev.GetTargetBlock().(block.Behavior); ok {
			targetBlock = b
		}
	}

	for x := baseX - radius; x <= baseX+radius; x++ {
		for z := baseZ - radius; z <= baseZ+radius; z++ {
			blk := w.GetBlockAtIfLoaded(x, y, z)
			if blk.GetStateId() != liquid.GetStateId() ||
				w.GetBlockAtIfLoaded(x, y+1, z).GetTypeId() != block.AIR ||
				len(w.GetNearbyEntitiesExcept(math.OneAABB().OffsetCopy(float64(x), float64(y), float64(z)), nil)) != 0 {
				continue
			}
			_ = w.SetBlockAt(x, y, z, targetBlock)
		}
	}
}

// GetFrostWalkerLevel is a port of Living::getFrostWalkerLevel (cached until the boots change).
func (l *Living) GetFrostWalkerLevel() int {
	if l.frostWalkerLevel == nil {
		level := l.armorInventory.GetBoots().GetEnchantmentLevel(enchantment.VanillaFrostWalker())
		l.frostWalkerLevel = &level
	}
	return *l.frostWalkerLevel
}

// DoAirSupplyTick is a port of Living::doAirSupplyTick: consumes the entity's air supply when
// underwater and regenerates it when out of water.
func (l *Living) DoAirSupplyTick(tickDiff int) bool {
	ticks := l.GetAirSupplyTicks()
	oldTicks := ticks
	if !l.lself.CanBreathe() {
		l.SetBreathing(false)

		respirationLevel := l.armorInventory.GetHelmet().GetEnchantmentLevel(enchantment.VanillaRespiration())
		if respirationLevel <= 0 || utils.GetRandomFloat() <= 1/float64(respirationLevel+1) {
			ticks -= tickDiff
			if ticks <= -20 {
				ticks = 0
				l.lself.OnAirExpired()
			}
		}
	} else if !l.IsBreathing() {
		maxTicks := l.GetMaxAirSupplyTicks()
		if ticks < maxTicks {
			ticks += tickDiff * 5
		}
		if ticks >= maxTicks {
			ticks = maxTicks
			l.SetBreathing(true)
		}
	}

	if ticks != oldTicks {
		l.SetAirSupplyTicks(ticks)
	}

	return ticks != oldTicks
}

// CanBreathe returns whether the entity can currently breathe.
func (l *Living) CanBreathe() bool {
	return l.effectManager.Has(effect.VanillaWaterBreathing()) || l.effectManager.Has(effect.VanillaConduitPower()) || !l.lself.IsUnderwater()
}

// IsBreathing returns whether the entity is currently breathing or not. If this is false, the
// entity's air supply will be used.
func (l *Living) IsBreathing() bool { return l.breathing }

// SetBreathing sets whether the entity is currently breathing. If false, it will cause the entity's
// air supply to be used. For players, this also shows the oxygen bar.
func (l *Living) SetBreathing(value bool) {
	l.breathing = value
	l.networkPropertiesDirty = true
}

// GetAirSupplyTicks returns the number of ticks remaining in the entity's air supply.
func (l *Living) GetAirSupplyTicks() int { return l.breathTicks }

func (l *Living) SetAirSupplyTicks(ticks int) {
	l.breathTicks = ticks
	l.networkPropertiesDirty = true
}

func (l *Living) GetMaxAirSupplyTicks() int { return l.maxBreathTicks }

func (l *Living) SetMaxAirSupplyTicks(ticks int) {
	l.maxBreathTicks = ticks
	l.networkPropertiesDirty = true
}

// OnAirExpired is called when the entity's air supply ticks reaches -20 or lower. The entity will
// usually take damage at this point and then the supply is reset to 0, so this method will be
// called roughly every second.
func (l *Living) OnAirExpired() {
	ev := entityevent.NewEntityDamageEvent(l.lself, entityevent.CauseDrowning, 2, nil)
	l.lself.Attack(ev)
}

// GetDrops is Living::getDrops' default: nothing.
func (l *Living) GetDrops() []item.Item { return nil }

// GetXpDropAmount returns the amount of XP this mob will drop on death.
func (l *Living) GetXpDropAmount() int { return 0 }

// GetLineOfSight is a port of Living::getLineOfSight. transparent is the set of block type IDs to
// look through (nil/empty means only air).
func (l *Living) GetLineOfSight(maxDistance, maxLength int, transparent map[int]bool) []block.Behavior {
	if maxDistance > 120 {
		maxDistance = 120
	}

	if len(transparent) == 0 {
		transparent = nil
	}

	var blocks []block.Behavior

	seq, err := math.InDirection(l.location.Add(0, l.Size.GetEyeHeight(), 0), l.GetDirectionVector(), float64(maxDistance))
	if err != nil {
		return nil
	}
	w := l.GetWorld()
	for v := range seq {
		blk := w.GetBlockAtIfLoaded(int(v.X), int(v.Y), int(v.Z))
		blocks = append(blocks, blk)

		if maxLength != 0 && len(blocks) > maxLength {
			blocks = blocks[1:]
		}

		id := blk.GetTypeId()

		if transparent == nil {
			if id != block.AIR {
				break
			}
		} else if !transparent[id] {
			break
		}
	}

	return blocks
}

// GetTargetBlock is a port of Living::getTargetBlock (nil if none).
func (l *Living) GetTargetBlock(maxDistance int, transparent map[int]bool) block.Behavior {
	line := l.GetLineOfSight(maxDistance, 1, transparent)
	if len(line) > 0 {
		return line[0]
	}
	return nil
}

// LookAt changes the entity's yaw and pitch to make it look at the specified position. For mobs,
// this will cause their heads to turn.
func (l *Living) LookAt(target math.Vector3) {
	xDist := target.X - l.location.X
	zDist := target.Z - l.location.Z

	horizontal := stdmath.Sqrt(xDist*xDist + zDist*zDist)
	vertical := target.Y - (l.location.Y + l.self.GetEyeHeight())
	pitch := -stdmath.Atan2(vertical, horizontal) / stdmath.Pi * 180 //negative is up, positive is down

	yaw := stdmath.Atan2(zDist, xDist)/stdmath.Pi*180 - 90
	if yaw < 0 {
		yaw += 360.0
	}

	l.SetRotation(yaw, pitch)
}

// SendSpawnPacket is a port of Living::sendSpawnPacket.
func (l *Living) SendSpawnPacket(player world.EntityViewer) {
	l.Entity.SendSpawnPacket(player)

	player.SendPacket(mobArmorChangePacket(l))
}

// SyncNetworkData is a port of Living::syncNetworkData.
func (l *Living) SyncNetworkData(properties *MetadataCollection) {
	l.Entity.SyncNetworkData(properties)

	visibleEffects := map[int]bool{}
	for _, e := range l.effectManager.All() {
		if !e.IsVisible() || !e.GetType().HasBubbles() {
			continue
		}
		visibleEffects[bedrock.EffectIdMap().ToID(e.GetType())] = e.IsAmbient()
	}

	//TODO: HACK! the client may not be able to identify effects if they are not sorted.
	effectIDs := make([]int, 0, len(visibleEffects))
	for id := range visibleEffects {
		effectIDs = append(effectIDs, id)
	}
	sort.Ints(effectIDs)

	var effectsData int64
	packedEffectsCount := 0
	for _, effectID := range effectIDs {
		ambient := int64(0)
		if visibleEffects[effectID] {
			ambient = 1
		}
		effectsData = (effectsData << 7) |
			(int64(effectID&0x3f) << 1) | //Why not use 7 bits instead of only 6? mojang...
			ambient

		packedEffectsCount++
		if packedEffectsCount >= 8 {
			break
		}
	}
	properties.SetLong(MetadataVisibleMobEffects, effectsData)

	properties.SetShort(MetadataAir, int16(l.breathTicks))
	properties.SetShort(MetadataMaxAir, int16(l.maxBreathTicks))

	properties.SetGenericFlag(FlagBreathing, l.breathing)
	properties.SetGenericFlag(FlagSneaking, l.sneaking)
	properties.SetGenericFlag(FlagSprinting, l.sprinting)
	properties.SetGenericFlag(FlagGliding, l.gliding)
	properties.SetGenericFlag(FlagSwimming, l.swimming)
}

// OnDispose is a port of Living::onDispose.
func (l *Living) OnDispose() {
	l.armorInventory.RemoveAllViewers()
	l.effectManager.GetEffectAddHooks().Clear()
	l.effectManager.GetEffectRemoveHooks().Clear()
	l.Entity.OnDispose()
}
