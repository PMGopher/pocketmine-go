package network

import (
	"encoding/base64"
	"fmt"
	stdmath "math"
	"sync"
	"time"

	"pocketmine-go/pocketmine/event"
	serverevent "pocketmine-go/pocketmine/event/server"
	"pocketmine-go/pocketmine/log"
)

// Network is a port of pocketmine\network\Network: network-related classes.
type Network struct {
	logger log.Logger

	// mu guards the raw packet handlers and banned addresses, which the interfaces read from
	// their own goroutines.
	mu                 sync.RWMutex
	interfaces         []NetworkInterface
	advancedInterfaces []AdvancedNetworkInterface
	rawPacketHandlers  []RawPacketHandler
	bannedIps          map[string]int64

	bandwidthTracker *BidirectionalBandwidthStatsTracker
	name             string
	sessionManager   *NetworkSessionManager
}

func NewNetwork(logger log.Logger) *Network {
	return &Network{
		logger:           logger,
		bannedIps:        map[string]int64{},
		bandwidthTracker: NewBidirectionalBandwidthStatsTracker(5),
		sessionManager:   NewNetworkSessionManager(),
	}
}

func (n *Network) GetBandwidthTracker() *BidirectionalBandwidthStatsTracker {
	return n.bandwidthTracker
}

func (n *Network) GetInterfaces() []NetworkInterface {
	return append([]NetworkInterface(nil), n.interfaces...)
}

func (n *Network) GetSessionManager() *NetworkSessionManager { return n.sessionManager }

func (n *Network) GetConnectionCount() int { return n.sessionManager.GetSessionCount() }

func (n *Network) GetValidConnectionCount() int { return n.sessionManager.GetValidSessionCount() }

func (n *Network) Tick() {
	for _, iface := range n.GetInterfaces() {
		iface.Tick()
	}
	n.sessionManager.Tick()
}

// RegisterInterface is a port of Network::registerInterface: returns false if the interface was
// not registered because a plugin cancelled NetworkInterfaceRegisterEvent, and an error (PHP's
// NetworkInterfaceStartException) if it failed to start.
func (n *Network) RegisterInterface(iface NetworkInterface) (bool, error) {
	ev := serverevent.NewNetworkInterfaceRegisterEvent(iface)
	event.Call(ev)
	if ev.IsCancelled() {
		return false, nil
	}
	if err := iface.Start(); err != nil {
		return false, err
	}
	n.mu.Lock()
	n.interfaces = append(n.interfaces, iface)
	advanced, isAdvanced := iface.(AdvancedNetworkInterface)
	var banned []string
	var handlers []RawPacketHandler
	if isAdvanced {
		n.advancedInterfaces = append(n.advancedInterfaces, advanced)
		for ip := range n.bannedIps {
			banned = append(banned, ip)
		}
		handlers = append(handlers, n.rawPacketHandlers...)
	}
	n.mu.Unlock()
	if isAdvanced {
		advanced.SetNetwork(n)
		for _, ip := range banned {
			advanced.BlockAddress(ip, 300)
		}
		for _, h := range handlers {
			advanced.AddRawPacketFilter(h.Matches)
		}
	}
	iface.SetName(n.name)
	return true, nil
}

// UnregisterInterface is a port of Network::unregisterInterface.
func (n *Network) UnregisterInterface(iface NetworkInterface) error {
	n.mu.Lock()
	index := -1
	for i, existing := range n.interfaces {
		if existing == iface {
			index = i
			break
		}
	}
	if index == -1 {
		n.mu.Unlock()
		return fmt.Errorf("interface %T is not registered on this network", iface)
	}
	n.interfaces = append(n.interfaces[:index], n.interfaces[index+1:]...)
	for i, existing := range n.advancedInterfaces {
		if NetworkInterface(existing) == iface {
			n.advancedInterfaces = append(n.advancedInterfaces[:i], n.advancedInterfaces[i+1:]...)
			break
		}
	}
	n.mu.Unlock()
	event.Call(serverevent.NewNetworkInterfaceUnregisterEvent(iface))
	iface.Shutdown()
	return nil
}

// SetName sets the server name shown on each interface Query.
func (n *Network) SetName(name string) {
	n.name = name
	n.UpdateName()
}

func (n *Network) GetName() string { return n.name }

func (n *Network) UpdateName() {
	for _, iface := range n.GetInterfaces() {
		iface.SetName(n.name)
	}
}

func (n *Network) advanced() []AdvancedNetworkInterface {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return append([]AdvancedNetworkInterface(nil), n.advancedInterfaces...)
}

// SendPacket is a port of Network::sendPacket: sends a raw packet through every advanced
// interface.
func (n *Network) SendPacket(address string, port int, payload []byte) {
	for _, iface := range n.advanced() {
		iface.SendRawPacket(address, port, payload)
	}
}

// BlockAddress blocks an IP address from the main interface for timeout seconds (<= 0 means
// forever).
func (n *Network) BlockAddress(address string, timeout int) {
	n.mu.Lock()
	if timeout > 0 {
		n.bannedIps[address] = time.Now().Unix() + int64(timeout)
	} else {
		n.bannedIps[address] = stdmath.MaxInt64
	}
	n.mu.Unlock()
	for _, iface := range n.advanced() {
		iface.BlockAddress(address, timeout)
	}
}

func (n *Network) UnblockAddress(address string) {
	n.mu.Lock()
	delete(n.bannedIps, address)
	n.mu.Unlock()
	for _, iface := range n.advanced() {
		iface.UnblockAddress(address)
	}
}

// RegisterRawPacketHandler registers a raw packet handler on the network.
func (n *Network) RegisterRawPacketHandler(handler RawPacketHandler) {
	n.mu.Lock()
	n.rawPacketHandlers = append(n.rawPacketHandlers, handler)
	n.mu.Unlock()
	for _, iface := range n.advanced() {
		iface.AddRawPacketFilter(handler.Matches)
	}
}

// UnregisterRawPacketHandler unregisters a previously-registered raw packet handler.
func (n *Network) UnregisterRawPacketHandler(handler RawPacketHandler) {
	n.mu.Lock()
	defer n.mu.Unlock()
	for i, h := range n.rawPacketHandlers {
		if h == handler {
			n.rawPacketHandlers = append(n.rawPacketHandlers[:i], n.rawPacketHandlers[i+1:]...)
			return
		}
	}
}

// ProcessRawPacket is a port of Network::processRawPacket, called by an interface for a packet
// one of the raw packet filters matched.
func (n *Network) ProcessRawPacket(iface AdvancedNetworkInterface, address string, port int, packet []byte) {
	n.mu.RLock()
	until, banned := n.bannedIps[address]
	handlers := append([]RawPacketHandler(nil), n.rawPacketHandlers...)
	n.mu.RUnlock()
	if banned && time.Now().Unix() < until {
		n.logger.Debug(fmt.Sprintf("Dropped raw packet from banned address %s %d", address, port))
		return
	}
	handled := false
	for _, handler := range handlers {
		if !handler.Matches(packet) {
			continue
		}
		ok, err := handler.Handle(iface, address, port, packet)
		if err != nil {
			handled = true
			n.logger.Error(fmt.Sprintf("Bad raw packet from /%s:%d: %v", address, port, err))
			n.BlockAddress(address, 600)
			break
		}
		handled = ok
	}
	if !handled {
		n.logger.Debug(fmt.Sprintf("Unhandled raw packet from /%s:%d: %s", address, port, base64.StdEncoding.EncodeToString(packet)))
	}
}
