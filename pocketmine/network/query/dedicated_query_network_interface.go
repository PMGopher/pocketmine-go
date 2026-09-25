package query

import (
	"errors"
	"fmt"
	stdmath "math"
	"net"
	"strconv"
	"sync"
	"time"

	"pocketmine-go/pocketmine/log"
	"pocketmine-go/pocketmine/network"
)

// DedicatedQueryNetworkInterface is a port of
// pocketmine\network\query\DedicatedQueryNetworkInterface: a UDP socket of its own for query,
// used when the game interface (which normally carries query packets) isn't registered.
type DedicatedQueryNetworkInterface struct {
	ip     string
	port   int
	ipV6   bool
	logger log.Logger

	conn    net.PacketConn
	network *network.Network

	mu         sync.Mutex
	blockedIps map[string]int64
	filters    []func([]byte) bool
	packets    chan rawPacket
	done       chan struct{}
}

type rawPacket struct {
	addr *net.UDPAddr
	data []byte
}

func NewDedicatedQueryNetworkInterface(ip string, port int, ipV6 bool, logger log.Logger) *DedicatedQueryNetworkInterface {
	return &DedicatedQueryNetworkInterface{ip: ip, port: port, ipV6: ipV6, logger: logger, blockedIps: map[string]int64{}, packets: make(chan rawPacket, 1024), done: make(chan struct{})}
}

func (d *DedicatedQueryNetworkInterface) Start() error {
	netw := "udp4"
	if d.ipV6 {
		netw = "udp6"
	}
	conn, err := net.ListenPacket(netw, net.JoinHostPort(d.ip, strconv.Itoa(d.port)))
	if err != nil {
		return &network.NetworkInterfaceStartError{Message: fmt.Sprintf("Failed to bind to %s %d: %v", d.ip, d.port, err), Cause: err}
	}
	d.conn = conn
	go d.read()
	d.logger.Info(fmt.Sprintf("Running on %s %d", d.ip, d.port))
	return nil
}

func (d *DedicatedQueryNetworkInterface) read() {
	buf := make([]byte, 65535)
	for {
		n, addr, err := d.conn.ReadFrom(buf)
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}
			d.logger.Debug("Failed to recv: " + err.Error())
			continue
		}
		udpAddr, ok := addr.(*net.UDPAddr)
		if !ok {
			continue
		}
		select {
		case d.packets <- rawPacket{addr: udpAddr, data: append([]byte(nil), buf[:n]...)}:
		case <-d.done:
			return
		}
	}
}

func (d *DedicatedQueryNetworkInterface) SetName(name string) {
	//NOOP
}

// Tick processes the packets received since the last tick on the main thread.
func (d *DedicatedQueryNetworkInterface) Tick() {
	for {
		select {
		case p := <-d.packets:
			address := p.addr.IP.String()
			d.mu.Lock()
			until, blocked := d.blockedIps[address]
			filters := append([]func([]byte) bool(nil), d.filters...)
			d.mu.Unlock()
			if blocked && until > time.Now().Unix() {
				d.logger.Debug("Dropped packet from banned address " + address)
				continue
			}
			for _, filter := range filters {
				if filter(p.data) {
					d.network.ProcessRawPacket(d, address, p.addr.Port, p.data)
					break
				}
			}
		default:
			return
		}
	}
}

func (d *DedicatedQueryNetworkInterface) BlockAddress(address string, timeout int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if timeout > 0 {
		d.blockedIps[address] = time.Now().Unix() + int64(timeout)
	} else {
		d.blockedIps[address] = stdmath.MaxInt64
	}
}

func (d *DedicatedQueryNetworkInterface) UnblockAddress(address string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.blockedIps, address)
}

func (d *DedicatedQueryNetworkInterface) SetNetwork(n *network.Network) { d.network = n }

func (d *DedicatedQueryNetworkInterface) SendRawPacket(address string, port int, payload []byte) {
	addr := &net.UDPAddr{IP: net.ParseIP(address), Port: port}
	if _, err := d.conn.WriteTo(payload, addr); err != nil {
		d.logger.Debug(fmt.Sprintf("Failed to send to %s %d: %v", address, port, err))
	}
}

func (d *DedicatedQueryNetworkInterface) AddRawPacketFilter(filter func(packet []byte) bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.filters = append(d.filters, filter)
}

func (d *DedicatedQueryNetworkInterface) Shutdown() {
	close(d.done)
	if d.conn != nil {
		_ = d.conn.Close()
	}
}

var _ network.AdvancedNetworkInterface = (*DedicatedQueryNetworkInterface)(nil)
