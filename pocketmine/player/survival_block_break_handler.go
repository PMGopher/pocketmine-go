package player

import (
	stdmath "math"

	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/world/particle"
	"pocketmine-go/pocketmine/world/sound"
)

// DefaultFxIntervalTicks mirrors SurvivalBlockBreakHandler::DEFAULT_FX_INTERVAL_TICKS.
const DefaultFxIntervalTicks = 5

// blockStartBreak/blockStopBreak/blockBreakSpeed mirror pocketmine\network\mcpe\protocol\types\
// LevelEvent::BLOCK_START_BREAK/BLOCK_STOP_BREAK/BLOCK_BREAK_SPEED - not otherwise exposed by
// gophertunnel's own (differently-named but numerically identical) LevelEventStartBlockCracking/
// StopBlockCracking/UpdateBlockCracking constants, kept as their real PHP names here for clarity
// at the call sites below.
const (
	blockStartBreak = packet.LevelEventStartBlockCracking
	blockStopBreak  = packet.LevelEventStopBlockCracking
	blockBreakSpeed = packet.LevelEventUpdateBlockCracking
)

// SurvivalBlockBreakHandler is a port of a slice of pocketmine\player\SurvivalBlockBreakHandler:
// the real break-time/break-progress state machine driving a held-down break action, tracked
// per-tick via Update. The BLOCK_START_BREAK/BLOCK_BREAK_SPEED/BLOCK_STOP_BREAK network broadcasts
// that drive the client's break-progress crack overlay are real (see World.BroadcastPacketToViewers),
// as is the punch particle/sound broadcast on every fx tick (World.AddParticle/AddSound).
//
// Not ported (each needs a real subsystem this port doesn't have yet, so each modifier below is
// simply never applied - documented, not guessed):
//   - Haste/Mining Fatigue effect modifiers on break speed: no EffectManager exists.
//   - Aqua Affinity underwater break-speed penalty removal: no ArmorInventory/enchantments exist.
//   - The ArmSwingAnimation broadcast in Update: no entity-animation packet plumbing exists yet
//     (a separate, smaller gap from the broadcast infrastructure itself, which now exists).
//   - The PHP destructor's BLOCK_STOP_BREAK broadcast: Go has no deterministic destructors: call
//     Close() explicitly when done with a handler instead (matching this port's tile.Close()-style
//     convention elsewhere).
type SurvivalBlockBreakHandler struct {
	player            *Player
	blockPos          math.Vector3
	block             block.Behavior
	targetedFace      math.Facing
	maxPlayerDistance int
	fxTickInterval    int

	fxTicker      int
	breakSpeed    float64
	breakProgress float64
}

// NewSurvivalBlockBreakHandler is a port of SurvivalBlockBreakHandler::__construct. heldItem is
// the item the player is holding right now (real PHP reads this from the player's inventory
// itself; this port's Player has no "selected hotbar slot" concept yet, so the caller supplies it
// directly instead - see Player's own doc comment on inventory windows not being fully ported).
func NewSurvivalBlockBreakHandler(p *Player, blockPos math.Vector3, blk block.Behavior, targetedFace math.Facing, maxPlayerDistance int, heldItem block.Item) *SurvivalBlockBreakHandler {
	h := &SurvivalBlockBreakHandler{
		player:            p,
		blockPos:          blockPos,
		block:             blk,
		targetedFace:      targetedFace,
		maxPlayerDistance: maxPlayerDistance,
		fxTickInterval:    DefaultFxIntervalTicks,
	}
	h.breakSpeed = h.calculateBreakProgressPerTick(heldItem)
	if h.breakSpeed > 0 {
		h.player.GetWorld().BroadcastPacketToViewers(h.blockPos, levelEvent(blockStartBreak, int32(65535*h.breakSpeed), h.blockPos))
	}
	return h
}

