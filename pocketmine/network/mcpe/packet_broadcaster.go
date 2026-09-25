package mcpe

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

	"pocketmine-go/pocketmine/event"
	serverevent "pocketmine-go/pocketmine/event/server"
	"pocketmine-go/pocketmine/timings"
)

// PacketBroadcaster is a port of pocketmine\network\mcpe\PacketBroadcaster.
type PacketBroadcaster interface {
	BroadcastPackets(recipients []*NetworkSession, packets []packet.Packet)
}

// StandardPacketBroadcaster is a port of pocketmine\network\mcpe\StandardPacketBroadcaster.
// PHP encodes the packets once and shares a compressed batch between the recipients;
// gophertunnel encodes and batches per connection, so each recipient just queues the packets
// (NetworkSession::addToSendBuffer) and gophertunnel flushes them together.
type StandardPacketBroadcaster struct{}

func NewStandardPacketBroadcaster() *StandardPacketBroadcaster { return &StandardPacketBroadcaster{} }

func (b *StandardPacketBroadcaster) BroadcastPackets(recipients []*NetworkSession, packets []packet.Packet) {
	//TODO: this shouldn't really be called here, since the broadcaster might be replaced by an alternative
	//implementation that doesn't fire events
	if event.HasHandlers[serverevent.DataPacketSendEvent]() {
		targets := make([]serverevent.NetworkSession, len(recipients))
		for i, r := range recipients {
			targets[i] = r
		}
		ev := serverevent.NewDataPacketSendEvent(targets, packets)
		event.Call(ev)
		if ev.IsCancelled() {
			return
		}
		packets = ev.GetPackets()
	}
	for _, target := range recipients {
		for _, pk := range packets {
			target.addToSendBuffer(pk)
		}
	}
}

// BroadcastPackets is a port of NetworkBroadcastUtils::broadcastPackets: sends packets to every
// connected recipient (players or sessions) through their session's broadcaster. It reports
// whether there were any connected recipients.
func BroadcastPackets(recipients []*NetworkSession, packets []packet.Packet) bool {
	if len(packets) == 0 {
		panic("Cannot broadcast empty list of packets")
	}
	timings.Init()
	timings.BroadcastPackets.StartTiming()
	defer timings.BroadcastPackets.StopTiming()

	var sessions []*NetworkSession
	for _, s := range recipients {
		if s != nil && s.IsConnected() {
			sessions = append(sessions, s)
		}
	}
	if len(sessions) == 0 {
		return false
	}

	var uniqueBroadcasters []PacketBroadcaster
	targets := map[PacketBroadcaster][]*NetworkSession{}
	for _, s := range sessions {
		b := s.GetBroadcaster()
		if _, ok := targets[b]; !ok {
			uniqueBroadcasters = append(uniqueBroadcasters, b)
		}
		targets[b] = append(targets[b], s)
	}
	for _, b := range uniqueBroadcasters {
		b.BroadcastPackets(targets[b], packets)
	}
	return true
}
