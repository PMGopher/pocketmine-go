package entity

import (
	"fmt"
	"strconv"

	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

	"pocketmine-go/pocketmine/data"
	"pocketmine-go/pocketmine/entity/animation"
	"pocketmine-go/pocketmine/entity/effect"
	entityevent "pocketmine-go/pocketmine/event/entity"
	playerevent "pocketmine-go/pocketmine/event/player"
	"pocketmine-go/pocketmine/inventory"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/item/enchantment"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
	"pocketmine-go/pocketmine/network/mcpe/convert"
	"pocketmine-go/pocketmine/world"
	"pocketmine-go/pocketmine/world/sound"
)

// Human NBT keys, a port of Human's TAG_* constants.
const (
	tagInventory             = "Inventory"             //TAG_List<TAG_Compound>
	tagOffHandItem           = "OffHandItem"           //TAG_Compound
	tagEnderChestInventory   = "EnderChestInventory"   //TAG_List<TAG_Compound>
	tagSelectedInventorySlot = "SelectedInventorySlot" //TAG_Int
	tagFoodLevel             = "foodLevel"             //TAG_Int
	tagFoodExhaustionLevel   = "foodExhaustionLevel"   //TAG_Float
	tagFoodSaturationLevel   = "foodSaturationLevel"   //TAG_Float
	tagFoodTickTimer         = "foodTickTimer"         //TAG_Int
	tagXpLevel               = "XpLevel"               //TAG_Int
	tagXpProgress            = "XpP"                   //TAG_Float
	tagLifetimeXpTotal       = "XpTotal"               //TAG_Int
	tagXpSeed                = "XpSeed"                //TAG_Int
	tagSkin                  = "Skin"                  //TAG_Compound
	tagSkinName              = "Name"                  //TAG_String
	tagSkinData              = "Data"                  //TAG_ByteArray
	tagSkinCapeData          = "CapeData"              //TAG_ByteArray
	tagSkinGeometryName      = "GeometryName"          //TAG_String
	tagSkinGeometryData      = "GeometryData"          //TAG_ByteArray
)

// HumanHooks extends LivingHooks with the Human hooks Human's own bodies call on $this.
type HumanHooks interface {
	LivingHooks
	InitHumanData(tag *nbt.CompoundTag)
	CanEat() bool
}

// Human is a port of pocketmine\entity\Human (also PHP's ProjectileSource and InventoryHolder).
type Human struct {
	Living

	hself HumanHooks

	inventory        *inventory.PlayerInventory
	offHandInventory *inventory.PlayerOffHandInventory
	enderInventory   *inventory.PlayerEnderInventory

	uuid uuid.UUID
	skin *Skin

	hungerManager *HungerManager
	xpManager     *ExperienceManager

	xpSeed int
}

// NewHuman is a port of Human::__construct for a plain Human entity (an NPC).
func NewHuman(location Location, skin *Skin, tag *nbt.CompoundTag) *Human {
	h := &Human{}
	h.ConstructHuman(h, location, skin, tag)
	return h
}

// ConstructHuman is a port of Human::__construct for subtypes (e.g. player.Player): set the skin,
// then run Living's construction.
func (h *Human) ConstructHuman(self HumanHooks, location Location, skin *Skin, tag *nbt.CompoundTag) {
	h.hself = self
	h.skin = skin
	h.ConstructLiving(self, location, tag)
}

func (h *Human) GetNetworkTypeID() string { return EntityIDPlayer }

func (h *Human) GetInitialSizeInfo() EntitySizeInfo {
	return NewEntitySizeInfoWithEyeHeight(1.8, 0.6, 1.62)
}

// ParseSkinNBT is a port of Human::parseSkinNBT.
func ParseSkinNBT(tag *nbt.CompoundTag) (*Skin, error) {
	skinTag, ok, _ := tag.GetCompoundTag(tagSkin)
	if !ok {
		return nil, data.NewSavedDataLoadingError("Missing skin data")
	}
	var skinData []byte
	if raw, ok := skinTag.GetTag(tagSkinData); ok {
		switch v := raw.(type) {
		case nbt.StringTag: //old data (this used to be saved as a StringTag in older versions of PM)
			skinData = []byte(v)
		case nbt.ByteArrayTag:
			skinData = []byte(v)
		}
	}
	return NewSkin( //this errors if the skin is invalid
		string(skinTag.GetStringOr(tagSkinName, "")),
		skinData,
		[]byte(skinTag.GetByteArrayOr(tagSkinCapeData, nil)),
		string(skinTag.GetStringOr(tagSkinGeometryName, "")),
		[]byte(skinTag.GetByteArrayOr(tagSkinGeometryData, nil)),
	)
}

