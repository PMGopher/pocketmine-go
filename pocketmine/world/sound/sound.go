// Package sound is a port of pocketmine\world\sound.
package sound

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/math"
)

// blockNetworkTranslator is the local surface Encode needs to turn an internal block state ID into
// the Bedrock network runtime ID a real LevelSoundEventPacket/LevelEventPacket needs - declared
// locally (matching this port's established forward-compatible-local-interface convention, see
// e.g. world.fullCubeChecker) rather than importing
// pocketmine/network/mcpe/convert.BlockTranslator directly, which would create an import cycle
// (convert imports block, block imports this package for RedstonePowerOnSound/OffSound).
// *convert.BlockTranslator satisfies this structurally via its own NetworkIDForCachedState.
type blockNetworkTranslator interface {
	NetworkIDForCachedState(internalStateID int32) int32
}

// Sound is a port of pocketmine\world\sound\Sound. Encode takes a blockNetworkTranslator (real PHP
// reaches a global TypeConverter::getInstance() singleton instead - this port has no singletons,
// every World owns its own translator instance, see World.AddSound) - unused by most Sound types,
// but kept uniform across every Encode method rather than only on the handful that need it.
type Sound interface {
	Encode(pos math.Vector3, translator blockNetworkTranslator) []packet.Packet
}

func vec3(pos math.Vector3) mgl32.Vec3 {
	return mgl32.Vec3{float32(pos.X), float32(pos.Y), float32(pos.Z)}
}

// nonActorSound mirrors LevelSoundEventPacket::nonActorSound's real defaults exactly (entityType
// ":", isBabyMob false, actorUniqueId -1, firePosition unset) - verified against
// github.com/pmmp/BedrockProtocol's LevelSoundEventPacket.php.
func nonActorSound(soundType string, pos math.Vector3, disableRelativeVolume bool, extraData int32) []packet.Packet {
	return []packet.Packet{&packet.LevelSoundEvent{
		SoundType:             soundType,
		Position:              vec3(pos),
		ExtraData:             extraData,
		EntityType:            ":",
		BabyMob:               false,
		DisableRelativeVolume: disableRelativeVolume,
		EntityUniqueID:        -1,
	}}
}

func levelEventSound(eventType int32, eventData int32, pos math.Vector3) []packet.Packet {
	return []packet.Packet{&packet.LevelEvent{EventType: eventType, Position: vec3(pos), EventData: eventData}}
}

// RedstonePowerOnSound is a port of pocketmine\world\sound\RedstonePowerOnSound.
type RedstonePowerOnSound struct{}

func (RedstonePowerOnSound) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return nonActorSound(packet.SoundEventPowerOn, pos, false, -1)
}

// RedstonePowerOffSound is a port of pocketmine\world\sound\RedstonePowerOffSound.
type RedstonePowerOffSound struct{}

func (RedstonePowerOffSound) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return nonActorSound(packet.SoundEventPowerOff, pos, false, -1)
}

// ItemUseOnBlockSound is a port of pocketmine\world\sound\ItemUseOnBlockSound.
//
// The PHP original stores the whole Block (its Encode() only ever reads getStateId() from it).
// Storing block.Behavior here instead would create an import cycle (block already imports this
// package for RedstonePowerOnSound/OffSound), so this stores just the state ID that Encode()
// actually needs.
type ItemUseOnBlockSound struct {
	BlockStateID int
}

func (s ItemUseOnBlockSound) Encode(pos math.Vector3, translator blockNetworkTranslator) []packet.Packet {
	return nonActorSound(packet.SoundEventItemUseOn, pos, false, translator.NetworkIDForCachedState(int32(s.BlockStateID)))
}

// AnvilFallSound is a port of pocketmine\world\sound\AnvilFallSound.
type AnvilFallSound struct{}

func (AnvilFallSound) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return levelEventSound(packet.LevelEventSoundAnvilLand, 0, pos)
}

// CopperWaxApplySound is a port of pocketmine\world\sound\CopperWaxApplySound.
type CopperWaxApplySound struct{}

func (CopperWaxApplySound) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return nonActorSound(packet.SoundEventCopperWaxOn, pos, false, -1)
}