// calculateBreakProgressPerTick is a port of SurvivalBlockBreakHandler::calculateBreakProgressPerTick -
// see the type's own doc comment for the effect/enchantment modifiers deliberately left unapplied.
func (h *SurvivalBlockBreakHandler) calculateBreakProgressPerTick(heldItem block.Item) float64 {
	if !h.block.GetBreakInfo().IsBreakable() {
		return 0
	}

	breakTime, err := h.block.GetBreakInfo().GetBreakTime(heldItem)
	if err != nil {
		return 0
	}
	breakTimePerTick := breakTime * 20

	if !h.player.IsOnGround() && !h.player.IsFlying() {
		breakTimePerTick *= 5
	}

	if breakTimePerTick > 0 {
		return 1 / breakTimePerTick
	}
	return 1
}

// Update is a port of SurvivalBlockBreakHandler::update - returns false once the player has moved
// too far from the block (the caller should cancel the break), true otherwise (including when
// breakProgress has reached 1 and the block should now actually break - matching the real method's
// own return value, which real PHP's own caller checks separately from breakProgress itself).
func (h *SurvivalBlockBreakHandler) Update(heldItem block.Item) bool {
	center := h.blockPos.Add(0.5, 0.5, 0.5)
	maxDistSq := float64(h.maxPlayerDistance * h.maxPlayerDistance)
	if h.player.GetPosition().DistanceSquared(center) > maxDistSq {
		return false
	}

	newBreakSpeed := h.calculateBreakProgressPerTick(heldItem)
	if stdmath.Abs(newBreakSpeed-h.breakSpeed) > 0.0001 {
		h.breakSpeed = newBreakSpeed
		h.player.GetWorld().BroadcastPacketToViewers(h.blockPos, levelEvent(blockBreakSpeed, int32(65535*h.breakSpeed), h.blockPos))
	}

	h.breakProgress += h.breakSpeed

	if h.fxTicker%h.fxTickInterval == 0 && h.breakProgress < 1 {
		w := h.player.GetWorld()
		w.AddParticle(h.blockPos, particle.BlockPunchParticle{BlockStateID: h.block.GetStateId(), Face: h.targetedFace})
		w.AddSound(h.blockPos, sound.BlockPunchSound{BlockStateID: h.block.GetStateId()})
	}
	h.fxTicker++

	return h.breakProgress < 1
}

// GetBlockPos is a port of SurvivalBlockBreakHandler::getBlockPos.
func (h *SurvivalBlockBreakHandler) GetBlockPos() math.Vector3 { return h.blockPos }

// GetTargetedFace is a port of SurvivalBlockBreakHandler::getTargetedFace.
func (h *SurvivalBlockBreakHandler) GetTargetedFace() math.Facing { return h.targetedFace }

// SetTargetedFace is a port of SurvivalBlockBreakHandler::setTargetedFace.
func (h *SurvivalBlockBreakHandler) SetTargetedFace(face math.Facing) {
	math.ValidateFacing(face)
	h.targetedFace = face
}

// GetBreakSpeed is a port of SurvivalBlockBreakHandler::getBreakSpeed.
func (h *SurvivalBlockBreakHandler) GetBreakSpeed() float64 { return h.breakSpeed }

// GetBreakProgress is a port of SurvivalBlockBreakHandler::getBreakProgress.
func (h *SurvivalBlockBreakHandler) GetBreakProgress() float64 { return h.breakProgress }

// Close is a Go-idiomatic stand-in for SurvivalBlockBreakHandler::__destruct (Go has no
// deterministic destructors) - broadcasts the real BLOCK_STOP_BREAK event that stops the client's
// break-progress crack overlay, guarded by the same isInLoadedTerrain check real PHP's destructor
// makes (a terrain-unloaded block has no viewers left to broadcast to anyway, but real PHP's check
// specifically guards against firing during shutdown/unload sequences).
func (h *SurvivalBlockBreakHandler) Close() {
	if h.player.GetWorld().IsChunkLoaded(h.blockPos.FloorX()>>4, h.blockPos.FloorZ()>>4) {
		h.player.GetWorld().BroadcastPacketToViewers(h.blockPos, levelEvent(blockStopBreak, 0, h.blockPos))
	}
}