func (h *Human) GetUniqueID() uuid.UUID { return h.uuid }

// SetUniqueID sets the UUID (PHP subclasses assign $this->uuid directly in initHumanData).
func (h *Human) SetUniqueID(id uuid.UUID) { h.uuid = id }

// GetSkin returns a Skin object containing information about this human's skin.
func (h *Human) GetSkin() *Skin { return h.skin }

// SetSkin sets the human's skin. This will not send any update to viewers, you need to do that
// manually using SendSkin.
func (h *Human) SetSkin(skin *Skin) { h.skin = skin }

// SendSkin is a port of Human::sendSkin (nil targets means every viewer).
func (h *Human) SendSkin(targets []world.EntityViewer) {
	if targets == nil {
		targets = h.hasSpawnedOrder
	}
	BroadcastPackets(targets, &packet.PlayerSkin{UUID: h.GetUniqueID(), Skin: SkinToNetwork(h.skin)})
}

// Jump is a port of Human::jump (jumping costs food).
func (h *Human) Jump() {
	h.Living.Jump()
	if h.hself.IsSprinting() {
		h.hungerManager.Exhaust(0.2, playerevent.ExhaustCauseSprintJumping)
	} else {
		h.hungerManager.Exhaust(0.05, playerevent.ExhaustCauseJumping)
	}
}

// Emote is a port of Human::emote.
func (h *Human) Emote(emoteID string) {
	BroadcastPackets(h.GetViewers(), &packet.Emote{
		EntityRuntimeID: uint64(h.GetID()),
		EmoteLength:     0, //seems to be irrelevant for the client, we cannot risk rebroadcasting random values received
		EmoteID:         emoteID,
		XUID:            "",
		PlatformID:      "",
		Flags:           packet.EmoteFlagServerSide | packet.EmoteFlagMuteChat,
	})
}

func (h *Human) GetHungerManager() *HungerManager { return h.hungerManager }

// CanEat returns whether the Human can eat food. This may return a different result than
// HungerManager.IsHungry, as HungerManager only handles the hunger bar.
func (h *Human) CanEat() bool {
	return h.hungerManager.IsHungry() || h.GetWorld().GetDifficulty() == world.DifficultyPeaceful
}

// ConsumeObject is a port of Human::consumeObject.
func (h *Human) ConsumeObject(consumable Consumable) bool {
	if food, ok := consumable.(FoodSource); ok && food.RequiresHunger() && !h.hself.CanEat() {
		return false
	}

	return h.Living.ConsumeObject(consumable)
}

// ApplyConsumptionResults is a port of Human::applyConsumptionResults.
func (h *Human) ApplyConsumptionResults(consumable Consumable) {
	if food, ok := consumable.(FoodSource); ok {
		h.hungerManager.AddFood(float64(food.GetFoodRestore()))
		h.hungerManager.AddSaturation(food.GetSaturationRestore())
	}

	h.Living.ApplyConsumptionResults(consumable)
}

func (h *Human) GetXpManager() *ExperienceManager { return h.xpManager }

func (h *Human) GetEnchantmentSeed() int { return h.xpSeed }

func (h *Human) SetEnchantmentSeed(seed int) { h.xpSeed = seed }

func (h *Human) RegenerateEnchantmentSeed() { h.xpSeed = enchantment.GenerateSeed() }

// GetXpDropAmount is a port of Human::getXpDropAmount: this causes some XP to be lost on death when
// above level 1 (by design), dropping at most enough points for about 7.5 levels of XP.
func (h *Human) GetXpDropAmount() int {
	return min(100, 7*h.xpManager.GetXpLevel())
}

func (h *Human) GetInventory() *inventory.PlayerInventory { return h.inventory }

func (h *Human) GetOffHandInventory() *inventory.PlayerOffHandInventory {
	return h.offHandInventory
}