// CopperWaxRemoveSound is a port of pocketmine\world\sound\CopperWaxRemoveSound.
type CopperWaxRemoveSound struct{}

func (CopperWaxRemoveSound) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return nonActorSound(packet.SoundEventCopperWaxOff, pos, false, -1)
}

// ScrapeSound is a port of pocketmine\world\sound\ScrapeSound.
type ScrapeSound struct{}

func (ScrapeSound) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return nonActorSound(packet.SoundEventScrape, pos, false, -1)
}

// DoorSound is a port of pocketmine\world\sound\DoorSound.
type DoorSound struct{ Pitch float64 }

func (s DoorSound) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return levelEventSound(packet.LevelEventSoundOpenDoor, int32(s.Pitch*1000), pos)
}

// PressurePlateActivateSound is a port of pocketmine\world\sound\PressurePlateActivateSound.
// Stores the block's state ID rather than the whole Block - same reasoning as ItemUseOnBlockSound.
type PressurePlateActivateSound struct {
	BlockStateID int
}

func (s PressurePlateActivateSound) Encode(pos math.Vector3, translator blockNetworkTranslator) []packet.Packet {
	return nonActorSound(packet.SoundEventPressurePlateClickOn, pos, false, translator.NetworkIDForCachedState(int32(s.BlockStateID)))
}

// PressurePlateDeactivateSound is a port of pocketmine\world\sound\PressurePlateDeactivateSound.
type PressurePlateDeactivateSound struct {
	BlockStateID int
}

func (s PressurePlateDeactivateSound) Encode(pos math.Vector3, translator blockNetworkTranslator) []packet.Packet {
	return nonActorSound(packet.SoundEventPressurePlateClickOff, pos, false, translator.NetworkIDForCachedState(int32(s.BlockStateID)))
}

// AmethystBlockChimeSound is a port of pocketmine\world\sound\AmethystBlockChimeSound.
type AmethystBlockChimeSound struct{}

func (AmethystBlockChimeSound) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return nonActorSound(packet.SoundEventAmethystBlockChime, pos, false, -1)
}

// BlockPunchSound is a port of pocketmine\world\sound\BlockPunchSound. Stores the block's state
// ID rather than the whole Block - same reasoning as ItemUseOnBlockSound.
type BlockPunchSound struct {
	BlockStateID int
}

func (s BlockPunchSound) Encode(pos math.Vector3, translator blockNetworkTranslator) []packet.Packet {
	return nonActorSound(packet.SoundEventHit, pos, false, translator.NetworkIDForCachedState(int32(s.BlockStateID)))
}

// BlockBreakSound is a port of pocketmine\world\sound\BlockBreakSound. Stores the block's state ID
// rather than the whole Block - same reasoning as ItemUseOnBlockSound. Not present in this port
// until now - a genuine gap found and closed while wiring real block-break networking.
type BlockBreakSound struct {
	BlockStateID int
}

func (s BlockBreakSound) Encode(pos math.Vector3, translator blockNetworkTranslator) []packet.Packet {
	return nonActorSound(packet.SoundEventBreak, pos, false, translator.NetworkIDForCachedState(int32(s.BlockStateID)))
}

// BlockPlaceSound is a port of pocketmine\world\sound\BlockPlaceSound. Stores the block's state ID
// rather than the whole Block - same reasoning as ItemUseOnBlockSound. Not present in this port
// until now - see BlockBreakSound's own doc comment.
type BlockPlaceSound struct {
	BlockStateID int
}

func (s BlockPlaceSound) Encode(pos math.Vector3, translator blockNetworkTranslator) []packet.Packet {
	return nonActorSound(packet.SoundEventPlace, pos, false, translator.NetworkIDForCachedState(int32(s.BlockStateID)))
}

// BlazeShootSound is a port of pocketmine\world\sound\BlazeShootSound.
type BlazeShootSound struct{}

func (BlazeShootSound) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return levelEventSound(packet.LevelEventSoundBlazeFireball, 0, pos)
}

// FireExtinguishSound is a port of pocketmine\world\sound\FireExtinguishSound.
type FireExtinguishSound struct{}

