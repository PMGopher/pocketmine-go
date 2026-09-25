// Package raklib is a port of pocketmine\network\mcpe\raklib: the network interface that accepts
// Bedrock clients. RakLib itself (and its thread) is replaced by gophertunnel's listener on top of
// go-raknet (see AGENTS.md §1); this package is the glue PocketMine-MP has between RakLib and
// NetworkSession.
package raklib

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/sandertv/go-raknet"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/sandertv/gophertunnel/minecraft/resource"

	"pocketmine-go/pocketmine/event"
	playerevent "pocketmine-go/pocketmine/event/player"
	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/log"
	"pocketmine-go/pocketmine/network"
	"pocketmine-go/pocketmine/network/mcpe"
	"pocketmine-go/pocketmine/network/query"
	"pocketmine-go/pocketmine/player"
	"pocketmine-go/pocketmine/resourcepacks"
	"pocketmine-go/pocketmine/utils"
)

// Server is what RakLibInterface needs from pocketmine\Server.
type Server interface {
	mcpe.Server
	GetName() string
	GetQueryInformation() *query.QueryInfo
	// GetPropertyInt is ServerConfigGroup::getPropertyInt (pocketmine.yml).
	GetPropertyInt(variable string, defaultValue int) int
	GetResourcePackManager() *resourcepacks.ResourcePackManager
	// ErrorLog is where gophertunnel's and go-raknet's own errors are logged.
	ErrorLog() *slog.Logger
}

// pendingLogin is what the login phase (LoginPacketHandler, run from gophertunnel's Allow
// callback) hands to the NetworkSession created once gophertunnel accepts the connection.
type pendingLogin struct {
	info          player.Info
	authenticated bool
	authRequired  bool
}

var networkIDCounter atomic.Int64

// RakLibInterface is a port of pocketmine\network\mcpe\raklib\RakLibInterface.
//
// Differences from PHP, all because gophertunnel runs the connection sequence itself:
//   - PlayerPreLoginEvent (LoginPacketHandler::processLoginCommon) is fired from gophertunnel's
//     Allow callback, before encryption, like PHP; its authRequired can only tighten
//     authentication, since gophertunnel already refused unauthenticated clients when xbox-auth
//     is on. With xbox-auth off, gophertunnel doesn't verify identity tokens, so every client is
//     treated as unauthenticated (PHP would still accept a verified XUID).
//   - PlayerResourcePackOfferEvent is fired from gophertunnel's FetchResourcePacks callback; its
//     mustAccept flag can't be set per player (resource_stack's force_resources applies).
//   - RakLib's ack receipts don't exist, and ping is gophertunnel's latency estimate.
type RakLibInterface struct {
	server      Server
	logger      log.Logger
	ip          string
	port        int
	ipV6        bool
	network     *network.Network
	rakServerID int64

	packetBroadcaster      mcpe.PacketBroadcaster
	entityEventBroadcaster mcpe.EntityEventBroadcaster

	listener *minecraft.Listener
	conn     *filteredPacketConn

	nameMu sync.Mutex
	name   string

	// sessions are only touched with the server lock held.
	sessions map[*minecraft.Conn]*mcpe.NetworkSession

	pendingMu sync.Mutex
	pending   map[string]*pendingLogin
}

// NewRakLibInterface is a port of RakLibInterface::__construct.
func NewRakLibInterface(server Server, ip string, port int, ipV6 bool, packetBroadcaster mcpe.PacketBroadcaster, entityEventBroadcaster mcpe.EntityEventBroadcaster) *RakLibInterface {
	return &RakLibInterface{
		server:                 server,
		logger:                 server.GetLogger(),
		ip:                     ip,
		port:                   port,
		ipV6:                   ipV6,
		packetBroadcaster:      packetBroadcaster,
		entityEventBroadcaster: entityEventBroadcaster,
		sessions:               map[*minecraft.Conn]*mcpe.NetworkSession{},
		pending:                map[string]*pendingLogin{},
	}
}