func (h *Human) GetEnderInventory() *inventory.PlayerEnderInventory { return h.enderInventory }

// GetSneakOffset is a port of Human::getSneakOffset.
func (h *Human) GetSneakOffset() float64 { return 0.31 }

// InitHumanData is a port of Human::initHumanData: for Human entities which are not players, sets
// their UUID from their ID, skin and name tag.
func (h *Human) InitHumanData(tag *nbt.CompoundTag) {
	//TODO: use of NIL UUID for namespace is a hack; we should provide a proper UUID for the namespace
	h.uuid = uuid.NewMD5(uuid.Nil, []byte(strconv.Itoa(h.GetID())+string(h.skin.GetSkinData())+h.GetNameTag()))
}

// InitEntity is a port of Human::initEntity.
func (h *Human) InitEntity(tag *nbt.CompoundTag) {
	h.Living.InitEntity(tag)

	h.hungerManager = NewHungerManager(h)
	h.xpManager = NewExperienceManager(h)

	h.inventory = inventory.NewPlayerInventory(h.hself)
	syncHeldItem := func() {
		BroadcastPackets(h.GetViewers(), mobMainHandItemChangePacket(h))
	}
	h.inventory.GetListeners().Add(inventory.NewCallbackInventoryListener(
		func(_ inventory.Inventory, slot int, _ item.Item) {
			if slot == h.inventory.GetHeldItemIndex() {
				syncHeldItem()
			}
		},
		func(_ inventory.Inventory, oldItems map[int]item.Item) {
			if _, ok := oldItems[h.inventory.GetHeldItemIndex()]; ok {
				syncHeldItem()
			}
		},
	))
	h.offHandInventory = inventory.NewPlayerOffHandInventory(h.hself)
	h.enderInventory = inventory.NewPlayerEnderInventory(h.hself, 27)
	h.hself.InitHumanData(tag)

	if inventoryTag, ok, _ := tag.GetListTag(tagInventory); ok {
		inventoryItems := map[int]item.Item{}
		armorInventoryItems := map[int]item.Item{}
		for _, t := range inventoryTag.Values() {
			itemTag, ok := t.(*nbt.CompoundTag)
			if !ok {
				continue
			}
			slot := int(itemTag.GetByteOr("Slot", 0))
			switch {
			case slot >= 0 && slot < 9: // Hotbar
				// Old hotbar saving stuff, ignore it
			case slot >= 100 && slot < 104: // Armor
				armorSlot := slot - 100
				armorInventoryItems[armorSlot] = item.SafeNbtDeserialize(itemTag, fmt.Sprintf("Human armor slot %d", armorSlot), nil)
			case slot >= 9 && slot < h.inventory.GetSize()+9:
				inventorySlot := slot - 9
				inventoryItems[inventorySlot] = item.SafeNbtDeserialize(itemTag, fmt.Sprintf("Human inventory slot %d", inventorySlot), nil)
			}
		}
		populateInventoryFromListTag(h.inventory, inventoryItems)
		populateInventoryFromListTag(h.armorInventory, armorInventoryItems)
	}

	if offHand, ok, _ := tag.GetCompoundTag(tagOffHandItem); ok {
		h.offHandInventory.SetItem(0, item.SafeNbtDeserialize(offHand, "Human off-hand item", nil))
	}

	h.offHandInventory.GetListeners().Add(inventory.OnAnyChange(func(inventory.Inventory) {
		BroadcastPackets(h.GetViewers(), mobOffHandItemChangePacket(h))
	}))

	// PHP throws for an out-of-range saved slot; falling back to the first slot keeps a corrupted
	// value from making the whole entity unloadable.
	selectedSlot := int(tag.GetIntOr(tagSelectedInventorySlot, 0))
	if !h.inventory.IsHotbarSlot(selectedSlot) {
		selectedSlot = 0
	}
	h.inventory.SetHeldItemIndex(selectedSlot)
	if enderTag, ok, _ := tag.GetListTag(tagEnderChestInventory); ok {
		enderChestInventoryItems := map[int]item.Item{}
		for _, t := range enderTag.Values() {
			itemTag, ok := t.(*nbt.CompoundTag)
			if !ok {
				continue
			}
			slot := int(itemTag.GetByteOr("Slot", 0))
			enderChestInventoryItems[slot] = item.SafeNbtDeserialize(itemTag, fmt.Sprintf("Human ender chest slot %d", slot), nil)
		}
		populateInventoryFromListTag(h.enderInventory, enderChestInventoryItems)
	}

	onHeldItemIndexChange := inventory.HeldItemIndexChangeListener(func(int) { syncHeldItem() })
	h.inventory.GetHeldItemIndexChangeListeners().Add(&onHeldItemIndexChange)

	h.hungerManager.SetFood(float64(tag.GetIntOr(tagFoodLevel, nbt.IntTag(h.hungerManager.GetFood()))))
	h.hungerManager.SetExhaustion(float64(tag.GetFloatOr(tagFoodExhaustionLevel, nbt.FloatTag(h.hungerManager.GetExhaustion()))))
	h.hungerManager.SetSaturation(float64(tag.GetFloatOr(tagFoodSaturationLevel, nbt.FloatTag(h.hungerManager.GetSaturation()))))
	h.hungerManager.SetFoodTickTimer(int(tag.GetIntOr(tagFoodTickTimer, nbt.IntTag(h.hungerManager.GetFoodTickTimer()))))

	h.xpManager.SetXpAndProgressNoEvent(
		int(tag.GetIntOr(tagXpLevel, 0)),
		float64(tag.GetFloatOr(tagXpProgress, 0.0)))
	h.xpManager.SetLifetimeTotalXp(int(tag.GetIntOr(tagLifetimeXpTotal, 0)))

	if seed, ok := tag.GetTag(tagXpSeed); ok {
		if intTag, ok := seed.(nbt.IntTag); ok {
			h.xpSeed = int(intTag)
		} else {
			h.xpSeed = enchantment.GenerateSeed()
		}
	} else {
		h.xpSeed = enchantment.GenerateSeed()
	}
}