func (FireExtinguishSound) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return nonActorSound(packet.SoundEventExtinguishFire, pos, false, -1)
}

// FlintSteelSound is a port of pocketmine\world\sound\FlintSteelSound.
type FlintSteelSound struct{}

func (FlintSteelSound) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return nonActorSound(packet.SoundEventIgnite, pos, false, -1)
}

// ItemFrameAddItemSound is a port of pocketmine\world\sound\ItemFrameAddItemSound.
type ItemFrameAddItemSound struct{}

func (ItemFrameAddItemSound) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return levelEventSound(packet.LevelEventSoundAddItem, 0, pos)
}

// ItemFrameRemoveItemSound is a port of pocketmine\world\sound\ItemFrameRemoveItemSound.
type ItemFrameRemoveItemSound struct{}

func (ItemFrameRemoveItemSound) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return levelEventSound(packet.LevelEventSoundItemFrameRemoveItem, 0, pos)
}

// ItemFrameRotateItemSound is a port of pocketmine\world\sound\ItemFrameRotateItemSound.
type ItemFrameRotateItemSound struct{}

func (ItemFrameRotateItemSound) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return levelEventSound(packet.LevelEventSoundItemFrameRotateItem, 0, pos)
}

// LecternPlaceBookSound is a port of pocketmine\world\sound\LecternPlaceBookSound.
type LecternPlaceBookSound struct{}

func (LecternPlaceBookSound) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return nonActorSound(packet.SoundEventLecternBookPlace, pos, false, -1)
}

// BellRingSound is a port of pocketmine\world\sound\BellRingSound.
type BellRingSound struct{}

func (BellRingSound) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return nonActorSound(packet.SoundEventBell, pos, false, -1)
}

// DripleafTiltDownSound is a port of pocketmine\world\sound\DripleafTiltDownSound.
type DripleafTiltDownSound struct{}

func (DripleafTiltDownSound) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return nonActorSound(packet.SoundEventBigDripleafTiltDown, pos, false, -1)
}

// DripleafTiltUpSound is a port of pocketmine\world\sound\DripleafTiltUpSound.
type DripleafTiltUpSound struct{}

func (DripleafTiltUpSound) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return nonActorSound(packet.SoundEventBigDripleafTiltUp, pos, false, -1)
}

// RespawnAnchorChargeSound is a port of pocketmine\world\sound\RespawnAnchorChargeSound.
type RespawnAnchorChargeSound struct{}

func (RespawnAnchorChargeSound) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return nonActorSound(packet.SoundEventRespawnAnchorCharge, pos, false, -1)
}

// RespawnAnchorSetSpawnSound is a port of pocketmine\world\sound\RespawnAnchorSetSpawnSound.
type RespawnAnchorSetSpawnSound struct{}

func (RespawnAnchorSetSpawnSound) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return nonActorSound(packet.SoundEventRespawnAnchorSetSpawn, pos, false, -1)
}

// BucketFillWaterSound is a port of pocketmine\world\sound\BucketFillWaterSound.
type BucketFillWaterSound struct{}

func (BucketFillWaterSound) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return nonActorSound(packet.SoundEventBucketFillWater, pos, false, -1)
}

// BucketEmptyWaterSound is a port of pocketmine\world\sound\BucketEmptyWaterSound.
type BucketEmptyWaterSound struct{}

func (BucketEmptyWaterSound) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return nonActorSound(packet.SoundEventBucketEmptyWater, pos, false, -1)
}

// BucketFillLavaSound is a port of pocketmine\world\sound\BucketFillLavaSound.
type BucketFillLavaSound struct{}

func (BucketFillLavaSound) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return nonActorSound(packet.SoundEventBucketFillLava, pos, false, -1)
}

// BucketEmptyLavaSound is a port of pocketmine\world\sound\BucketEmptyLavaSound.
type BucketEmptyLavaSound struct{}

func (BucketEmptyLavaSound) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return nonActorSound(packet.SoundEventBucketEmptyLava, pos, false, -1)
}

// DyeUseSound is a port of pocketmine\world\sound\DyeUseSound.
type DyeUseSound struct{}

func (DyeUseSound) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return levelEventSound(packet.LevelEventSoundDyeUsed, 0, pos)
}

