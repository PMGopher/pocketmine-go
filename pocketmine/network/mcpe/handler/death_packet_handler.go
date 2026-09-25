package handler

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/network/mcpe"
	"pocketmine-go/pocketmine/player"
)

// DeathPacketHandler is a port of pocketmine\network\mcpe\handler\DeathPacketHandler: handles
// packets while the player is dead, until it respawns.
type DeathPacketHandler struct {
	player           *player.Player
	session          *mcpe.NetworkSession
	inventoryManager *mcpe.InventoryManager
	deathMessage     any
}

func NewDeathPacketHandler(p *player.Player, session *mcpe.NetworkSession, inventoryManager *mcpe.InventoryManager, deathMessage any) *DeathPacketHandler {
	return &DeathPacketHandler{player: p, session: session, inventoryManager: inventoryManager, deathMessage: deathMessage}
}

func (h *DeathPacketHandler) respawnPacket(state byte) *packet.Respawn {
	pos := h.player.GetOffsetPosition(h.player.GetSpawn().Vector3)
	return &packet.Respawn{
		Position:        mgl32.Vec3{float32(pos.X), float32(pos.Y), float32(pos.Z)},
		State:           state,
		EntityRuntimeID: uint64(h.player.GetID()),
	}
}

// SetUp is a port of DeathPacketHandler::setUp.
func (h *DeathPacketHandler) SetUp() {
	h.session.SendDataPacket(h.respawnPacket(packet.RespawnStateSearchingForSpawn))

	var parameters []string
	var message string
	if t, ok := h.deathMessage.(*lang.Translatable); ok {
		if !h.session.GetServer().IsLanguageForced() {
			message, parameters = h.session.PrepareClientTranslatableMessage(t)
		} else {
			message = h.player.GetLanguage().Translate(t)
		}
	} else if h.deathMessage != nil {
		message = fmt.Sprint(h.deathMessage)
	}
	h.session.SendDataPacket(&packet.DeathInfo{Cause: message, Messages: parameters})
}

// CanHandle reports the packets DeathPacketHandler handles (PacketHandlerInspector).
func (h *DeathPacketHandler) CanHandle(pk packet.Packet) bool {
	switch pk.(type) {
	case *packet.PlayerAction, *packet.ContainerClose, *packet.Respawn:
		return true
	}
	return false
}

// HandleDataPacket dispatches pk to the matching handleX method.
func (h *DeathPacketHandler) HandleDataPacket(pk packet.Packet) (bool, error) {
	switch pk := pk.(type) {
	case *packet.PlayerAction:
		return h.handlePlayerAction(pk), nil
	case *packet.ContainerClose:
		h.inventoryManager.OnClientRemoveWindow(int(pk.WindowID))
		return true, nil
	case *packet.Respawn:
		return h.handleRespawn(pk), nil
	}
	return false, nil
}

// handlePlayerAction is a port of DeathPacketHandler::handlePlayerAction.
func (h *DeathPacketHandler) handlePlayerAction(pk *packet.PlayerAction) bool {
	if pk.ActionType == protocol.PlayerActionRespawn {
		h.player.Respawn()
		return true
	}
	return false
}

// handleRespawn is a port of DeathPacketHandler::handleRespawn.
func (h *DeathPacketHandler) handleRespawn(pk *packet.Respawn) bool {
	if pk.State == packet.RespawnStateClientReadyToSpawn {
		h.session.SendDataPacket(h.respawnPacket(packet.RespawnStateReadyToSpawn))
		return true
	}
	return false
}