// EntityBaseTick is a port of Human::entityBaseTick.
func (h *Human) EntityBaseTick(tickDiff int) bool {
	hasUpdate := h.Living.EntityBaseTick(tickDiff)

	h.hungerManager.Tick(tickDiff)
	h.xpManager.Tick(tickDiff)

	return hasUpdate
}

// GetName is a port of Human::getName: the name tag.
func (h *Human) GetName() string { return h.GetNameTag() }

// ApplyDamageModifiers is a port of Human::applyDamageModifiers: a held totem prevents death.
func (h *Human) ApplyDamageModifiers(source entityevent.DamageSource) {
	h.Living.ApplyDamageModifiers(source)

	cause := source.GetCause()
	if cause != entityevent.CauseSuicide && cause != entityevent.CauseVoid && (isTotem(h.inventory.GetItemInHand()) || isTotem(h.offHandInventory.GetItem(0))) {
		compensation := h.hself.GetHealth() - source.GetFinalDamage() - 1
		if compensation <= -1 {
			source.SetModifier(compensation, entityevent.ModifierTotem)
		}
	}
}

func isTotem(it item.Item) bool {
	_, ok := it.(*item.Totem)
	return ok
}

// ApplyPostDamageEffects is a port of Human::applyPostDamageEffects: a totem that prevented death
// grants its effects and is used up.
func (h *Human) ApplyPostDamageEffects(source entityevent.DamageSource) {
	h.Living.ApplyPostDamageEffects(source)
	totemModifier := source.GetModifier(entityevent.ModifierTotem)
	if totemModifier < 0 { //Totem prevented death
		h.effectManager.Clear()

		h.effectManager.Add(effect.NewEffectInstanceWith(effect.VanillaRegeneration(), 40*20, 1))
		h.effectManager.Add(effect.NewEffectInstanceWith(effect.VanillaFireResistance(), 40*20, 1))
		h.effectManager.Add(effect.NewEffectInstanceWith(effect.VanillaAbsorption(), 5*20, 1))

		h.hself.BroadcastAnimation(animation.TotemUseAnimation{Human: h}, nil)
		h.hself.BroadcastSound(sound.TotemUseSound{})

		hand := h.inventory.GetItemInHand()
		if isTotem(hand) {
			hand.Pop() //Plugins could alter max stack size
			h.inventory.SetItemInHand(hand)
		} else if offHand := h.offHandInventory.GetItem(0); isTotem(offHand) {
			offHand.Pop()
			h.offHandInventory.SetItem(0, offHand)
		}
	}
}