// SetNetwork is a port of RakLibInterface::setNetwork.
func (r *RakLibInterface) SetNetwork(n *network.Network) { r.network = n }

// Start is a port of RakLibInterface::start: binds the socket and starts accepting connections.
func (r *RakLibInterface) Start() error {
	r.logger.Debug("Waiting for RakLib to start...")

	netID := "pocketmine-raknet-" + strconv.FormatInt(networkIDCounter.Add(1), 10)
	udpNetwork := "udp4"
	if r.ipV6 {
		udpNetwork = "udp6"
	}
	minecraft.RegisterNetwork(netID, func(l *slog.Logger) minecraft.Network {
		return rakNetwork{iface: r, udpNetwork: udpNetwork, errorLog: l}
	})

	packs := r.server.GetResourcePackManager()
	cfg := minecraft.ListenConfig{
		ErrorLog:               r.server.ErrorLog(),
		AuthenticationDisabled: !r.server.RequiresAuthentication(),
		StatusProvider:         statusProvider{iface: r},
		ResourcePacks:          resourcepacks.NetworkPacks(packs.GetResourceStack(), packEncryptionKeys(packs)),
		TexturePacksRequired:   packs.ResourcePacksRequired(),
		FetchResourcePacks:     r.fetchResourcePacks,
		Allow:                  r.allow,
	}
	host := r.ip
	if r.ipV6 && host == "" {
		host = "::"
	}
	listener, err := cfg.Listen(netID, net.JoinHostPort(host, strconv.Itoa(r.port)))
	if err != nil {
		return &network.NetworkInterfaceStartError{Message: err.Error(), Cause: err}
	}
	r.listener = listener
	r.logger.Debug("RakLib booted successfully")

	go r.acceptLoop()
	return nil
}

// packEncryptionKeys is the pack ID => key map for the resource stack.
func packEncryptionKeys(m *resourcepacks.ResourcePackManager) map[string]string {
	keys := map[string]string{}
	for _, pack := range m.GetResourceStack() {
		if key, ok := m.GetPackEncryptionKey(pack.GetPackId()); ok {
			keys[strings.ToLower(pack.GetPackId())] = key
		}
	}
	return keys
}

// Tick is a port of RakLibInterface::tick. It also does RakLib's per-tick reporting: ping
// measurements (onPingMeasure) and bandwidth statistics (onBandwidthStatsUpdate).
func (r *RakLibInterface) Tick() {
	for conn, session := range r.sessions {
		session.UpdatePing(int(conn.Latency().Milliseconds()))
	}
	if r.conn != nil && r.network != nil {
		sent, received := r.conn.takeBandwidth()
		r.network.GetBandwidthTracker().Add(sent, received)
	}
}

// Shutdown is a port of RakLibInterface::shutdown.
func (r *RakLibInterface) Shutdown() {
	if r.listener != nil {
		_ = r.listener.Close()
	}
}

// SetName is a port of RakLibInterface::setName: the server list entry. gophertunnel assembles the
// "MCPE;..." pong string itself from statusProvider.
func (r *RakLibInterface) SetName(name string) {
	r.nameMu.Lock()
	r.name = name
	r.nameMu.Unlock()
}

// statusProvider serves the server list entry SetName describes.
type statusProvider struct{ iface *RakLibInterface }

func (p statusProvider) ServerStatus(playerCount, maxPlayers int) minecraft.ServerStatus {
	p.iface.nameMu.Lock()
	name := p.iface.name
	p.iface.nameMu.Unlock()
	status := minecraft.ServerStatus{ServerName: name, ServerSubName: p.iface.server.GetName(), PlayerCount: playerCount, MaxPlayers: maxPlayers}
	if info := p.iface.server.GetQueryInformation(); info != nil {
		status.PlayerCount = info.GetPlayerCount()
		status.MaxPlayers = info.GetMaxPlayerCount()
	}
	return status
}

