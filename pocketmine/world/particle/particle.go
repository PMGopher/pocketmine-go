// Package particle is a port of pocketmine\world\particle.
package particle

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

	"pocketmine-go/pocketmine/color"
	"pocketmine-go/pocketmine/math"
)

// blockNetworkTranslator mirrors world/sound.Sound's own local interface of the same name/purpose
// (see its doc comment for why this is declared locally rather than importing
// pocketmine/network/mcpe/convert.BlockTranslator directly - the exact same import-cycle concern
// applies here, since block also imports this package for BlockBreakParticle/BlockPunchParticle).
type blockNetworkTranslator interface {
	NetworkIDForCachedState(internalStateID int32) int32
}

// Particle is a port of pocketmine\world\particle\Particle. Encode takes a blockNetworkTranslator
// for the same reason as world/sound.Sound.Encode - see its doc comment.
type Particle interface {
	Encode(pos math.Vector3, translator blockNetworkTranslator) []packet.Packet
}

func vec3(pos math.Vector3) mgl32.Vec3 {
	return mgl32.Vec3{float32(pos.X), float32(pos.Y), float32(pos.Z)}
}

// standardParticle mirrors LevelEventPacket::standardParticle: eventId is ADD_PARTICLE_MASK
// (0x4000, matches packet.LevelEventParticleLegacyEvent) OR'd with the ParticleIds constant,
// verified against github.com/pmmp/BedrockProtocol's ParticleIds.php/LevelEventPacket.php.
func standardParticle(particleID int32, data int32, pos math.Vector3) []packet.Packet {
	return []packet.Packet{&packet.LevelEvent{
		EventType: packet.LevelEventParticleLegacyEvent | particleID,
		Position:  vec3(pos),
		EventData: data,
	}}
}

func levelEventParticle(eventType int32, data int32, pos math.Vector3) []packet.Packet {
	return []packet.Packet{&packet.LevelEvent{EventType: eventType, Position: vec3(pos), EventData: data}}
}

// Real ParticleIds numeric constants (pocketmine\network\mcpe\protocol\types\ParticleIds) that
// aren't otherwise exposed by gophertunnel - verified against github.com/pmmp/BedrockProtocol.
const (
	particleIDVillagerAngry    = 40
	particleIDBlockForceField  = 4
	particleIDBubble           = 1
	particleIDCritical         = 3
	particleIDDust             = 33
	particleIDMobSpell         = 34
	particleIDEnchantmentTable = 42
	particleIDMobFlame         = 19
	particleIDExplode          = 6
	particleIDFlame            = 8
	particleIDVillagerHappy    = 41
	particleIDHeart            = 20
	particleIDHugeExplode      = 16
	particleIDHugeExplodeSeed  = 17
	particleIDInk              = 37
	particleIDMobSpellInstant  = 36
	particleIDItemBreak        = 14
	particleIDDripLava         = 29
	particleIDLava             = 10
	particleIDPortal           = 23
	particleIDRainSplash       = 39
	particleIDRedstone         = 12
	particleIDSmoke            = 5
	particleIDSnowballPoof     = 15
	particleIDSonicExplosion   = 85
	particleIDWaterSplash      = 25
	particleIDTownAura         = 22
	particleIDTerrain          = 21
	particleIDDripWater        = 28
	particleIDWaterWake        = 27
)

// AngryVillagerParticle is a port of pocketmine\world\particle\AngryVillagerParticle.
type AngryVillagerParticle struct{}

func (AngryVillagerParticle) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return standardParticle(particleIDVillagerAngry, 0, pos)
}

// BlockBreakParticle is a port of pocketmine\world\particle\BlockBreakParticle.
//
// Stores the block's state ID rather than the whole block.Behavior - same reasoning as
// sound.ItemUseOnBlockSound: storing a block.Behavior here would create an import cycle once
// anything in the block package wants to construct one of these (block already imports
// world/sound for the identical reason).
type BlockBreakParticle struct{ BlockStateID int }

