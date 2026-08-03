// Package particle is a port of pocketmine\world\particle.
package particle

import (
	"fmt"

	"pocketmine-go/pocketmine/color"
	"pocketmine-go/pocketmine/math"
)

// Particle is a port of pocketmine\world\particle\Particle. The real interface has one method,
// Encode(pos Vector3) []ClientboundPacket - deferred here exactly like world/sound.Sound's own
// Encode (see its doc comment): ClientboundPacket belongs to the unported network/mcpe/protocol
// package, so every concrete type below is a marker only, carrying the real data PHP's own
// constructor takes but with nothing yet to turn it into an actual packet.
type Particle interface {
	//marker — Encode(pos math.Vector3) []ClientboundPacket once network/mcpe/protocol is ported
}

// AngryVillagerParticle is a port of pocketmine\world\particle\AngryVillagerParticle.
type AngryVillagerParticle struct{}

// BlockBreakParticle is a port of pocketmine\world\particle\BlockBreakParticle.
//
// Stores the block's state ID rather than the whole block.Behavior - same reasoning as
// sound.ItemUseOnBlockSound: storing a block.Behavior here would create an import cycle once
// anything in the block package wants to construct one of these (block already imports
// world/sound for the identical reason).
type BlockBreakParticle struct{ BlockStateID int }

// BlockForceFieldParticle is a port of pocketmine\world\particle\BlockForceFieldParticle.
type BlockForceFieldParticle struct{ Data int }

// BlockPunchParticle is a port of pocketmine\world\particle\BlockPunchParticle. Stores the block's
// state ID rather than the whole block.Behavior - same reasoning as BlockBreakParticle.
type BlockPunchParticle struct {
	BlockStateID int
	Face         math.Facing
}

// BubbleParticle is a port of pocketmine\world\particle\BubbleParticle.
type BubbleParticle struct{}

// CriticalParticle is a port of pocketmine\world\particle\CriticalParticle.
type CriticalParticle struct{ Scale int }

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

// DustParticle is a port of pocketmine\world\particle\DustParticle.
type DustParticle struct{ Color color.Color }

// EnchantParticle is a port of pocketmine\world\particle\EnchantParticle.
type EnchantParticle struct{ Color color.Color }

// EnchantmentTableParticle is a port of pocketmine\world\particle\EnchantmentTableParticle.
type EnchantmentTableParticle struct{}

// EndermanTeleportParticle is a port of pocketmine\world\particle\EndermanTeleportParticle.
type EndermanTeleportParticle struct{}

// EntityFlameParticle is a port of pocketmine\world\particle\EntityFlameParticle.
type EntityFlameParticle struct{}

// ExplodeParticle is a port of pocketmine\world\particle\ExplodeParticle.
type ExplodeParticle struct{}

// FlameParticle is a port of pocketmine\world\particle\FlameParticle.
type FlameParticle struct{}

// FloatingTextParticle is a port of pocketmine\world\particle\FloatingTextParticle.
type FloatingTextParticle struct{ Text, Title string }

// HappyVillagerParticle is a port of pocketmine\world\particle\HappyVillagerParticle.
type HappyVillagerParticle struct{}

// HeartParticle is a port of pocketmine\world\particle\HeartParticle.
type HeartParticle struct{ Scale int }

// HugeExplodeParticle is a port of pocketmine\world\particle\HugeExplodeParticle.
type HugeExplodeParticle struct{}

// HugeExplodeSeedParticle is a port of pocketmine\world\particle\HugeExplodeSeedParticle.
type HugeExplodeSeedParticle struct{}

// InkParticle is a port of pocketmine\world\particle\InkParticle.
type InkParticle struct{ Scale int }

// InstantEnchantParticle is a port of pocketmine\world\particle\InstantEnchantParticle.
type InstantEnchantParticle struct{ Color color.Color }

// ItemBreakParticle is a port of pocketmine\world\particle\ItemBreakParticle. Stores the item's
// bare type ID rather than a whole item.Item - same reasoning as BlockBreakParticle, and this
// port's item package isn't imported here for the same "avoid a future import cycle" reason.
type ItemBreakParticle struct{ ItemTypeID int }

// LavaDripParticle is a port of pocketmine\world\particle\LavaDripParticle.
type LavaDripParticle struct{}

// LavaParticle is a port of pocketmine\world\particle\LavaParticle.
type LavaParticle struct{}

// MobSpawnParticle is a port of pocketmine\world\particle\MobSpawnParticle.
type MobSpawnParticle struct{ Width, Height int }

// PortalParticle is a port of pocketmine\world\particle\PortalParticle.
type PortalParticle struct{}

// PotionSplashParticle is a port of pocketmine\world\particle\PotionSplashParticle.
type PotionSplashParticle struct{ Color color.Color }

// DefaultPotionSplashColor mirrors PotionSplashParticle::getWaterBottleSplashColor's default
// water-bottle splash colour (0x385dc6), matching the real PHP TODO's own placeholder value.
var DefaultPotionSplashColor = color.FromRGB(0x385dc6)

// RainSplashParticle is a port of pocketmine\world\particle\RainSplashParticle.
type RainSplashParticle struct{}

// RedstoneParticle is a port of pocketmine\world\particle\RedstoneParticle.
type RedstoneParticle struct{ Lifetime int }

// SmokeParticle is a port of pocketmine\world\particle\SmokeParticle.
type SmokeParticle struct{ Scale int }

// SnowballPoofParticle is a port of pocketmine\world\particle\SnowballPoofParticle.
type SnowballPoofParticle struct{}

// SonicExplosionParticle is a port of pocketmine\world\particle\SonicExplosionParticle.
type SonicExplosionParticle struct{}

// SplashParticle is a port of pocketmine\world\particle\SplashParticle.
type SplashParticle struct{}

// SporeParticle is a port of pocketmine\world\particle\SporeParticle.
type SporeParticle struct{}

// TerrainParticle is a port of pocketmine\world\particle\TerrainParticle. Stores the block's state
// ID rather than the whole block.Behavior - same reasoning as BlockBreakParticle.
type TerrainParticle struct{ BlockStateID int }

// WaterDripParticle is a port of pocketmine\world\particle\WaterDripParticle.
type WaterDripParticle struct{}

// WaterParticle is a port of pocketmine\world\particle\WaterParticle.
type WaterParticle struct{}
