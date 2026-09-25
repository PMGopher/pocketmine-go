package mcpe

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

	"pocketmine-go/pocketmine/player"
)

// PacketHandler is a port of pocketmine\network\mcpe\handler\PacketHandler: the object a
// NetworkSession currently hands received packets to. The concrete handlers live in the handler
// package; NetworkSession creates them through the New*PacketHandler variables below, which that
// package sets in its init() (the handlers need NetworkSession, so this package can't import them
// directly).
type PacketHandler interface {
	// SetUp is a port of PacketHandler::setUp, called when the handler becomes the session's
	// handler.
	SetUp()
	// HandleDataPacket dispatches pk to the matching handleX method (PHP's $packet->handle($this))
	// and reports whether it was handled. An error is PHP's PacketHandlingException.
	HandleDataPacket(pk packet.Packet) (bool, error)
}

// PacketFilter is implemented by handlers that know which packets they handle: it's
// PacketHandlerInspector::getHandlerActions, used to tell DataPacketDecodeEvent whether a packet
// would be discarded.
type PacketFilter interface {
	CanHandle(pk packet.Packet) bool
}

// Handler constructors, set by the handler package's init().
var (
	NewPreSpawnPacketHandler func(server Server, p *player.Player, session *NetworkSession, invManager *InventoryManager) PacketHandler
	NewInGamePacketHandler   func(p *player.Player, session *NetworkSession, invManager *InventoryManager) PacketHandler
	NewDeathPacketHandler    func(p *player.Player, session *NetworkSession, invManager *InventoryManager, deathMessage any) PacketHandler
)
