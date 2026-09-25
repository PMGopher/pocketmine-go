// Package server is a port of pocketmine\event\server: server-wide events (commands, packets,
// network interfaces, query, memory).
//
// Importers conventionally alias this package as serverevent.
package server

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

	"pocketmine-go/pocketmine/command"
	"pocketmine-go/pocketmine/event"
	"pocketmine-go/pocketmine/utils"
)

// NetworkSession is the surface the packet events need from pocketmine\network\mcpe\NetworkSession.
type NetworkSession interface {
	GetIp() string
	GetPort() int
	GetDisplayName() string
}

// NetworkInterface is the surface the network interface events need from
// pocketmine\network\NetworkInterface.
type NetworkInterface interface {
	Start() error
	Shutdown()
}

// CommandEvent is a port of pocketmine\event\server\CommandEvent: called when any CommandSender
// runs a command, before it is parsed. This can be used for logging commands, or preprocessing
// the command string to add custom features (e.g. selectors).
//
// WARNING: DO NOT use this to block commands. Many commands have aliases. For example, /version
// can also be invoked using /ver or /about. To prevent command senders from using certain
// commands, deny them access to the permissions of those commands.
type CommandEvent struct {
	event.CancellableTrait

	sender  command.Sender
	command string
}

func NewCommandEvent(sender command.Sender, commandLine string) *CommandEvent {
	return &CommandEvent{sender: sender, command: commandLine}
}

func (e *CommandEvent) GetSender() command.Sender { return e.sender }

func (e *CommandEvent) GetCommand() string { return e.command }

func (e *CommandEvent) SetCommand(commandLine string) { e.command = commandLine }

// DataPacketDecodeEvent is a port of pocketmine\event\server\DataPacketDecodeEvent: called
// before a packet is decoded and handled by the network session. Cancelling this event will drop
// the packet without decoding it, minimizing wasted CPU time.
type DataPacketDecodeEvent struct {
	event.CancellableTrait

	origin       NetworkSession
	packetID     uint32
	packetBuffer []byte
}

func NewDataPacketDecodeEvent(origin NetworkSession, packetID uint32, packetBuffer []byte) *DataPacketDecodeEvent {
	return &DataPacketDecodeEvent{origin: origin, packetID: packetID, packetBuffer: packetBuffer}
}

func (e *DataPacketDecodeEvent) GetOrigin() NetworkSession { return e.origin }
func (e *DataPacketDecodeEvent) GetPacketId() uint32       { return e.packetID }
func (e *DataPacketDecodeEvent) GetPacketBuffer() []byte   { return e.packetBuffer }

// DataPacketReceiveEvent is a port of pocketmine\event\server\DataPacketReceiveEvent.
type DataPacketReceiveEvent struct {
	event.CancellableTrait

	origin NetworkSession
	packet packet.Packet
}

func NewDataPacketReceiveEvent(origin NetworkSession, pk packet.Packet) *DataPacketReceiveEvent {
	return &DataPacketReceiveEvent{origin: origin, packet: pk}
}

func (e *DataPacketReceiveEvent) GetPacket() packet.Packet  { return e.packet }
func (e *DataPacketReceiveEvent) GetOrigin() NetworkSession { return e.origin }

// DataPacketSendEvent is a port of pocketmine\event\server\DataPacketSendEvent: called when
// packets are sent to network sessions.
type DataPacketSendEvent struct {
	event.CancellableTrait

	targets []NetworkSession
	packets []packet.Packet
}

func NewDataPacketSendEvent(targets []NetworkSession, packets []packet.Packet) *DataPacketSendEvent {
	return &DataPacketSendEvent{targets: targets, packets: packets}
}

func (e *DataPacketSendEvent) GetTargets() []NetworkSession { return e.targets }

func (e *DataPacketSendEvent) GetPackets() []packet.Packet { return e.packets }

func (e *DataPacketSendEvent) SetPackets(packets []packet.Packet) { e.packets = packets }

// LowMemoryEvent is a port of pocketmine\event\server\LowMemoryEvent: called when the server is
// in a low-memory state as defined by the properties. Plugins should free caches or other
// non-essential data.
type LowMemoryEvent struct {
	memory       uint64
	memoryLimit  uint64
	isGlobal     bool
	triggerCount int
}

func NewLowMemoryEvent(memory, memoryLimit uint64, isGlobal bool, triggerCount int) *LowMemoryEvent {
	return &LowMemoryEvent{memory: memory, memoryLimit: memoryLimit, isGlobal: isGlobal, triggerCount: triggerCount}
}

// GetMemory returns the memory usage at the time of the event call (in bytes).
func (e *LowMemoryEvent) GetMemory() uint64 { return e.memory }

// GetMemoryLimit returns the memory limit defined (in bytes).
func (e *LowMemoryEvent) GetMemoryLimit() uint64 { return e.memoryLimit }

// GetTriggerCount returns the times this event has been called in the current low-memory state.
func (e *LowMemoryEvent) GetTriggerCount() int { return e.triggerCount }

func (e *LowMemoryEvent) IsGlobal() bool { return e.isGlobal }

// GetMemoryFreed returns the amount of memory already freed.
func (e *LowMemoryEvent) GetMemoryFreed() int64 {
	_, rss, virt := utils.AdvancedMemoryUsage()
	usage := rss
	if e.isGlobal {
		usage = virt
	}
	return int64(e.memory) - int64(usage)
}

// NetworkInterfaceEvent is a port of pocketmine\event\server\NetworkInterfaceEvent. It isn't
// abstract, so its handlers also receive the register/unregister events.
type NetworkInterfaceEvent struct {
	iface NetworkInterface
}

func NewNetworkInterfaceEvent(iface NetworkInterface) *NetworkInterfaceEvent {
	return &NetworkInterfaceEvent{iface: iface}
}

func (e *NetworkInterfaceEvent) GetInterface() NetworkInterface { return e.iface }

// NetworkInterfaceRegisterEvent is a port of pocketmine\event\server\NetworkInterfaceRegisterEvent:
// called when a network interface is registered into the network, for example the RakLib
// interface.
type NetworkInterfaceRegisterEvent struct {
	NetworkInterfaceEvent
	event.CancellableTrait
}

func NewNetworkInterfaceRegisterEvent(iface NetworkInterface) *NetworkInterfaceRegisterEvent {
	return &NetworkInterfaceRegisterEvent{NetworkInterfaceEvent: NetworkInterfaceEvent{iface: iface}}
}

// NetworkInterfaceUnregisterEvent is a port of
// pocketmine\event\server\NetworkInterfaceUnregisterEvent: called when a network interface is
// unregistered.
type NetworkInterfaceUnregisterEvent struct {
	NetworkInterfaceEvent
}

func NewNetworkInterfaceUnregisterEvent(iface NetworkInterface) *NetworkInterfaceUnregisterEvent {
	return &NetworkInterfaceUnregisterEvent{NetworkInterfaceEvent: NetworkInterfaceEvent{iface: iface}}
}

func init() {
	event.DeclareParent[NetworkInterfaceRegisterEvent, NetworkInterfaceEvent]()
	event.DeclareParent[NetworkInterfaceUnregisterEvent, NetworkInterfaceEvent]()
}