func (p BlockBreakParticle) Encode(pos math.Vector3, translator blockNetworkTranslator) []packet.Packet {
	return levelEventParticle(packet.LevelEventParticlesDestroyBlock, translator.NetworkIDForCachedState(int32(p.BlockStateID)), pos)
}

// BlockForceFieldParticle is a port of pocketmine\world\particle\BlockForceFieldParticle.
type BlockForceFieldParticle struct{ Data int }

func (p BlockForceFieldParticle) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return standardParticle(particleIDBlockForceField, int32(p.Data), pos)
}

// BlockPunchParticle is a port of pocketmine\world\particle\BlockPunchParticle. Stores the block's
// state ID rather than the whole block.Behavior - same reasoning as BlockBreakParticle.
type BlockPunchParticle struct {
	BlockStateID int
	Face         math.Facing
}

func (p BlockPunchParticle) Encode(pos math.Vector3, translator blockNetworkTranslator) []packet.Packet {
	data := translator.NetworkIDForCachedState(int32(p.BlockStateID)) | (int32(p.Face) << 24)
	return levelEventParticle(packet.LevelEventParticlesCrackBlock, data, pos) //LevelEvent::PARTICLE_PUNCH_BLOCK = 2014
}

// BubbleParticle is a port of pocketmine\world\particle\BubbleParticle.
type BubbleParticle struct{}

func (BubbleParticle) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return standardParticle(particleIDBubble, 0, pos)
}

// CriticalParticle is a port of pocketmine\world\particle\CriticalParticle.
type CriticalParticle struct{ Scale int }

func (p CriticalParticle) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return standardParticle(particleIDCritical, int32(p.Scale), pos)
}

// dragonEggTeleportBound mirrors DragonEggTeleportParticle's own private boundOrThrow limit.
const dragonEggTeleportBound = 255

// DragonEggTeleportParticle is a port of pocketmine\world\particle\DragonEggTeleportParticle.
type DragonEggTeleportParticle struct {
	XDiff, YDiff, ZDiff int
}

// NewDragonEggTeleportParticle is a port of DragonEggTeleportParticle::__construct, including its
// boundOrThrow validation (-255..255 for each axis).
func NewDragonEggTeleportParticle(xDiff, yDiff, zDiff int) (*DragonEggTeleportParticle, error) {
	for _, v := range [3]int{xDiff, yDiff, zDiff} {
		if v < -dragonEggTeleportBound || v > dragonEggTeleportBound {
			return nil, fmt.Errorf("particle: value must be between -%d and %d, got %d", dragonEggTeleportBound, dragonEggTeleportBound, v)
		}
	}
	return &DragonEggTeleportParticle{XDiff: xDiff, YDiff: yDiff, ZDiff: zDiff}, nil
}

func abs(v int) int32 {
	if v < 0 {
		return int32(-v)
	}
	return int32(v)
}

func (p *DragonEggTeleportParticle) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	var data int32
	if p.ZDiff < 0 {
		data |= 1 << 26
	}
	if p.YDiff < 0 {
		data |= 1 << 25
	}
	if p.XDiff < 0 {
		data |= 1 << 24
	}
	data |= abs(p.XDiff) << 16
	data |= abs(p.YDiff) << 8
	data |= abs(p.ZDiff)
	return levelEventParticle(packet.LevelEventParticlesDragonEgg, data, pos) //LevelEvent::PARTICLE_DRAGON_EGG_TELEPORT = 2010
}

// DustParticle is a port of pocketmine\world\particle\DustParticle.
type DustParticle struct{ Color color.Color }

func (p DustParticle) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return standardParticle(particleIDDust, p.Color.ToARGB(), pos)
}

// EnchantParticle is a port of pocketmine\world\particle\EnchantParticle.
type EnchantParticle struct{ Color color.Color }

func (p EnchantParticle) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return standardParticle(particleIDMobSpell, p.Color.ToARGB(), pos)
}

// EnchantmentTableParticle is a port of pocketmine\world\particle\EnchantmentTableParticle.
type EnchantmentTableParticle struct{}

func (EnchantmentTableParticle) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return standardParticle(particleIDEnchantmentTable, 0, pos)
}

