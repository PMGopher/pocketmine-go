package player

import (
	stdmath "math"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/entity/animation"
	"pocketmine-go/pocketmine/entity/effect"
	"pocketmine-go/pocketmine/item/enchantment"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/world/particle"
	"pocketmine-go/pocketmine/world/sound"
)

// DefaultFxIntervalTicks is SurvivalBlockBreakHandler::DEFAULT_FX_INTERVAL_TICKS.
const DefaultFxIntervalTicks = 5

// SurvivalBlockBreakHandler is a port of pocketmine\player\SurvivalBlockBreakHandler: the break
// progress of a block a survival player is holding left click on. The client shows the crack
// overlay from the BLOCK_START_BREAK/BLOCK_BREAK_SPEED/BLOCK_STOP_BREAK level events.
//
// PHP sends BLOCK_STOP_BREAK from __destruct when the handler is dropped; Go has no destructors,
// so the player calls Close when it drops the handler.
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

// NewSurvivalBlockBreakHandler is a port of SurvivalBlockBreakHandler::__construct.
func NewSurvivalBlockBreakHandler(p *Player, blockPos math.Vector3, blk block.Behavior, targetedFace math.Facing, maxPlayerDistance int, fxTickInterval int) *SurvivalBlockBreakHandler {
	h := &SurvivalBlockBreakHandler{
		player:            p,
		blockPos:          blockPos,
		block:             blk,
		targetedFace:      targetedFace,
		maxPlayerDistance: maxPlayerDistance,
		fxTickInterval:    fxTickInterval,
	}
	h.breakSpeed = h.calculateBreakProgressPerTick()
	if h.breakSpeed > 0 {
		h.player.GetWorld().BroadcastPacketToViewers(h.blockPos, levelEvent(packet.LevelEventStartBlockCracking, int32(65535*h.breakSpeed), h.blockPos))
	}
	return h
}

// calculateBreakProgressPerTick is a port of
// SurvivalBlockBreakHandler::calculateBreakProgressPerTick: the percentage of the block's break
// time that one tick of breaking takes.
func (h *SurvivalBlockBreakHandler) calculateBreakProgressPerTick() float64 {
	if !h.block.GetBreakInfo().IsBreakable() {
		return 0.0
	}
	breakTime, err := h.block.GetBreakInfo().GetBreakTime(h.player.GetInventory().GetItemInHand())
	if err != nil {
		return 0.0
	}
	breakTimePerTick := breakTime * 20
	if !h.player.IsOnGround() && !h.player.IsFlying() {
		breakTimePerTick *= 5
	}
	if h.player.IsUnderwater() && !h.player.GetArmorInventory().GetHelmet().HasEnchantment(enchantment.VanillaAquaAffinity(), -1) {
		breakTimePerTick *= 5
	}
	if breakTimePerTick > 0 {
		progressPerTick := 1 / breakTimePerTick

		if haste := h.player.GetEffects().Get(effect.VanillaHaste()); haste != nil {
			hasteLevel := float64(haste.GetEffectLevel())
			progressPerTick *= (1 + 0.2*hasteLevel) * stdmath.Pow(1.2, hasteLevel)
		}

		if miningFatigue := h.player.GetEffects().Get(effect.VanillaMiningFatigue()); miningFatigue != nil {
			miningFatigueLevel := float64(miningFatigue.GetEffectLevel())
			progressPerTick *= stdmath.Pow(0.21, miningFatigueLevel)
		}

		return progressPerTick
	}
	return 1
}

// Update is a port of SurvivalBlockBreakHandler::update: false once the player is too far away
// or the block is broken.
func (h *SurvivalBlockBreakHandler) Update() bool {
	if h.player.GetPosition().DistanceSquared(h.blockPos.Add(0.5, 0.5, 0.5)) > float64(h.maxPlayerDistance*h.maxPlayerDistance) {
		return false
	}

	newBreakSpeed := h.calculateBreakProgressPerTick()
	if stdmath.Abs(newBreakSpeed-h.breakSpeed) > 0.0001 {
		h.breakSpeed = newBreakSpeed
		h.player.GetWorld().BroadcastPacketToViewers(h.blockPos, levelEvent(packet.LevelEventUpdateBlockCracking, int32(65535*h.breakSpeed), h.blockPos))
	}

	h.breakProgress += h.breakSpeed

	fx := h.fxTicker % h.fxTickInterval
	h.fxTicker++
	if fx == 0 && h.breakProgress < 1 {
		w := h.player.GetWorld()
		w.AddParticle(h.blockPos, particle.BlockPunchParticle{BlockStateID: h.block.GetStateId(), Face: h.targetedFace})
		w.AddSound(h.blockPos, sound.BlockPunchSound{BlockStateID: h.block.GetStateId()})
		h.player.BroadcastAnimation(animation.ArmSwingAnimation{Entity: h.player}, h.player.GetViewers())
	}

	return h.breakProgress < 1
}

func (h *SurvivalBlockBreakHandler) GetBlockPos() math.Vector3 { return h.blockPos }

func (h *SurvivalBlockBreakHandler) GetTargetedFace() math.Facing { return h.targetedFace }

func (h *SurvivalBlockBreakHandler) SetTargetedFace(face math.Facing) {
	math.ValidateFacing(face)
	h.targetedFace = face
}

func (h *SurvivalBlockBreakHandler) GetBreakSpeed() float64 { return h.breakSpeed }

func (h *SurvivalBlockBreakHandler) GetBreakProgress() float64 { return h.breakProgress }

// Close is SurvivalBlockBreakHandler::__destruct: the client's crack overlay is stopped.
func (h *SurvivalBlockBreakHandler) Close() {
	if h.player.GetWorld().IsInLoadedTerrain(h.blockPos) {
		h.player.GetWorld().BroadcastPacketToViewers(h.blockPos, levelEvent(packet.LevelEventStopBlockCracking, 0, h.blockPos))
	}
}

// levelEvent is LevelEventPacket::create.
func levelEvent(eventType int32, eventData int32, pos math.Vector3) packet.Packet {
	return &packet.LevelEvent{
		EventType: eventType,
		Position:  mgl32.Vec3{float32(pos.X), float32(pos.Y), float32(pos.Z)},
		EventData: eventData,
	}
}