// InkSacUseSound is a port of pocketmine\world\sound\InkSacUseSound.
type InkSacUseSound struct{}

func (InkSacUseSound) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return levelEventSound(packet.LevelEventSoundInkSacUsed, 0, pos)
}

// FireworkExplosionSound is a port of pocketmine\world\sound\FireworkExplosionSound.
type FireworkExplosionSound struct{}

func (FireworkExplosionSound) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return nonActorSound(packet.SoundEventBlast, pos, false, -1)
}

// FireworkLargeExplosionSound is a port of pocketmine\world\sound\FireworkLargeExplosionSound.
type FireworkLargeExplosionSound struct{}

func (FireworkLargeExplosionSound) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return nonActorSound(packet.SoundEventLargeBlast, pos, false, -1)
}

// ChestOpenSound is a port of pocketmine\world\sound\ChestOpenSound.
type ChestOpenSound struct{}

func (ChestOpenSound) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return nonActorSound(packet.SoundEventChestOpen, pos, false, -1)
}

// ChestCloseSound is a port of pocketmine\world\sound\ChestCloseSound.
type ChestCloseSound struct{}

func (ChestCloseSound) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return nonActorSound(packet.SoundEventChestClosed, pos, false, -1)
}

// recordSoundTypes is a port of RecordSound::encode's match expression, verified against real PHP.
var recordSoundTypes = map[blockutils.RecordType]string{
	blockutils.RecordTypeDisk13:              packet.SoundEventRecord13,
	blockutils.RecordTypeDisk5:               packet.SoundEventRecord5,
	blockutils.RecordTypeDiskCat:             packet.SoundEventRecordCat,
	blockutils.RecordTypeDiskBlocks:          packet.SoundEventRecordBlocks,
	blockutils.RecordTypeDiskChirp:           packet.SoundEventRecordChirp,
	blockutils.RecordTypeDiskCreator:         packet.SoundEventRecordCreator,
	blockutils.RecordTypeDiskCreatorMusicBox: packet.SoundEventRecordCreatorMusicBox,
	blockutils.RecordTypeDiskFar:             packet.SoundEventRecordFar,
	blockutils.RecordTypeDiskLavaChicken:     packet.SoundEventRecordLavaChicken,
	blockutils.RecordTypeDiskMall:            packet.SoundEventRecordMall,
	blockutils.RecordTypeDiskMellohi:         packet.SoundEventRecordMellohi,
	blockutils.RecordTypeDiskOtherside:       packet.SoundEventRecordOtherside,
	blockutils.RecordTypeDiskPigstep:         packet.SoundEventRecordPigstep,
	blockutils.RecordTypeDiskPrecipice:       packet.SoundEventRecordPrecipice,
	blockutils.RecordTypeDiskRelic:           packet.SoundEventRecordRelic,
	blockutils.RecordTypeDiskStal:            packet.SoundEventRecordStal,
	blockutils.RecordTypeDiskStrad:           packet.SoundEventRecordStrad,
	blockutils.RecordTypeDiskWard:            packet.SoundEventRecordWard,
	blockutils.RecordTypeDisk11:              packet.SoundEventRecord11,
	blockutils.RecordTypeDiskWait:            packet.SoundEventRecordWait,
}

// RecordSound is a port of pocketmine\world\sound\RecordSound. Stores the record's RecordType
// rather than encoding it to a LevelSoundEvent - same reasoning as BlockPunchSound storing a
// state ID instead of encoding a full packet.
type RecordSound struct {
	RecordType blockutils.RecordType
}

func (s RecordSound) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return nonActorSound(recordSoundTypes[s.RecordType], pos, false, -1)
}

// RecordStopSound is a port of pocketmine\world\sound\RecordStopSound.
type RecordStopSound struct{}

func (RecordStopSound) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return nonActorSound(packet.SoundEventRecordNull, pos, false, -1)
}

// FizzSound is a port of pocketmine\world\sound\FizzSound.
type FizzSound struct {
	Pitch float64
}

func (s FizzSound) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return levelEventSound(packet.LevelEventSoundFizz, int32(s.Pitch*1000), pos)
}