// BlockAddress is a port of RakLibInterface::blockAddress.
func (r *RakLibInterface) BlockAddress(address string, timeout int) {
	if r.conn != nil {
		r.conn.blockAddress(address, timeout)
	}
}

// UnblockAddress is a port of RakLibInterface::unblockAddress.
func (r *RakLibInterface) UnblockAddress(address string) {
	if r.conn != nil {
		r.conn.unblockAddress(address)
	}
}

// SendRawPacket is a port of RakLibInterface::sendRawPacket.
func (r *RakLibInterface) SendRawPacket(address string, port int, payload []byte) {
	if r.conn == nil {
		return
	}
	addr, err := net.ResolveUDPAddr("udp", net.JoinHostPort(address, strconv.Itoa(port)))
	if err != nil {
		return
	}
	_, _ = r.conn.WriteTo(payload, addr)
}

// AddRawPacketFilter is a port of RakLibInterface::addRawPacketFilter.
func (r *RakLibInterface) AddRawPacketFilter(filter func(packet []byte) bool) {
	if r.conn != nil {
		r.conn.addRawPacketFilter(filter)
	}
}

// onRawPacketReceive is a port of RakLibInterface::onRawPacketReceive.
func (r *RakLibInterface) onRawPacketReceive(address string, port int, payload []byte) {
	if r.network == nil {
		return
	}
	r.server.Lock()
	defer r.server.Unlock()
	r.network.ProcessRawPacket(r, address, port, payload)
}

func loginKey(identity login.IdentityData) string {
	return identity.Identity + "\x00" + identity.DisplayName
}

func splitAddr(addr net.Addr) (string, int) {
	host, portString, err := net.SplitHostPort(addr.String())
	if err != nil {
		return addr.String(), 0
	}
	port, _ := strconv.Atoi(portString)
	return host, port
}

// allow is the login phase of LoginPacketHandler::handleLogin + processLoginCommon, run by
// gophertunnel once it has parsed (and, with xbox-auth on, verified) the login request.
func (r *RakLibInterface) allow(addr net.Addr, identity login.IdentityData, clientData login.ClientData) (string, bool) {
	r.server.Lock()
	defer r.server.Unlock()
	language := r.server.GetLanguage()
	translate := func(message any) string {
		if t, ok := message.(*lang.Translatable); ok {
			return language.Translate(t)
		}
		return fmt.Sprint(message)
	}

	ip, port := splitAddr(addr)
	// The connection has no NetworkSession yet (it's created once gophertunnel accepts it, and
	// logs "Session opened" then). A connection refused here is logged like a session would log
	// it.
	logger := log.NewPrefixedLogger(r.logger, "NetworkSession: "+ip+" "+strconv.Itoa(port))
	logRefused := func(username string, reason any) {
		logger.Info(language.Translate(lang.KnownTranslationFactory.PocketmineNetworkSessionOpen()))
		if username != "" {
			logger.Info(language.Translate(lang.KnownTranslationFactory.PocketmineNetworkSessionPlayerName(utils.Aqua + username + utils.Reset)))
			logger.SetPrefix("NetworkSession: " + username)
		}
		logger.Info(language.Translate(lang.KnownTranslationFactory.PocketmineNetworkSessionClose(reason)))
	}

	authenticated := r.server.RequiresAuthentication() && identity.XUID != ""
	info, reason, screenMessage := mcpe.NewPlayerInfoFromLogin(identity, clientData, authenticated)
	if info == nil {
		if screenMessage == nil {
			screenMessage = reason
		}
		logRefused("", reason)
		return translate(screenMessage), false
	}

	ev := playerevent.NewPlayerPreLoginEvent(info, ip, port, r.server.RequiresAuthentication())
	// The connecting client isn't counted by the session manager yet (PHP has marked its session
	// as logged in by now).
	if r.network.GetValidConnectionCount()+1 > r.server.GetMaxPlayers() {
		ev.SetKickFlag(playerevent.KickFlagServerFull, lang.KnownTranslationFactory.DisconnectionScreenServerFull(), nil)
	}
	if !r.server.IsWhitelisted(info.GetUsername()) {
		ev.SetKickFlag(playerevent.KickFlagServerWhitelisted, lang.KnownTranslationFactory.PocketmineDisconnectWhitelisted(), nil)
	}

	var banMessage any
	if banEntry := r.server.GetNameBans().GetEntry(info.GetUsername()); banEntry != nil {
		if banReason := banEntry.Reason(); banReason == "" {
			banMessage = lang.KnownTranslationFactory.PocketmineDisconnectBanNoReason()
		} else {
			banMessage = lang.KnownTranslationFactory.PocketmineDisconnectBan(banReason)
		}
	} else if banEntry := r.server.GetIPBans().GetEntry(ip); banEntry != nil {
		var banReason any = banEntry.Reason()
		if banReason == "" {
			banReason = lang.KnownTranslationFactory.PocketmineDisconnectBanIp()
		}
		banMessage = lang.KnownTranslationFactory.PocketmineDisconnectBan(banReason)
	}
	if banMessage != nil {
		ev.SetKickFlag(playerevent.KickFlagBanned, banMessage, nil)
	}

	event.Call(ev)
	if !ev.IsAllowed() {
		logRefused(info.GetUsername(), ev.GetFinalDisconnectReason())
		return translate(ev.GetFinalDisconnectScreenMessage()), false
	}

	r.pendingMu.Lock()
	r.pending[loginKey(identity)] = &pendingLogin{info: info, authenticated: authenticated, authRequired: ev.IsAuthRequired()}
	r.pendingMu.Unlock()
	return "", true
}