// GetDrops is a port of Human::getDrops: everything carried, except Curse of Vanishing items and
// items kept on death.
func (h *Human) GetDrops() []item.Item {
	var drops []item.Item
	for _, contents := range []map[int]item.Item{
		h.inventory.GetContents(false),
		h.armorInventory.GetContents(false),
		h.offHandInventory.GetContents(false),
	} {
		for _, it := range orderedContents(contents) {
			if !it.HasEnchantment(enchantment.VanillaVanishing(), -1) && !it.KeepOnDeath() {
				drops = append(drops, it)
			}
		}
	}
	return drops
}

func orderedContents(contents map[int]item.Item) []item.Item {
	maxSlot := -1
	for slot := range contents {
		maxSlot = max(maxSlot, slot)
	}
	result := make([]item.Item, 0, len(contents))
	for slot := 0; slot <= maxSlot; slot++ {
		if it, ok := contents[slot]; ok {
			result = append(result, it)
		}
	}
	return result
}

// SaveNBT is a port of Human::saveNBT (see Human's doc comment for the inventory contents gap).
func (h *Human) SaveNBT() *nbt.CompoundTag {
	tag := h.Living.SaveNBT()

	tag.SetInt(tagFoodLevel, nbt.IntTag(h.hungerManager.GetFood()))
	tag.SetFloat(tagFoodExhaustionLevel, nbt.FloatTag(h.hungerManager.GetExhaustion()))
	tag.SetFloat(tagFoodSaturationLevel, nbt.FloatTag(h.hungerManager.GetSaturation()))
	tag.SetInt(tagFoodTickTimer, nbt.IntTag(h.hungerManager.GetFoodTickTimer()))

	tag.SetInt(tagXpLevel, nbt.IntTag(h.xpManager.GetXpLevel()))
	tag.SetFloat(tagXpProgress, nbt.FloatTag(h.xpManager.GetXpProgress()))
	tag.SetInt(tagLifetimeXpTotal, nbt.IntTag(h.xpManager.GetLifetimeTotalXp()))
	tag.SetInt(tagXpSeed, nbt.IntTag(h.xpSeed))

	var inventoryTags []nbt.Tag
	// Normal inventory
	hotbarSize := h.inventory.GetHotbarSize()
	slotCount := h.inventory.GetSize() + hotbarSize
	for slot := hotbarSize; slot < slotCount; slot++ {
		if it := h.inventory.GetItem(slot - 9); !it.IsNull() {
			if itemTag, err := item.NbtSerialize(it, slot); err == nil {
				inventoryTags = append(inventoryTags, itemTag)
			}
		}
	}
	// Armor
	for slot := 100; slot < 104; slot++ {
		if it := h.armorInventory.GetItem(slot - 100); !it.IsNull() {
			if itemTag, err := item.NbtSerialize(it, slot); err == nil {
				inventoryTags = append(inventoryTags, itemTag)
			}
		}
	}
	inventoryList, _ := nbt.NewListTag(inventoryTags, nbt.TagCompound)
	tag.SetTag(tagInventory, inventoryList)

	tag.SetInt(tagSelectedInventorySlot, nbt.IntTag(h.inventory.GetHeldItemIndex()))

	if offHandItem := h.offHandInventory.GetItem(0); !offHandItem.IsNull() {
		if itemTag, err := item.NbtSerialize(offHandItem, -1); err == nil {
			tag.SetTag(tagOffHandItem, itemTag)
		}
	}

	var enderTags []nbt.Tag
	for slot := 0; slot < h.enderInventory.GetSize(); slot++ {
		if it := h.enderInventory.GetItem(slot); !it.IsNull() {
			if itemTag, err := item.NbtSerialize(it, slot); err == nil {
				enderTags = append(enderTags, itemTag)
			}
		}
	}
	enderList, _ := nbt.NewListTag(enderTags, nbt.TagCompound)
	tag.SetTag(tagEnderChestInventory, enderList)

	tag.SetTag(tagSkin, nbt.NewCompoundTag().
		SetString(tagSkinName, nbt.StringTag(h.skin.GetSkinID())).
		SetByteArray(tagSkinData, nbt.ByteArrayTag(h.skin.GetSkinData())).
		SetByteArray(tagSkinCapeData, nbt.ByteArrayTag(h.skin.GetCapeData())).
		SetString(tagSkinGeometryName, nbt.StringTag(h.skin.GetGeometryName())).
		SetByteArray(tagSkinGeometryData, nbt.ByteArrayTag(h.skin.GetGeometryData())))

	return tag
}