// ChorusFlowerGrowSound is a port of pocketmine\world\sound\ChorusFlowerGrowSound.
type ChorusFlowerGrowSound struct{}

func (ChorusFlowerGrowSound) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return nonActorSound(packet.SoundEventChorusGrow, pos, false, -1)
}

// ChorusFlowerDieSound is a port of pocketmine\world\sound\ChorusFlowerDieSound.
type ChorusFlowerDieSound struct{}

func (ChorusFlowerDieSound) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return nonActorSound(packet.SoundEventChorusDeath, pos, false, -1)
}

// ExplodeSound is a port of pocketmine\world\sound\ExplodeSound.
type ExplodeSound struct{}

func (ExplodeSound) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return nonActorSound(packet.SoundEventExplode, pos, false, -1)
}

// entityAttackSound mirrors the exact (non-nonActorSound) LevelSoundEventPacket::create call both
// EntityAttackSound/EntityAttackNoDamageSound use in real PHP - entityType "minecraft:player"
// rather than the ":" default, verified against real PHP source.
func entityAttackSound(soundType string, pos math.Vector3) []packet.Packet {
	return []packet.Packet{&packet.LevelSoundEvent{
		SoundType:             soundType,
		Position:              vec3(pos),
		ExtraData:             -1,
		EntityType:            "minecraft:player",
		BabyMob:               false,
		DisableRelativeVolume: false,
		EntityUniqueID:        -1,
	}}
}

// EntityAttackSound is a port of pocketmine\world\sound\EntityAttackSound. Real PHP uses
// LevelSoundEvent::ATTACK_STRONG rather than the plain ATTACK constant, per its own "TODO: seems
// like ATTACK is dysfunctional" comment - kept exactly as-is.
type EntityAttackSound struct{}

func (EntityAttackSound) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return entityAttackSound(packet.SoundEventAttackStrong, pos)
}

// EntityAttackNoDamageSound is a port of pocketmine\world\sound\EntityAttackNoDamageSound.
type EntityAttackNoDamageSound struct{}

func (EntityAttackNoDamageSound) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return entityAttackSound(packet.SoundEventAttackNoDamage, pos)
}

// EntityLongFallSound is a port of pocketmine\world\sound\EntityLongFallSound. Stores the entity's
// network type ID/unique ID rather than the whole Entity - same reasoning as
// ItemUseOnBlockSound storing a bare state ID (this package can't import entity.Entity without
// risking a future cycle once entity ever needs a Sound of its own).
type EntityLongFallSound struct {
	EntityNetworkTypeID string
	EntityUniqueID      int64
}

func (s EntityLongFallSound) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return []packet.Packet{&packet.LevelSoundEvent{
		SoundType:      packet.SoundEventFallBig,
		Position:       vec3(pos),
		ExtraData:      -1,
		EntityType:     s.EntityNetworkTypeID,
		EntityUniqueID: s.EntityUniqueID,
	}}
}

// EntityShortFallSound is a port of pocketmine\world\sound\EntityShortFallSound.
type EntityShortFallSound struct{ EntityNetworkTypeID string }

func (s EntityShortFallSound) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return []packet.Packet{&packet.LevelSoundEvent{
		SoundType:      packet.SoundEventFallSmall,
		Position:       vec3(pos),
		ExtraData:      -1,
		EntityType:     s.EntityNetworkTypeID,
		EntityUniqueID: -1,
	}}
}

// EntityLandSound is a port of pocketmine\world\sound\EntityLandSound. Stores the landed-on
// block's state ID rather than the whole Block - same reasoning as ItemUseOnBlockSound.
type EntityLandSound struct {
	BlockStateID        int
	EntityNetworkTypeID string
	EntityUniqueID      int64
}

func (s EntityLandSound) Encode(pos math.Vector3, translator blockNetworkTranslator) []packet.Packet {
	return []packet.Packet{&packet.LevelSoundEvent{
		SoundType:      packet.SoundEventLand,
		Position:       vec3(pos),
		ExtraData:      translator.NetworkIDForCachedState(int32(s.BlockStateID)),
		EntityType:     s.EntityNetworkTypeID,
		EntityUniqueID: s.EntityUniqueID,
	}}
}
