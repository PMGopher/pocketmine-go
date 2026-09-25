package handler

import (
	stdmath "math"
	"strconv"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/network/mcpe"
	"pocketmine-go/pocketmine/player"
	"pocketmine-go/pocketmine/world"
)

// InGamePacketHandler is a port of pocketmine\network\mcpe\handler\InGamePacketHandler: handles
// the packets a spawned player sends.
//
// Not ported yet (these packets are reported as unhandled at debug level): item use and inventory
// transactions other than attacking/interacting with entities (placing blocks, using items),
// ItemStackRequest execution (requests are answered with an error, so the client rolls them back),
// containers, book editing, forms, commands and the other handleX methods that depend on systems
// this port doesn't have yet.
type InGamePacketHandler struct {
	session *mcpe.NetworkSession
	player  *player.Player

	lastPlayerAuthInputPosition *mgl32.Vec3
	lastPlayerAuthInputYaw      float32
	lastPlayerAuthInputPitch    float32
	lastPlayerAuthInputFlags    *protocol.InputFlags
	lastBlockAttacked           *protocol.BlockPos
}

func NewInGamePacketHandler(session *mcpe.NetworkSession) *InGamePacketHandler {
	return &InGamePacketHandler{session: session, player: session.GetPlayer(), lastPlayerAuthInputYaw: float32(stdmath.NaN()), lastPlayerAuthInputPitch: float32(stdmath.NaN())}
}

// SetUp is a port of InGamePacketHandler's (empty) setUp.
func (h *InGamePacketHandler) SetUp() error { return nil }

// HandleDataPacket dispatches pk to the handleX method for its type.
func (h *InGamePacketHandler) HandleDataPacket(pk packet.Packet) bool {
	switch pk := pk.(type) {
	case *packet.Text:
		return h.handleText(pk)
	case *packet.PlayerAuthInput:
		return h.handlePlayerAuthInput(pk)
	case *packet.SubChunkRequest:
		h.session.SendDataPacket(mcpe.HandleSubChunkRequest(h.player.GetWorld(), pk, h.session.GetBlobCache()))
		return true
	case *packet.ClientCacheBlobStatus:
		// No InGamePacketHandler counterpart: see mcpe.ClientBlobCache.
		if cache := h.session.GetBlobCache(); cache != nil {
			if resp := cache.HandleBlobStatus(pk); resp != nil {
				h.session.SendDataPacket(resp)
			}
			return true
		}
		return false
	case *packet.InventoryTransaction:
		return h.handleInventoryTransaction(pk)
	case *packet.ItemStackRequest:
		return h.handleItemStackRequest(pk)
	case *packet.MobEquipment:
		return h.handleMobEquipment(pk)
	case *packet.RequestAbility:
		return h.handleRequestAbility(pk)
	}
	return false
}

// handleText is a port of InGamePacketHandler::handleText.
func (h *InGamePacketHandler) handleText(pk *packet.Text) bool {
	if pk.TextType == packet.TextTypeChat {
		return h.player.Chat(pk.Message)
	}
	return false
}

// inputFlagsEqual is PHP's BitSet::equals for gophertunnel's InputFlags.
func inputFlagsEqual(a, b protocol.InputFlags) bool {
	if a.Len() != b.Len() {
		return false
	}
	for i := 0; i < a.Len(); i++ {
		if a.Load(i) != b.Load(i) {
			return false
		}
	}
	return true
}

func itoa(i int) string { return strconv.Itoa(i) }

// resolveOnOffInputFlags is a port of InGamePacketHandler::resolveOnOffInputFlags: nil when
// neither flag was set, or both were set.
func resolveOnOffInputFlags(inputFlags protocol.InputFlags, startFlag, stopFlag int) *bool {
	enabled := inputFlags.Load(startFlag)
	disabled := inputFlags.Load(stopFlag)
	if enabled != disabled {
		return &enabled
	}
	return nil
}

