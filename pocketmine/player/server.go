package player

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
)

// Server is what Player needs from pocketmine\Server (PHP's $this->server). It's declared here so
// the server package, which imports this one, can satisfy it without an import cycle.
type Server interface {
	// BroadcastMessage is a port of Server::broadcastMessage. message is a string or
	// *lang.Translatable; nil recipients means every online player.
	BroadcastMessage(message any, recipients []*Player) int
	// DispatchCommand is a port of Server::dispatchCommand for a player-typed command line
	// (without the leading slash).
	DispatchCommand(sender *Player, commandLine string) bool
	// GetTick is a port of Server::getTick.
	GetTick() int64
	// GetOnlinePlayers is a port of Server::getOnlinePlayers.
	GetOnlinePlayers() []*Player
	// RemoveOnlinePlayer is a port of Server::removeOnlinePlayer.
	RemoveOnlinePlayer(p *Player)
	// SaveOfflinePlayerData is a port of Server::saveOfflinePlayerData.
	SaveOfflinePlayerData(name string, data *nbt.CompoundTag)
}

// NetworkSession is what Player needs from pocketmine\network\mcpe\NetworkSession (PHP's
// $this->networkSession).
type NetworkSession interface {
	SendDataPacket(pk packet.Packet)
	// OnChatMessage is a port of NetworkSession::onChatMessage (message is a string or
	// *lang.Translatable).
	OnChatMessage(message any)
	// SyncViewAreaCenterPoint is a port of NetworkSession::syncViewAreaCenterPoint.
	SyncViewAreaCenterPoint(pos math.Vector3, viewDistance int)
}

// SetServer sets the server this player belongs to (the $server constructor argument in PHP).
func (p *Player) SetServer(server Server) { p.server = server }

// GetServer is a port of Player::getServer.
func (p *Player) GetServer() Server { return p.server }

// SetNetworkSession connects the player to its session (the $session constructor argument in PHP).
// Packets sent to the player go to the session from then on.
func (p *Player) SetNetworkSession(session NetworkSession) {
	p.networkSession = session
	p.packetSender = session.SendDataPacket
}

// GetNetworkSession is a port of Player::getNetworkSession.
func (p *Player) GetNetworkSession() NetworkSession { return p.networkSession }