// EndermanTeleportParticle is a port of pocketmine\world\particle\EndermanTeleportParticle.
type EndermanTeleportParticle struct{}

func (EndermanTeleportParticle) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return levelEventParticle(packet.LevelEventParticlesTeleport, 0, pos) //LevelEvent::PARTICLE_ENDERMAN_TELEPORT = 2013
}

// EntityFlameParticle is a port of pocketmine\world\particle\EntityFlameParticle.
type EntityFlameParticle struct{}

func (EntityFlameParticle) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return standardParticle(particleIDMobFlame, 0, pos)
}

// ExplodeParticle is a port of pocketmine\world\particle\ExplodeParticle.
type ExplodeParticle struct{}

func (ExplodeParticle) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return standardParticle(particleIDExplode, 0, pos)
}

// FlameParticle is a port of pocketmine\world\particle\FlameParticle.
type FlameParticle struct{}

func (FlameParticle) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return standardParticle(particleIDFlame, 0, pos)
}

// FloatingTextParticle is a port of pocketmine\world\particle\FloatingTextParticle. Real PHP spawns
// a real (invisible, no-AI) fake entity carrying the text as its nametag, using
// AddActorPacket/SetActorDataPacket with a real EntityMetadataFlags/EntityMetadataProperties
// dictionary and Entity::nextRuntimeId() for identity. None of that entity-metadata/spawn-packet
// infrastructure exists anywhere in this port yet (no other feature needs it either), so this is a
// documented gap rather than a guess, matching this port's rule for gaps requiring an entire
// unported subsystem.
type FloatingTextParticle struct{ Text, Title string }

func (p *FloatingTextParticle) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return nil
}

// HappyVillagerParticle is a port of pocketmine\world\particle\HappyVillagerParticle.
type HappyVillagerParticle struct{}

func (HappyVillagerParticle) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return standardParticle(particleIDVillagerHappy, 0, pos)
}

// HeartParticle is a port of pocketmine\world\particle\HeartParticle.
type HeartParticle struct{ Scale int }

func (p HeartParticle) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return standardParticle(particleIDHeart, int32(p.Scale), pos)
}

// HugeExplodeParticle is a port of pocketmine\world\particle\HugeExplodeParticle.
type HugeExplodeParticle struct{}

func (HugeExplodeParticle) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return standardParticle(particleIDHugeExplode, 0, pos)
}

// HugeExplodeSeedParticle is a port of pocketmine\world\particle\HugeExplodeSeedParticle.
type HugeExplodeSeedParticle struct{}

func (HugeExplodeSeedParticle) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return standardParticle(particleIDHugeExplodeSeed, 0, pos)
}

// InkParticle is a port of pocketmine\world\particle\InkParticle.
type InkParticle struct{ Scale int }

func (p InkParticle) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return standardParticle(particleIDInk, int32(p.Scale), pos)
}

// InstantEnchantParticle is a port of pocketmine\world\particle\InstantEnchantParticle.
type InstantEnchantParticle struct{ Color color.Color }

func (p InstantEnchantParticle) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return standardParticle(particleIDMobSpellInstant, p.Color.ToARGB(), pos)
}

// ItemBreakParticle is a port of pocketmine\world\particle\ItemBreakParticle. Stores the item's
// bare type ID rather than a whole item.Item - same reasoning as BlockBreakParticle, and this
// port's item package isn't imported here for the same "avoid a future import cycle" reason.
//
// Real PHP encodes ($networkId << 16) | $networkMeta via a real item network translator
// (TypeConverter::getItemTranslator). This port has no item network translator yet (no other
// feature needs one either - items aren't sent over the network anywhere in this port so far), so
// ItemTypeID is used directly as a placeholder network ID with meta 0, a documented approximation
// rather than a fabricated translator.
type ItemBreakParticle struct{ ItemTypeID int }

func (p ItemBreakParticle) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return standardParticle(particleIDItemBreak, int32(p.ItemTypeID)<<16, pos)
}

// LavaDripParticle is a port of pocketmine\world\particle\LavaDripParticle.
type LavaDripParticle struct{}