// handlePlayerAuthInput is a port of InGamePacketHandler::handlePlayerAuthInput. Not ported: the
// forceMoveSync teleport acknowledgement (teleports aren't ported), item interaction data inside
// the packet (item use isn't ported) and ItemStackRequest execution (answered with an error).
func (h *InGamePacketHandler) handlePlayerAuthInput(pk *packet.PlayerAuthInput) bool {
	rawPos := pk.Position
	rawYaw, rawPitch := pk.Yaw, pk.Pitch
	for _, f := range []float32{rawPos[0], rawPos[1], rawPos[2], rawYaw, pk.HeadYaw, rawPitch} {
		if stdmath.IsInf(float64(f), 0) || stdmath.IsNaN(float64(f)) {
			h.session.GetLogger().Debug("Invalid movement received, contains NAN/INF components")
			return false
		}
	}

	if rawYaw != h.lastPlayerAuthInputYaw || rawPitch != h.lastPlayerAuthInputPitch {
		h.lastPlayerAuthInputYaw, h.lastPlayerAuthInputPitch = rawYaw, rawPitch

		yaw := stdmath.Mod(float64(rawYaw), 360)
		pitch := stdmath.Mod(float64(rawPitch), 360)
		if yaw < 0 {
			yaw += 360
		}
		h.player.SetRotation(yaw, pitch)
	}

	hasMoved := h.lastPlayerAuthInputPosition == nil || *h.lastPlayerAuthInputPosition != rawPos
	newPos := math.NewVector3(float64(rawPos[0]), float64(rawPos[1])-1.62, float64(rawPos[2])).Round(4)

	inputFlags := pk.InputData
	if h.lastPlayerAuthInputFlags == nil || !inputFlagsEqual(*h.lastPlayerAuthInputFlags, inputFlags) {
		flags := inputFlags
		h.lastPlayerAuthInputFlags = &flags

		sneakPressed := inputFlags.Load(packet.InputFlagSneaking)

		sneaking := resolveOnOffInputFlags(inputFlags, packet.InputFlagStartSneaking, packet.InputFlagStopSneaking)
		sprinting := resolveOnOffInputFlags(inputFlags, packet.InputFlagStartSprinting, packet.InputFlagStopSprinting)
		swimming := resolveOnOffInputFlags(inputFlags, packet.InputFlagStartSwimming, packet.InputFlagStopSwimming)
		gliding := resolveOnOffInputFlags(inputFlags, packet.InputFlagStartGliding, packet.InputFlagStopGliding)
		flying := resolveOnOffInputFlags(inputFlags, packet.InputFlagStartFlying, packet.InputFlagStopFlying)

		sneak := h.player.IsSneaking()
		if sneaking != nil {
			sneak = *sneaking
		}
		// PHP evaluates every toggle (bitwise |), so every one of them runs.
		mismatch := !h.player.ToggleSneak(sneak, sneakPressed)
		if sprinting != nil && !h.player.ToggleSprint(*sprinting) {
			mismatch = true
		}
		if swimming != nil && !h.player.ToggleSwim(*swimming) {
			mismatch = true
		}
		if gliding != nil && !h.player.ToggleGlide(*gliding) {
			mismatch = true
		}
		if flying != nil && !h.player.ToggleFlight(*flying) {
			mismatch = true
		}
		if mismatch {
			h.player.SendData([]world.EntityViewer{h.player}, nil)
			// The client already applied a flight toggle it predicted; PHP resyncs the abilities
			// from Player::setFlying, which isn't reached when flight is refused.
			if flying != nil {
				h.session.SyncAbilities(h.player)
			}
		}

		if inputFlags.Load(packet.InputFlagStartJumping) {
			h.player.Jump()
		}
		if inputFlags.Load(packet.InputFlagMissedSwing) {
			h.player.MissSwing()
		}
	}

	if hasMoved {
		pos := rawPos
		h.lastPlayerAuthInputPosition = &pos
		//TODO: this packet has WAYYYYY more useful information that we're not using
		if !h.player.HandleMovement(newPos) {
			h.revertMovement()
		}
	}

	packetHandled := true

	if _, ok := pk.ItemInteractionData.Value(); ok {
		packetHandled = false
		h.session.GetLogger().Debug("Unhandled transaction in PlayerAuthInputPacket (item use isn't ported)")
	}

	if request, ok := pk.ItemStackRequest.Value(); ok {
		h.session.SendDataPacket(&packet.ItemStackResponse{Responses: []protocol.ItemStackResponse{{Status: protocol.ItemStackResponseStatusError, RequestID: request.RequestID}}})
	}

	//itemstack request or transaction may set predictions for the outcome of these actions, so these need to be
	//processed last
	if blockActions, ok := pk.BlockActions.Value(); ok {
		if len(blockActions) > 100 {
			h.session.GetLogger().Debug("Too many block actions in PlayerAuthInputPacket")
			return false
		}
		for k, blockAction := range blockActions {
			if !h.handlePlayerActionFromData(blockAction.Action, blockAction.BlockPos, blockAction.Face) {
				packetHandled = false
				h.session.GetLogger().Debug("Unhandled player block action at offset " + itoa(k) + " in PlayerAuthInputPacket")
			}
		}
	}

	return packetHandled
}