// fetchResourcePacks is NetworkSession::onServerLoginSuccess's PlayerResourcePackOfferEvent.
func (r *RakLibInterface) fetchResourcePacks(identity login.IdentityData, _ login.ClientData, current []*resource.Pack) []*resource.Pack {
	r.pendingMu.Lock()
	pending := r.pending[loginKey(identity)]
	r.pendingMu.Unlock()
	if pending == nil {
		return current
	}

	r.server.Lock()
	defer r.server.Unlock()
	packManager := r.server.GetResourcePackManager()
	stack := packManager.GetResourceStack()
	packs := make([]playerevent.ResourcePack, len(stack))
	for i, p := range stack {
		packs[i] = p
	}
	ev := playerevent.NewPlayerResourcePackOfferEvent(pending.info, packs, packEncryptionKeys(packManager), packManager.ResourcePacksRequired())
	event.Call(ev)

	offered := make([]resourcepacks.ResourcePack, 0, len(ev.GetResourcePacks()))
	for _, p := range ev.GetResourcePacks() {
		if rp, ok := p.(resourcepacks.ResourcePack); ok {
			offered = append(offered, rp)
		}
	}
	return resourcepacks.NetworkPacks(offered, ev.GetEncryptionKeys())
}

// acceptLoop hands every accepted connection to onClientConnect.
func (r *RakLibInterface) acceptLoop() {
	for {
		c, err := r.listener.Accept()
		if err != nil {
			return
		}
		go r.onClientConnect(c.(*minecraft.Conn))
	}
}

// onClientConnect is a port of RakLibInterface::onClientConnect, followed by the connection's
// packet loop (onPacketReceive) until it closes (onClientDisconnect).
func (r *RakLibInterface) onClientConnect(conn *minecraft.Conn) {
	ip, port := splitAddr(conn.RemoteAddr())

	r.pendingMu.Lock()
	pending := r.pending[loginKey(conn.IdentityData())]
	delete(r.pending, loginKey(conn.IdentityData()))
	r.pendingMu.Unlock()

	r.server.Lock()
	session := mcpe.NewNetworkSession(r.server, r.network.GetSessionManager(), conn, conn, r.packetBroadcaster, r.entityEventBroadcaster, ip, port)
	r.sessions[conn] = session
	if pending == nil {
		session.DisconnectWithError("Login data missing", lang.KnownTranslationFactory.PocketmineDisconnectErrorInternal())
	} else {
		session.Login(pending.info, pending.authenticated, pending.authRequired)
	}
	r.server.Unlock()

	for {
		pk, err := conn.ReadPacket()
		if err != nil {
			r.server.Lock()
			r.onClientDisconnect(conn, err)
			r.server.Unlock()
			return
		}
		r.server.Lock()
		r.onPacketReceive(conn, session, pk)
		r.server.Unlock()
		// Send the responses now instead of waiting for gophertunnel's flush timer (up to 50 ms)
		// or the end of the tick: RakLib sends a session's queued packets as soon as they're
		// handed over.
		_ = conn.Flush()
	}
}