func (LavaDripParticle) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return standardParticle(particleIDDripLava, 0, pos)
}

// LavaParticle is a port of pocketmine\world\particle\LavaParticle.
type LavaParticle struct{}

func (LavaParticle) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return standardParticle(particleIDLava, 0, pos)
}

// MobSpawnParticle is a port of pocketmine\world\particle\MobSpawnParticle.
type MobSpawnParticle struct{ Width, Height int }

func (p MobSpawnParticle) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	data := int32(p.Width&0xff) | (int32(p.Height&0xff) << 8)
	return levelEventParticle(packet.LevelEventParticlesMobBlockSpawn, data, pos) //LevelEvent::PARTICLE_SPAWN = 2004
}

// PortalParticle is a port of pocketmine\world\particle\PortalParticle.
type PortalParticle struct{}

func (PortalParticle) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return standardParticle(particleIDPortal, 0, pos)
}

// PotionSplashParticle is a port of pocketmine\world\particle\PotionSplashParticle.
type PotionSplashParticle struct{ Color color.Color }

// DefaultPotionSplashColor mirrors PotionSplashParticle::getWaterBottleSplashColor's default
// water-bottle splash colour (0x385dc6), matching the real PHP TODO's own placeholder value.
var DefaultPotionSplashColor = color.FromRGB(0x385dc6)

func (p PotionSplashParticle) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return levelEventParticle(packet.LevelEventParticlesPotionSplash, p.Color.ToARGB(), pos)
}

// RainSplashParticle is a port of pocketmine\world\particle\RainSplashParticle.
type RainSplashParticle struct{}

func (RainSplashParticle) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return standardParticle(particleIDRainSplash, 0, pos)
}

// RedstoneParticle is a port of pocketmine\world\particle\RedstoneParticle.
type RedstoneParticle struct{ Lifetime int }

func (p RedstoneParticle) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return standardParticle(particleIDRedstone, int32(p.Lifetime), pos)
}

// SmokeParticle is a port of pocketmine\world\particle\SmokeParticle.
type SmokeParticle struct{ Scale int }

func (p SmokeParticle) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return standardParticle(particleIDSmoke, int32(p.Scale), pos)
}

// SnowballPoofParticle is a port of pocketmine\world\particle\SnowballPoofParticle.
type SnowballPoofParticle struct{}

func (SnowballPoofParticle) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return standardParticle(particleIDSnowballPoof, 0, pos)
}

// SonicExplosionParticle is a port of pocketmine\world\particle\SonicExplosionParticle.
type SonicExplosionParticle struct{}

func (SonicExplosionParticle) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return standardParticle(particleIDSonicExplosion, 0, pos)
}

// SplashParticle is a port of pocketmine\world\particle\SplashParticle.
type SplashParticle struct{}

func (SplashParticle) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return standardParticle(particleIDWaterSplash, 0, pos)
}

// SporeParticle is a port of pocketmine\world\particle\SporeParticle.
type SporeParticle struct{}

func (SporeParticle) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return standardParticle(particleIDTownAura, 0, pos)
}

// TerrainParticle is a port of pocketmine\world\particle\TerrainParticle. Stores the block's state
// ID rather than the whole block.Behavior - same reasoning as BlockBreakParticle.
type TerrainParticle struct{ BlockStateID int }

func (p TerrainParticle) Encode(pos math.Vector3, translator blockNetworkTranslator) []packet.Packet {
	return standardParticle(particleIDTerrain, translator.NetworkIDForCachedState(int32(p.BlockStateID)), pos)
}

// WaterDripParticle is a port of pocketmine\world\particle\WaterDripParticle.
type WaterDripParticle struct{}

func (WaterDripParticle) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return standardParticle(particleIDDripWater, 0, pos)
}

// WaterParticle is a port of pocketmine\world\particle\WaterParticle.
type WaterParticle struct{}

func (WaterParticle) Encode(pos math.Vector3, _ blockNetworkTranslator) []packet.Packet {
	return standardParticle(particleIDWaterWake, 0, pos)
}