// revertMovement is a port of Player::revertMovement's network half
// (NetworkSession::syncMovement with MovePlayerPacket::MODE_RESET).
func (h *InGamePacketHandler) revertMovement() {
	location := h.player.GetLocation()
	eye := h.player.GetOffsetPosition(location.Vector3)
	h.session.SendDataPacket(&packet.MovePlayer{
		EntityRuntimeID: uint64(h.player.GetID()),
		Position:        mgl32.Vec3{float32(eye.X), float32(eye.Y), float32(eye.Z)},
		Pitch:           float32(location.Pitch),
		Yaw:             float32(location.Yaw),
		HeadYaw:         float32(location.Yaw),
		Mode:            packet.MoveModeReset,
		OnGround:        h.player.IsOnGround(),
	})
}

// handlePlayerActionFromData is a port of InGamePacketHandler::handlePlayerActionFromData.
func (h *InGamePacketHandler) handlePlayerActionFromData(action int32, blockPosition protocol.BlockPos, face int32) bool {
	pos := math.NewVector3(float64(blockPosition[0]), float64(blockPosition[1]), float64(blockPosition[2]))
	facing := math.Facing(face)

	switch action {
	case protocol.PlayerActionStartBreak, protocol.PlayerActionContinueDestroyBlock: //destroy the next block while holding down left click
		if !validFacing(face) {
			return false
		}
		if h.lastBlockAttacked != nil && *h.lastBlockAttacked == blockPosition {
			//the client will send CONTINUE_DESTROY_BLOCK for the currently targeted block directly before it
			//sends PREDICT_DESTROY_BLOCK, but also when it starts to break the block
			//this seems like a bug in the client and would cause spurious left-click events if we allowed it to
			//be delivered to the player
			h.session.GetLogger().Debug("Ignoring PlayerAction " + itoa(int(action)) + " because we were already destroying this block")
			break
		}
		if !h.player.AttackBlock(pos, facing, h.player.GetInventory().GetItemInHand()) {
			h.syncBlocksNearby(pos, facing)
		}
		attacked := blockPosition
		h.lastBlockAttacked = &attacked

	case protocol.PlayerActionAbortBreak, protocol.PlayerActionStopBreak:
		h.player.StopBreakBlock(pos)
		h.lastBlockAttacked = nil
	case protocol.PlayerActionStartSleeping:
		//unused
	case protocol.PlayerActionStopSleeping:
		// Player::stopSleep: sleeping isn't ported.
	case protocol.PlayerActionCrackBreak:
		if !validFacing(face) {
			return false
		}
		h.player.ContinueBreakBlock(pos, facing)
		attacked := blockPosition
		h.lastBlockAttacked = &attacked
	case protocol.PlayerActionInteractWithBlock: //TODO: ignored (for now)
	case protocol.PlayerActionCreativePlayerDestroyBlock:
		//in server auth block breaking, we get PREDICT_DESTROY_BLOCK anyway, so this action is redundant
	case protocol.PlayerActionPredictDestroyBlock:
		if !validFacing(face) {
			return false
		}
		if !h.player.BreakBlock(pos) {
			h.syncBlocksNearby(pos, facing)
		}
		h.lastBlockAttacked = nil
	case protocol.PlayerActionStartItemUseOn, protocol.PlayerActionStopItemUseOn:
		//TODO: this has no obvious use and seems only used for analytics in vanilla - ignore it
	default:
		h.session.GetLogger().Debug("Unhandled/unknown player action type " + itoa(int(action)))
		return false
	}

	h.player.SetUsingItem(false)

	return true
}