// onClientDisconnect is a port of RakLibInterface::onClientDisconnect.
func (r *RakLibInterface) onClientDisconnect(conn *minecraft.Conn, err error) {
	session, ok := r.sessions[conn]
	if !ok {
		return
	}
	delete(r.sessions, conn)
	var reason any = lang.KnownTranslationFactory.PocketmineDisconnectClientDisconnect()
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		reason = lang.KnownTranslationFactory.PocketmineDisconnectErrorTimeout()
	}
	session.OnClientDisconnect(reason)
}

// onPacketReceive is a port of RakLibInterface::onPacketReceive.
func (r *RakLibInterface) onPacketReceive(conn *minecraft.Conn, session *mcpe.NetworkSession, pk packet.Packet) {
	if _, ok := r.sessions[conn]; !ok {
		return
	}
	//get this now for blocking in case the player was closed before the exception was raised
	address := session.GetIp()
	name := session.GetDisplayName()
	defer func() {
		if p := recover(); p != nil {
			//record the name of the player who caused the crash, to make it easier to find the reproducing steps
			r.logger.Emergency("Crash occurred while handling a packet from session: " + name)
			panic(p)
		}
	}()
	if err := session.HandleDataPacket(pk); err != nil {
		session.DisconnectWithError("Bad packet: "+err.Error(), lang.KnownTranslationFactory.PocketmineDisconnectErrorBadPacket())
		//intentionally doesn't use logException, we don't want spammy packet error traces to appear in release mode
		session.GetLogger().Debug(err.Error())

		r.BlockAddress(address, 5)
	}
}

// rakNetwork is the minecraft.Network gophertunnel listens on: go-raknet on top of
// filteredPacketConn.
type rakNetwork struct {
	iface      *RakLibInterface
	udpNetwork string
	errorLog   *slog.Logger
}

func (n rakNetwork) DialContext(ctx context.Context, address string) (net.Conn, error) {
	return nil, errors.New("dialing is not supported")
}

func (n rakNetwork) PingContext(ctx context.Context, address string) ([]byte, error) {
	return nil, errors.New("pinging is not supported")
}

func (n rakNetwork) Listen(address string) (minecraft.NetworkListener, error) {
	cfg := raknet.ListenConfig{
		ErrorLog:               n.errorLog.With("net origin", "raknet"),
		UpstreamPacketListener: packetListener{iface: n.iface, udpNetwork: n.udpNetwork},
		MaxMTU:                 uint16(n.iface.server.GetPropertyInt("network.max-mtu-size", 1492)),
	}
	l, err := cfg.Listen(address)
	if err != nil {
		return nil, err
	}
	// go-raknet picks its own random server ID ($rakServerId).
	n.iface.rakServerID = l.ID()
	return l, nil
}

// packetListener opens the UDP socket wrapped in filteredPacketConn.
type packetListener struct {
	iface      *RakLibInterface
	udpNetwork string
}

func (p packetListener) ListenPacket(_, address string) (net.PacketConn, error) {
	conn, err := net.ListenPacket(p.udpNetwork, address)
	if err != nil {
		return nil, err
	}
	filtered := newFilteredPacketConn(conn, p.iface.onRawPacketReceive)
	p.iface.conn = filtered
	return filtered, nil
}
