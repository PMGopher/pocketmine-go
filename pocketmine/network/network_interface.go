// Package network is a port of pocketmine\network: the registry of network interfaces (the
// gophertunnel-backed Bedrock listener, the query interface, UPnP), raw packet handling, address
// blocking, the session manager and bandwidth statistics.
package network

import "errors"

// NetworkInterface is a port of pocketmine\network\NetworkInterface: network interfaces are
// transport layers which can be used to transmit packets between the server and clients.
type NetworkInterface interface {
	// Start performs actions needed to start the interface after it is registered.
	Start() error
	// SetName sets the server name (MOTD) shown to clients in the server list.
	SetName(name string)
	// Tick is called every tick to process events on the interface.
	Tick()
	// Shutdown shuts the interface down.
	Shutdown()
}

// AdvancedNetworkInterface is a port of pocketmine\network\AdvancedNetworkInterface: an interface
// with raw packet access, address blocking and a back-reference to the Network.
//
// PHP's addRawPacketFilter takes a regex matched against the raw bytes; Go's regexp works on UTF-8
// text, not bytes, so a filter here is a function matching the raw packet instead.
type AdvancedNetworkInterface interface {
	NetworkInterface
	// BlockAddress prevents packets received from the IP address getting processed for the given
	// timeout in seconds (-1 or less means forever).
	BlockAddress(address string, timeout int)
	// UnblockAddress unblocks a previously-blocked address.
	UnblockAddress(address string)
	SetNetwork(network *Network)
	// SendRawPacket sends a raw payload to the network interface, bypassing any sessions.
	SendRawPacket(address string, port int, payload []byte)
	// AddRawPacketFilter adds a filter: packets it matches are passed to Network.ProcessRawPacket.
	AddRawPacketFilter(filter func(packet []byte) bool)
}

// RawPacketHandler is a port of pocketmine\network\RawPacketHandler.
type RawPacketHandler interface {
	// Matches is RawPacketHandler::getPattern: whether packet is one this handler wants.
	Matches(packet []byte) bool
	// Handle returns whether the packet was handled, or an error (PHP's PacketHandlingException)
	// if it was malformed.
	Handle(iface AdvancedNetworkInterface, address string, port int, packet []byte) (bool, error)
}

// ErrFilterNoisyPacket is a port of pocketmine\network\FilterNoisyPacketException: returned by a
// packet handler to have identical following packets dropped without processing.
var ErrFilterNoisyPacket = errors.New("noisy packet")

// PacketHandlingError is a port of pocketmine\network\PacketHandlingException.
type PacketHandlingError struct {
	Message string
	Cause   error
}

func (e *PacketHandlingError) Error() string { return e.Message }
func (e *PacketHandlingError) Unwrap() error { return e.Cause }

// WrapPacketHandlingError is PacketHandlingException::wrap.
func WrapPacketHandlingError(previous error, prefix string) *PacketHandlingError {
	msg := previous.Error()
	if prefix != "" {
		msg = prefix + ": " + msg
	}
	return &PacketHandlingError{Message: msg, Cause: previous}
}

// NetworkInterfaceStartError is a port of pocketmine\network\NetworkInterfaceStartException.
type NetworkInterfaceStartError struct {
	Message string
	Cause   error
}

func (e *NetworkInterfaceStartError) Error() string { return e.Message }
func (e *NetworkInterfaceStartError) Unwrap() error { return e.Cause }