// validFacing is InGamePacketHandler::validateFacing, which throws a PacketHandlingException for
// an invalid face; here the action is rejected instead.
func validFacing(face int32) bool { return face >= 0 && face <= 5 }

// syncBlocksNearby is a port of InGamePacketHandler::syncBlocksNearby: the client predicted a block
// change the server refused, so the blocks around it are resent.
func (h *InGamePacketHandler) syncBlocksNearby(blockPos math.Vector3, face math.Facing) {
	if blockPos.DistanceSquared(h.player.GetLocation().Vector3) >= 10000 {
		return
	}
	blocks := []math.Vector3{blockPos}
	for _, side := range math.AllFacing {
		blocks = append(blocks, blockPos.GetSide(side, 1))
	}
	sidePos := blockPos.GetSide(face, 1)
	blocks = append(blocks, sidePos)
	for _, side := range math.AllFacing {
		blocks = append(blocks, sidePos.GetSide(side, 1))
	}
	for _, pk := range h.player.GetWorld().CreateBlockUpdatePackets(blocks) {
		h.session.SendDataPacket(pk)
	}
}

// handleInventoryTransaction is a port of InGamePacketHandler::handleInventoryTransaction for the
// one transaction type that's ported: using an item on an entity (handleUseItemOnEntityTransaction).
func (h *InGamePacketHandler) handleInventoryTransaction(pk *packet.InventoryTransaction) bool {
	data, ok := pk.TransactionData.(*protocol.UseItemOnEntityTransactionData)
	if !ok {
		return false
	}
	return h.handleUseItemOnEntityTransaction(data)
}

// handleUseItemOnEntityTransaction is a port of
// InGamePacketHandler::handleUseItemOnEntityTransaction. Player::interactEntity isn't ported.
func (h *InGamePacketHandler) handleUseItemOnEntityTransaction(data *protocol.UseItemOnEntityTransactionData) bool {
	target, ok := h.player.GetWorld().GetEntity(int(data.TargetEntityRuntimeID))
	//TODO: HACK! We really shouldn't be keeping disconnected players (and generally flagged-for-despawn entities)
	//in the world's entity table, but changing that is too risky for a hotfix. This workaround will do for now.
	if !ok || target.IsFlaggedForDespawn() {
		return false
	}

	h.player.SelectHotbarSlot(int(data.HotBarSlot))

	switch data.ActionType {
	case protocol.UseItemOnEntityActionAttack:
		h.player.AttackEntity(target)
		return true
	}
	return false
}

// handleItemStackRequest answers every request with an error: ItemStackRequestExecutor isn't
// ported, so no request can be executed and the client rolls its prediction back (PHP answers a
// request that failed to execute the same way).
func (h *InGamePacketHandler) handleItemStackRequest(pk *packet.ItemStackRequest) bool {
	responses := make([]protocol.ItemStackResponse, 0, len(pk.Requests))
	for _, request := range pk.Requests {
		responses = append(responses, protocol.ItemStackResponse{Status: protocol.ItemStackResponseStatusError, RequestID: request.RequestID})
	}
	h.session.SendDataPacket(&packet.ItemStackResponse{Responses: responses})
	return true
}

// handleMobEquipment is a port of InGamePacketHandler::handleMobEquipment.
func (h *InGamePacketHandler) handleMobEquipment(pk *packet.MobEquipment) bool {
	if pk.WindowID == protocol.WindowIDOffHand {
		return true //this happens when we put an item into the offhand
	}
	if pk.WindowID == protocol.WindowIDInventory {
		if !h.player.SelectHotbarSlot(int(pk.HotBarSlot)) {
			h.session.SyncSelectedHotbarSlot()
		}
		return true
	}
	return false
}

// handleRequestAbility has no InGamePacketHandler counterpart: PocketMine-MP 5.44 doesn't handle
// RequestAbility. The client predicts the requested ability (flying, when the swim-up gesture is
// read as a double jump) until told otherwise, so the abilities are resynced to refuse anything
// the player isn't allowed (see commit 697dfa7).
func (h *InGamePacketHandler) handleRequestAbility(pk *packet.RequestAbility) bool {
	if pk.Ability == protocol.AbilityFlying {
		if value, ok := pk.Value.(bool); ok && h.player.ToggleFlight(value) {
			return true
		}
	}
	h.session.SyncAbilities(h.player)
	return true
}