// SpawnTo is a port of Human::spawnTo: a player is never spawned to itself.
func (h *Human) SpawnTo(player world.EntityViewer) {
	if viewerEntity, ok := player.(world.Entity); !ok || viewerEntity != world.Entity(h.self) {
		h.Living.SpawnTo(player)
	}
}

// SendSpawnPacket is a port of Human::sendSpawnPacket.
func (h *Human) SendSpawnPacket(player world.EntityViewer) {
	_, isPlayer := AsPlayer(h.self)
	if !isPlayer {
		player.SendPacket(&packet.PlayerList{Entries: []protocol.PlayerListEntry{{
			ActionType:     protocol.PlayerListActionAdd,
			UUID:           h.uuid,
			EntityUniqueID: int64(h.id),
			Username:       h.hself.GetName(),
			Skin:           SkinToNetwork(h.skin),
		}}})
	}

	player.SendPacket(&packet.AddPlayer{
		UUID:            h.GetUniqueID(),
		Username:        h.hself.GetName(),
		EntityRuntimeID: uint64(h.GetID()),
		PlatformChatID:  "",
		Position:        vec32(h.location.Vector3),
		Velocity:        vec32(h.GetMotion()),
		Pitch:           float32(h.location.Pitch),
		Yaw:             float32(h.location.Yaw),
		HeadYaw:         float32(h.location.Yaw), //TODO: head yaw
		HeldItem:        convert.ItemStackWrapperLegacy(h.GetInventory().GetItemInHand()),
		GameType:        0, // GameMode::SURVIVAL
		EntityMetadata:  h.GetAllNetworkData(),
		AbilityData: protocol.AbilityData{
			EntityUniqueID:     int64(h.GetID()), //TODO: this should be unique ID
			PlayerPermissions:  packet.PermissionLevelVisitor,
			CommandPermissions: protocol.CommandPermissionLevelAny,
			Layers: []protocol.AbilityLayer{{
				Type:      protocol.AbilityLayerTypeBase,
				Abilities: protocol.AbilityCount - 1, // every ability present, all false (array_fill(..., false))
				Values:    0,
			}},
		},
		//TODO: entity links
		DeviceID:      "", //we intentionally don't send this - secvuln
		BuildPlatform: -1, //DeviceOS::UNKNOWN - we intentionally don't send this (secvuln)
	})

	//TODO: Hack for MCPE 1.2.13: DATA_NAMETAG is useless in AddPlayerPacket, so it has to be sent separately
	h.SendData([]world.EntityViewer{player}, protocol.EntityMetadata{MetadataNametag: h.GetNameTag()})

	player.SendPacket(mobArmorChangePacket(&h.Living))
	player.SendPacket(mobOffHandItemChangePacket(h))

	if !isPlayer {
		player.SendPacket(&packet.PlayerList{Entries: []protocol.PlayerListEntry{{ActionType: protocol.PlayerListActionRemove, UUID: h.uuid}}})
	}
}

// GetOffsetPosition is a port of Human::getOffsetPosition.
func (h *Human) GetOffsetPosition(v math.Vector3) math.Vector3 {
	return v.Add(0, 1.621, 0) //TODO: +0.001 hack for MCPE falling underground
}

// OnDispose is a port of Human::onDispose.
func (h *Human) OnDispose() {
	h.inventory.RemoveAllViewers()
	h.inventory.GetHeldItemIndexChangeListeners().Clear()
	h.offHandInventory.RemoveAllViewers()
	h.enderInventory.RemoveAllViewers()
	h.Living.OnDispose()
}

// populateInventoryFromListTag is a port of Human::populateInventoryFromListTag: sets the contents
// without notifying the inventory's listeners.
func populateInventoryFromListTag(inv inventory.Inventory, items map[int]item.Item) {
	listeners := inv.GetListeners().ToSlice()
	inv.GetListeners().Clear()
	inv.SetContents(items)
	inv.GetListeners().Add(listeners...)
}
