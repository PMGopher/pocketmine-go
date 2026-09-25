package raklib

import (
	stdmath "math"
	"net"
	"sync"
	"time"
)

// filteredPacketConn is the part of RakLib's socket handling that RakLibInterface configures: it
// drops datagrams from blocked addresses (blockAddress), hands datagrams matching a raw packet
// filter (addRawPacketFilter, e.g. the query protocol) to RakLibInterface::onRawPacketReceive
// instead of RakNet, and counts the bytes sent and received (onBandwidthStatsUpdate). Everything
// else goes to go-raknet.
type filteredPacketConn struct {
	net.PacketConn
	onRawPacket func(address string, port int, payload []byte)

	mu      sync.RWMutex
	blocked map[string]time.Time
	filters []func(packet []byte) bool

	bandwidthMu   sync.Mutex
	bytesSent     int
	bytesReceived int
}

func newFilteredPacketConn(conn net.PacketConn, onRawPacket func(address string, port int, payload []byte)) *filteredPacketConn {
	return &filteredPacketConn{PacketConn: conn, onRawPacket: onRawPacket, blocked: map[string]time.Time{}}
}

// blockAddress blocks address for timeout seconds (-1 or less: forever), like RakLib's
// Server::blockAddress.
func (c *filteredPacketConn) blockAddress(address string, timeout int) {
	until := time.Now().Add(time.Duration(timeout) * time.Second)
	if timeout < 0 {
		until = time.Unix(stdmath.MaxInt64/2, 0)
	}
	c.mu.Lock()
	c.blocked[address] = until
	c.mu.Unlock()
}

func (c *filteredPacketConn) unblockAddress(address string) {
	c.mu.Lock()
	delete(c.blocked, address)
	c.mu.Unlock()
}

func (c *filteredPacketConn) addRawPacketFilter(filter func(packet []byte) bool) {
	c.mu.Lock()
	c.filters = append(c.filters, filter)
	c.mu.Unlock()
}

func (c *filteredPacketConn) isBlocked(address string) bool {
	c.mu.RLock()
	until, ok := c.blocked[address]
	c.mu.RUnlock()
	if !ok {
		return false
	}
	if time.Now().After(until) {
		c.unblockAddress(address)
		return false
	}
	return true
}

func (c *filteredPacketConn) matchesRawFilter(payload []byte) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, filter := range c.filters {
		if filter(payload) {
			return true
		}
	}
	return false
}

// ReadFrom returns the next datagram meant for RakNet.
func (c *filteredPacketConn) ReadFrom(b []byte) (int, net.Addr, error) {
	for {
		n, addr, err := c.PacketConn.ReadFrom(b)
		if err != nil {
			return n, addr, err
		}
		c.bandwidthMu.Lock()
		c.bytesReceived += n
		c.bandwidthMu.Unlock()

		udpAddr, ok := addr.(*net.UDPAddr)
		if !ok {
			return n, addr, nil
		}
		ip := udpAddr.IP.String()
		if c.isBlocked(ip) {
			continue
		}
		if c.matchesRawFilter(b[:n]) {
			payload := append([]byte(nil), b[:n]...)
			c.onRawPacket(ip, udpAddr.Port, payload)
			continue
		}
		return n, addr, nil
	}
}

func (c *filteredPacketConn) WriteTo(b []byte, addr net.Addr) (int, error) {
	n, err := c.PacketConn.WriteTo(b, addr)
	c.bandwidthMu.Lock()
	c.bytesSent += n
	c.bandwidthMu.Unlock()
	return n, err
}

// takeBandwidth returns the bytes sent and received since the last call.
func (c *filteredPacketConn) takeBandwidth() (sent, received int) {
	c.bandwidthMu.Lock()
	defer c.bandwidthMu.Unlock()
	sent, received = c.bytesSent, c.bytesReceived
	c.bytesSent, c.bytesReceived = 0, 0
	return sent, received
}
