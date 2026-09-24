package mcpe

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// PacketHandler is a port of pocketmine\network\mcpe\handler\PacketHandler: the object a
// NetworkSession currently hands received packets to. The concrete handlers live in the handler
// package; NetworkSession creates them through NewPreSpawnPacketHandler/NewInGamePacketHandler,
// which that package sets in its init() (the handlers need NetworkSession, so this package can't
// import them directly).
type PacketHandler interface {
	// SetUp is a port of PacketHandler::setUp, called when the handler becomes the session's
	// handler.
	SetUp() error
	// HandleDataPacket dispatches pk to the matching handleX method (PHP's $packet->handle($this))
	// and reports whether it was handled.
	HandleDataPacket(pk packet.Packet) bool
}

// Handler constructors, set by the handler package's init().
var (
	NewPreSpawnPacketHandler func(session *NetworkSession) PacketHandler
	NewInGamePacketHandler   func(session *NetworkSession) PacketHandler
)
