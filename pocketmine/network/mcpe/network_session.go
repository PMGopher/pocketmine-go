// Package mcpe is a port of pocketmine\network\mcpe: the connection-level code between the
// Bedrock protocol and the game. The protocol itself (packets, RakNet, login, encryption) is
// github.com/sandertv/gophertunnel instead of a port of pmmp/BedrockProtocol and RakLib (see
// AGENTS.md §1), so NetworkSession wraps a gophertunnel *minecraft.Conn, which has already
// completed the login sequence by the time a session is created.
package mcpe

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

	"pocketmine-go/pocketmine/command"
	"pocketmine-go/pocketmine/entity"
	"pocketmine-go/pocketmine/entity/effect"
	"pocketmine-go/pocketmine/event"
	playerevent "pocketmine-go/pocketmine/event/player"
	serverevent "pocketmine-go/pocketmine/event/server"
	"pocketmine-go/pocketmine/form"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/log"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
	"pocketmine-go/pocketmine/network"
	"pocketmine-go/pocketmine/network/mcpe/convert"
	"pocketmine-go/pocketmine/permission"
	"pocketmine-go/pocketmine/player"
	"pocketmine-go/pocketmine/utils"
)

// Server is what the network code needs from pocketmine\Server. It's declared here so the server
// package, which imports this one, satisfies it without an import cycle.
type Server interface {
	// Locker serialises packet handling with the server tick: PocketMine-MP does both on one
	// thread, and the world, entities and players aren't safe for concurrent use.
	sync.Locker
	GetLogger() log.Logger
	GetLanguage() *lang.Language
	IsLanguageForced() bool
	GetMotd() string
	GetMaxPlayers() int
	RequiresAuthentication() bool
	// GetPropertyBool is ServerConfigGroup::getPropertyBool (pocketmine.yml).
	GetPropertyBool(variable string, defaultValue bool) bool
	IsWhitelisted(name string) bool
	GetNameBans() *permission.BanList
	GetIPBans() *permission.BanList
	GetNetwork() *network.Network
	GetCommandMap() command.CommandMap
	// GetOfflinePlayerData is a port of Server::getOfflinePlayerData.
	GetOfflinePlayerData(name string) *nbt.CompoundTag
	// CreatePlayer is a port of Server::createPlayer. World::requestSafeSpawn resolves
	// synchronously in this port (generation is synchronous), so it returns the player directly.
	CreatePlayer(session *NetworkSession, playerInfo player.Info, authenticated bool, offlinePlayerData *nbt.CompoundTag) (*player.Player, error)
	// AddOnlinePlayer is a port of Server::addOnlinePlayer.
	AddOnlinePlayer(p *player.Player) bool
	// GetOnlinePlayers is a port of Server::getOnlinePlayers.
	GetOnlinePlayers() []*player.Player
	// GetAllowedViewDistance is a port of Server::getAllowedViewDistance.
	GetAllowedViewDistance(distance int) int
	// GetGamemode is a port of Server::getGamemode.
	GetGamemode() player.GameMode
}

// PacketSender is a port of pocketmine\network\mcpe\PacketSender: gophertunnel's connection plays
// this role (it batches, compresses, encrypts and flushes packets itself).
type PacketSender interface {
	WritePacket(pk packet.Packet) error
	Close() error
}

// Packet rate limits, NetworkSession::INCOMING_*.
const (
	incomingGamePacketsPerTick     = 2
	incomingGamePacketsBufferTicks = 100
)

// NetworkSession is a port of pocketmine\network\mcpe\NetworkSession.
//
// gophertunnel owns the parts of the PHP class below the packet layer: batching, compression,
// encryption, the raw send buffer and ack receipts (sendDataPacketWithReceipt), and the login
// sequence handlers (SessionStart, Login, Handshake, ResourcePacks). The login-phase game logic
// those handlers call into (PlayerPreLoginEvent, setAuthenticationStatus, the resource pack offer
// event) is done by RakLibInterface and Login.
type NetworkSession struct {
	server  Server
	manager *network.NetworkSessionManager
	sender  PacketSender
	conn    *minecraft.Conn

	broadcaster            PacketBroadcaster
	entityEventBroadcaster EntityEventBroadcaster

	gamePacketLimiter *PacketRateLimiter

	logger *log.PrefixedLogger
	player *player.Player
	info   player.Info
	ping   *int

	handler PacketHandler

	connected               bool
	disconnectGuard         bool
	loggedIn                bool
	authenticated           bool
	connectTime             time.Time
	cachedOfflinePlayerData *nbt.CompoundTag

	invManager *InventoryManager

	disposeHooks []func()

	ip   string
	port int

	// blobCache is non-nil when the client enabled the client blob cache (ClientCacheStatus).
	blobCache *ClientBlobCache
	// playerListSkin is this client's skin as sent in PlayerList (see playerListSkin).
	playerListSkin *protocol.Skin

	// trace is the recent packet history printed when the client disconnects by itself.
	trace packetTrace
}

// NewNetworkSession is a port of NetworkSession::__construct, for a connection gophertunnel has
// accepted. conn may be nil in tests (sender is then the only way packets leave).
func NewNetworkSession(server Server, manager *network.NetworkSessionManager, sender PacketSender, conn *minecraft.Conn, broadcaster PacketBroadcaster, entityEventBroadcaster EntityEventBroadcaster, ip string, port int) *NetworkSession {
	s := &NetworkSession{
		server:                 server,
		manager:                manager,
		sender:                 sender,
		conn:                   conn,
		broadcaster:            broadcaster,
		entityEventBroadcaster: entityEventBroadcaster,
		connected:              true,
		connectTime:            time.Now(),
		ip:                     ip,
		port:                   port,
		gamePacketLimiter:      NewPacketRateLimiter("Game Packets", incomingGamePacketsPerTick, incomingGamePacketsBufferTicks),
	}
	s.logger = log.NewPrefixedLogger(server.GetLogger(), s.getLogPrefix())
	if conn != nil && conn.ClientCacheEnabled() {
		s.blobCache = NewClientBlobCache()
	}

	manager.Add(s)
	s.logger.Info(server.GetLanguage().Translate(lang.KnownTranslationFactory.PocketmineNetworkSessionOpen()))
	return s
}

// getLogPrefix is a port of NetworkSession::getLogPrefix.
func (s *NetworkSession) getLogPrefix() string {
	return "NetworkSession: " + s.GetDisplayName()
}

// GetLogger is a port of NetworkSession::getLogger.
func (s *NetworkSession) GetLogger() log.Logger { return s.logger }

func (s *NetworkSession) GetConn() *minecraft.Conn { return s.conn }
func (s *NetworkSession) GetServer() Server        { return s.server }

// GetPlayer is a port of NetworkSession::getPlayer.
func (s *NetworkSession) GetPlayer() *player.Player { return s.player }

// GetPlayerInfo is a port of NetworkSession::getPlayerInfo (nil before login).
func (s *NetworkSession) GetPlayerInfo() player.Info { return s.info }

// IsConnected is a port of NetworkSession::isConnected.
func (s *NetworkSession) IsConnected() bool { return s.connected && !s.disconnectGuard }

// GetIp is a port of NetworkSession::getIp.
func (s *NetworkSession) GetIp() string { return s.ip }

// GetPort is a port of NetworkSession::getPort.
func (s *NetworkSession) GetPort() int { return s.port }

// GetDisplayName is a port of NetworkSession::getDisplayName.
func (s *NetworkSession) GetDisplayName() string {
	if s.info != nil {
		return s.info.GetUsername()
	}
	return s.ip + " " + strconv.Itoa(s.port)
}

// GetPing is a port of NetworkSession::getPing: the last recorded ping in milliseconds, or nil.
func (s *NetworkSession) GetPing() *int { return s.ping }

// UpdatePing is a port of NetworkSession::updatePing, called by the network interface.
func (s *NetworkSession) UpdatePing(ping int) { s.ping = &ping }

// GetBlobCache returns the session's client blob cache, or nil when the client doesn't use one.
func (s *NetworkSession) GetBlobCache() *ClientBlobCache { return s.blobCache }

// GetHandler is a port of NetworkSession::getHandler.
func (s *NetworkSession) GetHandler() PacketHandler { return s.handler }

// SetHandler is a port of NetworkSession::setHandler.
func (s *NetworkSession) SetHandler(handler PacketHandler) {
	if s.connected { //TODO: this is fine since we can't handle anything from a disconnected session, but it might produce surprises in some cases
		s.handler = handler
		if handler != nil {
			handler.SetUp()
		}
	}
}

// GetBroadcaster is a port of NetworkSession::getBroadcaster.
func (s *NetworkSession) GetBroadcaster() PacketBroadcaster { return s.broadcaster }

// GetEntityEventBroadcaster is a port of NetworkSession::getEntityEventBroadcaster.
func (s *NetworkSession) GetEntityEventBroadcaster() EntityEventBroadcaster {
	return s.entityEventBroadcaster
}

// GetInvManager is a port of NetworkSession::getInvManager (nil until the player is created).
func (s *NetworkSession) GetInvManager() player.InventoryManager {
	if s.invManager == nil {
		return nil
	}
	return s.invManager
}

// GetInventoryManager is GetInvManager as the concrete type (the packet handlers need all of it).
func (s *NetworkSession) GetInventoryManager() *InventoryManager { return s.invManager }

// translate renders a Translatable|string message with the server language.
func (s *NetworkSession) translate(message any) string {
	if t, ok := message.(*lang.Translatable); ok {
		return s.server.GetLanguage().Translate(t)
	}
	if message == nil {
		return ""
	}
	return fmt.Sprint(message)
}

// Login runs the rest of NetworkSession's login phase for a connection gophertunnel has accepted:
// LoginPacketHandler's playerInfoConsumer (info is what RakLibInterface built from the login
// data and passed through PlayerPreLoginEvent), then setAuthenticationStatus. authRequired is
// PlayerPreLoginEvent::isAuthRequired.
func (s *NetworkSession) Login(info player.Info, authenticated, authRequired bool) {
	s.info = info
	s.logger.Info(s.server.GetLanguage().Translate(lang.KnownTranslationFactory.PocketmineNetworkSessionPlayerName(utils.Aqua + info.GetUsername() + utils.Reset)))
	s.logger.SetPrefix(s.getLogPrefix())
	s.manager.MarkLoginReceived(s)

	if s.conn != nil {
		cd := s.conn.ClientData()
		skin := playerListSkin(cd)
		s.playerListSkin = &skin
		s.logger.Debug(fmt.Sprintf("Client skin: persona=%v size=%dx%d geometry=%d bytes cape=%dx%d animations=%d pieces=%d",
			cd.PersonaSkin, cd.SkinImageWidth, cd.SkinImageHeight, len(skin.SkinGeometry), cd.CapeImageWidth, cd.CapeImageHeight, len(cd.AnimatedImageData), len(cd.PersonaPieces)))
		s.logger.Debug(fmt.Sprintf("Client blob cache enabled: %v", s.conn.ClientCacheEnabled()))
	}

	// gophertunnel verified the client's identity and public key; the key is always present.
	s.setAuthenticationStatus(authenticated, authRequired, nil, true)
}

// setAuthenticationStatus is a port of NetworkSession::setAuthenticationStatus. Encryption was
// already enabled by gophertunnel, so onServerLoginSuccess follows directly.
func (s *NetworkSession) setAuthenticationStatus(authenticated, authRequired bool, err any, hasClientPubKey bool) {
	if !s.connected {
		return
	}
	if err == nil {
		if _, xbl := s.info.(*player.XboxLivePlayerInfo); authenticated && !xbl {
			err = "Expected XUID but none found"
		} else if !hasClientPubKey {
			err = "Missing client public key" //failsafe
		}
	}

	if err != nil {
		s.DisconnectWithError(
			lang.KnownTranslationFactory.PocketmineDisconnectInvalidSession(err),
			lang.KnownTranslationFactory.PocketmineDisconnectErrorAuthentication(),
		)
		return
	}

	s.authenticated = authenticated

	if !s.authenticated {
		if authRequired {
			s.Disconnect("Not authenticated", lang.KnownTranslationFactory.DisconnectionScreenNotAuthenticated())
			return
		}
		if xbl, ok := s.info.(*player.XboxLivePlayerInfo); ok {
			s.logger.Warning("Discarding unexpected XUID for non-authenticated player")
			s.info = xbl.WithoutXboxData()
		}
	}
	yes := "NO"
	if s.authenticated {
		yes = "YES"
	}
	s.logger.Debug("Xbox Live authenticated: " + yes)

	checkXUID := s.server.GetPropertyBool("player.verify-xuid", true)
	myXUID := ""
	if xbl, ok := s.info.(*player.XboxLivePlayerInfo); ok {
		myXUID = xbl.GetXuid()
	}
	kickForXUIDMismatch := func(xuid string) bool {
		if checkXUID && myXUID != xuid {
			s.logger.Debug(fmt.Sprintf("XUID mismatch: expected '%s', but got '%s'", xuid, myXUID))
			//TODO: Longer term, we should be identifying playerdata using something more reliable, like XUID or UUID.
			//However, that would be a very disruptive change, so this will serve as a stopgap for now.
			//Side note: this will also prevent offline players hijacking XBL playerdata on online servers, since their
			//XUID will always be empty.
			s.Disconnect("XUID does not match (possible impersonation attempt)", nil)
			return true
		}
		return false
	}

	for _, existing := range s.manager.GetSessions() {
		existingSession, ok := existing.(*NetworkSession)
		if !ok || existingSession == s {
			continue
		}
		info := existingSession.GetPlayerInfo()
		if info != nil && (strings.EqualFold(info.GetUsername(), s.info.GetUsername()) || info.GetUUID() == s.info.GetUUID()) {
			xuid := ""
			if xbl, ok := info.(*player.XboxLivePlayerInfo); ok {
				xuid = xbl.GetXuid()
			}
			if kickForXUIDMismatch(xuid) {
				return
			}
			ev := playerevent.NewPlayerDuplicateLoginEvent(s, existingSession, lang.KnownTranslationFactory.DisconnectionScreenLoggedinOtherLocation(), nil)
			event.Call(ev)
			if ev.IsCancelled() {
				s.Disconnect(ev.GetDisconnectReason(), ev.GetDisconnectScreenMessage())
				return
			}

			existingSession.Disconnect(ev.GetDisconnectReason(), ev.GetDisconnectScreenMessage())
		}
	}

	//TODO: make player data loading async
	//TODO: we shouldn't be loading player data here at all, but right now we don't have any choice :(
	s.cachedOfflinePlayerData = s.server.GetOfflinePlayerData(s.info.GetUsername())
	if checkXUID {
		var recorded nbt.Tag
		if s.cachedOfflinePlayerData != nil {
			recorded, _ = s.cachedOfflinePlayerData.GetTag(player.TagLastKnownXUID)
		}
		if recordedXUID, ok := recorded.(nbt.StringTag); !ok {
			s.logger.Debug("No previous XUID recorded, no choice but to trust this player")
		} else if !kickForXUIDMismatch(string(recordedXUID)) {
			s.logger.Debug("XUID match")
		}
	}
	if !s.connected {
		return
	}

	// EncryptionContext / PrepareEncryptionTask / HandshakePacketHandler: gophertunnel has already
	// done the encryption handshake.
	s.onServerLoginSuccess()
}

// onServerLoginSuccess is a port of NetworkSession::onServerLoginSuccess. PlayStatus(LOGIN_SUCCESS)
// and the resource pack phase were done by gophertunnel (PlayerResourcePackOfferEvent is fired by
// RakLibInterface when gophertunnel asks for the packs), so the player is created right away.
func (s *NetworkSession) onServerLoginSuccess() {
	s.loggedIn = true
	s.logger.Debug("Initiating resource packs phase")
	s.createPlayer()
}

// createPlayer is a port of NetworkSession::createPlayer.
func (s *NetworkSession) createPlayer() {
	p, err := s.server.CreatePlayer(s, s.info, s.authenticated, s.cachedOfflinePlayerData)
	if err != nil {
		//TODO: this should never actually occur... right?
		s.logger.Error(fmt.Sprintf("Failed to create player: %v", err))
		s.DisconnectWithError("Failed to create player", lang.KnownTranslationFactory.PocketmineDisconnectErrorInternal())
		return
	}
	s.onPlayerCreated(p)
}

// onPlayerCreated is a port of NetworkSession::onPlayerCreated.
func (s *NetworkSession) onPlayerCreated(p *player.Player) {
	if !s.IsConnected() {
		//the remote player might have disconnected before spawn terrain generation was finished
		return
	}
	s.player = p
	if !s.server.AddOnlinePlayer(p) {
		return
	}

	s.invManager = NewInventoryManager(p, s)

	effectManager := p.GetEffects()
	effectAddHook := effect.EffectAddHook(func(instance *effect.EffectInstance, replacesOldEffect bool) {
		s.entityEventBroadcaster.OnEntityEffectAdded([]*NetworkSession{s}, s.player, instance, replacesOldEffect)
	})
	effectRemoveHook := effect.EffectRemoveHook(func(instance *effect.EffectInstance) {
		s.entityEventBroadcaster.OnEntityEffectRemoved([]*NetworkSession{s}, s.player, instance)
	})
	effectManager.GetEffectAddHooks().Add(&effectAddHook)
	effectManager.GetEffectRemoveHooks().Add(&effectRemoveHook)
	s.disposeHooks = append(s.disposeHooks, func() {
		effectManager.GetEffectAddHooks().Remove(&effectAddHook)
		effectManager.GetEffectRemoveHooks().Remove(&effectRemoveHook)
	})

	permissionHooks := p.GetPermissionRecalculationCallbacks()
	permHook := permission.PermissionRecalculationCallback(func(map[string]bool) {
		s.logger.Debug("Syncing available commands and abilities/permissions due to permission recalculation")
		s.SyncAbilities(s.player)
		s.SyncAvailableCommands()
	})
	permissionHooks.Add(&permHook)
	s.disposeHooks = append(s.disposeHooks, func() { permissionHooks.Remove(&permHook) })

	s.beginSpawnSequence()
}

// beginSpawnSequence is a port of NetworkSession::beginSpawnSequence.
func (s *NetworkSession) beginSpawnSequence() {
	s.SetHandler(NewPreSpawnPacketHandler(s.server, s.player, s, s.invManager))
	s.player.SetNoClientPredictions(true) //TODO: HACK: fix client-side falling pre-spawn

	s.logger.Debug("Waiting for chunk radius request")
	// gophertunnel's StartGame (PreSpawnPacketHandler::setUp) has already received the client's
	// RequestChunkRadius by now.
	if s.conn != nil && s.IsConnected() {
		s.OnClientRequestChunkRadius(s.conn.ChunkRadius())
	}
}

// OnClientRequestChunkRadius is PreSpawnPacketHandler::handleRequestChunkRadius, called once
// gophertunnel's StartGame has answered the client's RequestChunkRadius: the player's view
// distance is applied and the spawn terrain is sent right away.
//
// gophertunnel sends PlayStatus(PLAYER_SPAWN) as soon as the client asks for its chunk radius and
// waits for the spawn response before StartGame returns, where PocketMine-MP only sends it after
// Player::$spawnThreshold chunks (notifyTerrainReady). To make up for it, the spawn chunks are sent
// in one go here instead of chunksPerTick per tick, and notifyTerrainReady then completes the
// spawn (the client has already answered it).
func (s *NetworkSession) OnClientRequestChunkRadius(radius int) {
	s.player.SetViewDistance(radius)
	for s.IsConnected() && s.player != nil && !s.player.IsSpawned() {
		before := s.sentChunkCount()
		s.player.DoChunkRequests()
		if s.player.IsSpawned() || !s.IsConnected() {
			return
		}
		if s.sentChunkCount() == before {
			// Every chunk in view was sent, but fewer than the spawn threshold (a small view
			// distance): nothing else would complete the spawn.
			s.NotifyTerrainReady()
			return
		}
	}
}

// sentChunkCount is the number of chunks the player has been sent.
func (s *NetworkSession) sentChunkCount() int {
	n := 0
	for _, status := range s.player.GetUsedChunks() {
		if status == player.UsedChunkStatusSent {
			n++
		}
	}
	return n
}

// NotifyTerrainReady is a port of NetworkSession::notifyTerrainReady. PlayStatus(PLAYER_SPAWN) was
// already sent by gophertunnel, and the client's spawn response (SetLocalPlayerAsInitialised,
// SpawnResponsePacketHandler) was already received, so the in-game phase starts right away.
func (s *NetworkSession) NotifyTerrainReady() {
	s.logger.Debug("Sending spawn notification, waiting for spawn response")
	s.onClientSpawnResponse()
}

// onClientSpawnResponse is a port of NetworkSession::onClientSpawnResponse.
func (s *NetworkSession) onClientSpawnResponse() {
	if s.player.IsSpawned() {
		return
	}
	s.logger.Debug("Received spawn response, entering in-game phase")
	s.player.SetNoClientPredictions(false) //TODO: HACK: we set this during the spawn sequence to prevent the client sending junk movements
	s.player.DoFirstSpawn()
	s.SetHandler(NewInGamePacketHandler(s.player, s, s.invManager))
}

// OnServerDeath is a port of NetworkSession::onServerDeath.
func (s *NetworkSession) OnServerDeath(deathMessage any) {
	if _, inGame := s.handler.(inGameHandler); inGame { //TODO: this is a bad fix for pre-spawn death, this shouldn't be reachable at all at this stage :(
		s.SetHandler(NewDeathPacketHandler(s.player, s, s.invManager, deathMessage))
	}
}

// inGameHandler marks InGamePacketHandler (PHP's instanceof checks).
type inGameHandler interface{ SetForceMoveSync(bool) }

// OnServerRespawn is a port of NetworkSession::onServerRespawn.
func (s *NetworkSession) OnServerRespawn() {
	s.entityEventBroadcaster.SyncAttributes([]*NetworkSession{s}, s.player, s.player.GetAttributeMap().GetAll())
	s.player.SendData(nil, nil)

	s.SyncAbilities(s.player)
	s.invManager.SyncAll()
	s.SetHandler(NewInGamePacketHandler(s.player, s, s.invManager))
}

// HandleDataPacket is a port of NetworkSession::handleDataPacket (plus the per-packet part of
// handleEncoded: the rate limit). gophertunnel has decoded the packet already, so
// DataPacketDecodeEvent gets the packet ID and no buffer.
func (s *NetworkSession) HandleDataPacket(pk packet.Packet) error {
	if !s.connected {
		return nil
	}
	s.trace.record("<-", pk)
	if err := s.gamePacketLimiter.Decrement(1); err != nil {
		return err
	}

	handled := s.handler != nil
	if f, ok := s.handler.(PacketFilter); ok {
		handled = f.CanHandle(pk)
	}
	if event.HasHandlers[serverevent.DataPacketDecodeEvent]() {
		ev := serverevent.NewDataPacketDecodeEvent(s, pk.ID(), nil)
		cancel := !handled
		if cancel {
			ev.Cancel()
		}
		event.Call(ev)
		if cancel && !ev.IsCancelled() {
			//uncancelled by a plugin, let it through to DataPacketReceiveEvent
			handled = true
		} else if !cancel && ev.IsCancelled() {
			//explicitly cancelled by plugin, drop it quietly
			return nil
		}
	}

	if !handled {
		s.unhandledPacketDebug(pk, "Discarded without decoding")
		return nil
	}

	if event.HasHandlers[serverevent.DataPacketReceiveEvent]() {
		ev := serverevent.NewDataPacketReceiveEvent(s, pk)
		event.Call(ev)
		if ev.IsCancelled() {
			return nil
		}
	}
	if s.handler == nil {
		s.unhandledPacketDebug(pk, "Handler rejected")
		return nil
	}
	ok, err := s.handler.HandleDataPacket(pk)
	if err != nil {
		s.unhandledPacketDebug(pk, "Packet processing error")
		return network.WrapPacketHandlingError(err, fmt.Sprintf("Error processing %T", pk))
	}
	if !ok {
		s.unhandledPacketDebug(pk, "Handler rejected")
	}
	return nil
}

// unhandledPacketDebug is a port of NetworkSession::unhandledPacketDebug (the packet is shown
// decoded instead of as base64).
func (s *NetworkSession) unhandledPacketDebug(pk packet.Packet, label string) {
	if violation, ok := pk.(*packet.PacketViolationWarning); ok {
		// Logged as a warning: it's the client telling us which packet of ours it rejected.
		s.logger.Warning(fmt.Sprintf("Client reported a packet violation: type=%d severity=%d packetID=%d context=%q",
			violation.Type, violation.Severity, violation.PacketID, violation.ViolationContext))
		return
	}
	debug := fmt.Sprintf("%+v", pk)
	if len(debug) > 1024 {
		debug = debug[:1024] + fmt.Sprintf(" ... (%d bytes not shown)", len(debug)-1024)
	}
	s.logger.Debug(fmt.Sprintf("%s: %T: %s", label, pk, debug))
}

// SendDataPacket is a port of NetworkSession::sendDataPacket: fires DataPacketSendEvent, then
// queues the packets (gophertunnel flushes them).
func (s *NetworkSession) SendDataPacket(pk packet.Packet) {
	s.sendDataPacket(pk, false)
}

// SendDataPacketImmediate is sendDataPacket($packet, immediate: true).
func (s *NetworkSession) SendDataPacketImmediate(pk packet.Packet) {
	s.sendDataPacket(pk, true)
}

func (s *NetworkSession) sendDataPacket(pk packet.Packet, immediate bool) bool {
	if !s.connected {
		return false
	}
	packets := []packet.Packet{pk}
	if event.HasHandlers[serverevent.DataPacketSendEvent]() {
		ev := serverevent.NewDataPacketSendEvent([]serverevent.NetworkSession{s}, packets)
		event.Call(ev)
		if ev.IsCancelled() {
			return false
		}
		packets = ev.GetPackets()
	}
	for _, evPacket := range packets {
		s.addToSendBuffer(evPacket)
	}
	if immediate {
		s.flushGamePacketQueue()
	}
	return true
}

// addToSendBuffer is a port of NetworkSession::addToSendBuffer.
func (s *NetworkSession) addToSendBuffer(pk packet.Packet) {
	s.trace.record("->", pk)
	if err := s.sender.WritePacket(pk); err != nil {
		s.logger.Debug(fmt.Sprintf("Failed to send %T: %v", pk, err))
	}
}

// flushGamePacketQueue is a port of NetworkSession::flushGamePacketQueue.
func (s *NetworkSession) flushGamePacketQueue() {
	if s.conn != nil {
		_ = s.conn.Flush()
	}
}

// tryDisconnect is a port of NetworkSession::tryDisconnect.
func (s *NetworkSession) tryDisconnect(fn func(), reason any) {
	if s.connected && !s.disconnectGuard {
		s.disconnectGuard = true
		fn()
		s.disconnectGuard = false
		s.flushGamePacketQueue()
		_ = s.sender.Close()
		for _, callback := range s.disposeHooks {
			callback()
		}
		s.disposeHooks = nil
		s.SetHandler(nil)
		s.connected = false

		s.logger.Info(s.server.GetLanguage().Translate(lang.KnownTranslationFactory.PocketmineNetworkSessionClose(reason)))
	}
}

// dispose is a port of NetworkSession::dispose: after the session has been disconnected, it's safe
// to destroy any cycles.
func (s *NetworkSession) dispose() {
	if s.invManager != nil {
		s.invManager.dispose()
		s.invManager = nil
	}
}

// sendDisconnectPacket is a port of NetworkSession::sendDisconnectPacket.
func (s *NetworkSession) sendDisconnectPacket(message any) {
	s.sendDataPacket(&packet.Disconnect{Reason: 0, Message: s.translate(message)}, false)
}

// Disconnect is a port of NetworkSession::disconnect: disconnects the session, destroying the
// associated player (if it exists). reason is shown in the server log; disconnectScreenMessage on
// the player's disconnection screen (nil uses the reason). Both are strings or *lang.Translatable.
func (s *NetworkSession) Disconnect(reason, disconnectScreenMessage any) {
	s.DisconnectNotify(reason, disconnectScreenMessage, true)
}

// DisconnectNotify is NetworkSession::disconnect with its $notify parameter: when false, no
// Disconnect packet is sent to the client.
func (s *NetworkSession) DisconnectNotify(reason, disconnectScreenMessage any, notify bool) {
	s.tryDisconnect(func() {
		if notify {
			if disconnectScreenMessage == nil {
				disconnectScreenMessage = reason
			}
			s.sendDisconnectPacket(disconnectScreenMessage)
		}
		if s.player != nil {
			s.player.OnPostDisconnect(reason, nil)
		}
	}, reason)
}

// DisconnectWithError is a port of NetworkSession::disconnectWithError.
func (s *NetworkSession) DisconnectWithError(reason, disconnectScreenMessage any) {
	var b [6]byte
	_, _ = rand.Read(b[:])
	hexID := hex.EncodeToString(b[:])
	errorID := hexID[0:4] + "-" + hexID[4:8] + "-" + hexID[8:12]

	if disconnectScreenMessage == nil {
		disconnectScreenMessage = reason
	}
	s.Disconnect(
		lang.KnownTranslationFactory.PocketmineDisconnectError(reason, errorID).Prefix(utils.Red),
		lang.KnownTranslationFactory.PocketmineDisconnectError(disconnectScreenMessage, errorID),
	)
}

// Transfer is a port of NetworkSession::transfer: instructs the remote client to connect to a
// different server.
func (s *NetworkSession) Transfer(ip string, port int, reason any) {
	if reason == nil {
		reason = lang.KnownTranslationFactory.PocketmineDisconnectTransfer()
	}
	s.tryDisconnect(func() {
		s.sendDataPacket(&packet.Transfer{Address: ip, Port: uint16(port)}, true)
		if s.player != nil {
			s.player.OnPostDisconnect(reason, nil)
		}
	}, reason)
}

// OnPlayerDestroyed is a port of NetworkSession::onPlayerDestroyed: called by the Player when it
// is closed (for example due to getting kicked).
func (s *NetworkSession) OnPlayerDestroyed(reason, disconnectScreenMessage any) {
	s.tryDisconnect(func() {
		s.sendDisconnectPacket(disconnectScreenMessage)
	}, reason)
}

// OnClientDisconnect is a port of NetworkSession::onClientDisconnect: called by the network
// interface when the client disconnects without server input (timeout or voluntary disconnect).
func (s *NetworkSession) OnClientDisconnect(reason any) {
	if s.connected {
		s.logger.Debug("The client closed the connection. Last packets (-> sent, <- received):" + s.trace.String())
	}
	s.tryDisconnect(func() {
		if s.player != nil {
			s.player.OnPostDisconnect(reason, nil)
		}
	}, reason)
}

// SyncMovement is a port of NetworkSession::syncMovement.
func (s *NetworkSession) SyncMovement(pos math.Vector3, yaw, pitch *float64, mode byte) {
	if s.player == nil {
		return
	}
	location := s.player.GetLocation()
	y, p := location.Yaw, location.Pitch
	if yaw != nil {
		y = *yaw
	}
	if pitch != nil {
		p = *pitch
	}
	offset := s.player.GetOffsetPosition(pos)
	s.SendDataPacket(&packet.MovePlayer{
		EntityRuntimeID: uint64(s.player.GetID()),
		Position:        mgl32.Vec3{float32(offset.X), float32(offset.Y), float32(offset.Z)},
		Pitch:           float32(p),
		Yaw:             float32(y),
		HeadYaw:         float32(y), //TODO: head yaw
		Mode:            mode,
		OnGround:        s.player.IsOnGround(),
		//TODO: riding entity ID, tick
	})

	if h, ok := s.handler.(inGameHandler); ok {
		h.SetForceMoveSync(true)
	}
}

// SyncViewAreaRadius is a port of NetworkSession::syncViewAreaRadius.
func (s *NetworkSession) SyncViewAreaRadius(distance int) {
	s.SendDataPacket(&packet.ChunkRadiusUpdated{ChunkRadius: int32(distance)})
}

// SyncViewAreaCenterPoint is a port of NetworkSession::syncViewAreaCenterPoint.
func (s *NetworkSession) SyncViewAreaCenterPoint(pos math.Vector3, viewDistance int) {
	s.SendDataPacket(&packet.NetworkChunkPublisherUpdate{
		Position: blockPos(pos),
		Radius:   uint32(viewDistance * 16), //blocks, not chunks >.>
	})
}

// SyncPlayerSpawnPoint is a port of NetworkSession::syncPlayerSpawnPoint.
func (s *NetworkSession) SyncPlayerSpawnPoint(newSpawn math.Vector3) {
	pos := blockPos(newSpawn)
	//TODO: respawn causing block position (bed, respawn anchor)
	s.SendDataPacket(&packet.SetSpawnPosition{SpawnType: packet.SpawnTypePlayer, Position: pos, Dimension: packet.DimensionOverworld, SpawnPosition: pos})
}

// SyncWorldSpawnPoint is a port of NetworkSession::syncWorldSpawnPoint.
func (s *NetworkSession) SyncWorldSpawnPoint(newSpawn math.Vector3) {
	s.SendDataPacket(&packet.SetSpawnPosition{SpawnType: packet.SpawnTypeWorld, Position: blockPos(newSpawn), Dimension: packet.DimensionOverworld, SpawnPosition: protocol.BlockPos{int32Min, int32Min, int32Min}})
}

// int32Min is PHP's Limits::INT32_MIN, which SetSpawnPositionPacket::worldSpawn uses for the
// causing block position.
const int32Min = -2147483648

func blockPos(v math.Vector3) protocol.BlockPos {
	return protocol.BlockPos{int32(v.FloorX()), int32(v.FloorY()), int32(v.FloorZ())}
}

// SyncGameMode is a port of NetworkSession::syncGameMode.
func (s *NetworkSession) SyncGameMode(mode player.GameMode, isRollback bool) {
	s.SendDataPacket(&packet.SetPlayerGameType{GameType: convert.CoreGameModeToProtocol(int(mode))})
	if s.player != nil {
		s.SyncAbilities(s.player)
		s.SyncAdventureSettings() //TODO: we might be able to do this with the abilities packet alone
	}
	if !isRollback && s.invManager != nil {
		s.invManager.SyncCreative()
	}
}

// SyncAbilities is a port of NetworkSession::syncAbilities.
func (s *NetworkSession) SyncAbilities(p *player.Player) {
	isOp := p.HasPermission(permission.RootOperator)

	//ALL of these need to be set for the base layer, otherwise the client will cry
	boolAbilities := map[uint32]bool{
		protocol.AbilityMayFly:            p.GetAllowFlight(),
		protocol.AbilityFlying:            p.IsFlying(),
		protocol.AbilityNoClip:            !p.HasBlockCollision(),
		protocol.AbilityOperatorCommands:  isOp,
		protocol.AbilityTeleport:          p.HasPermission(permission.CommandTeleportSelf),
		protocol.AbilityInvulnerable:      p.IsCreative(),
		protocol.AbilityMuted:             false,
		protocol.AbilityWorldBuilder:      false,
		protocol.AbilityInstantBuild:      !p.HasFiniteResources(),
		protocol.AbilityLightning:         false,
		protocol.AbilityBuild:             !p.IsSpectator(),
		protocol.AbilityMine:              !p.IsSpectator(),
		protocol.AbilityDoorsAndSwitches:  !p.IsSpectator(),
		protocol.AbilityOpenContainers:    !p.IsSpectator(),
		protocol.AbilityAttackPlayers:     !p.IsSpectator(),
		protocol.AbilityAttackMobs:        !p.IsSpectator(),
		protocol.AbilityPrivilegedBuilder: false,
	}
	var values uint32
	for ability, value := range boolAbilities {
		if value {
			values |= ability
		}
	}

	layers := []protocol.AbilityLayer{{
		Type:             protocol.AbilityLayerTypeBase,
		Abilities:        protocol.AbilityCount - 1, // every ability above, plus the three speeds
		Values:           values,
		FlySpeed:         float32(p.GetFlightSpeedMultiplier()),
		VerticalFlySpeed: 1,
		WalkSpeed:        0.1,
	}}
	if !p.HasBlockCollision() {
		//TODO: HACK! In 1.19.80, the client starts falling in our faux spectator mode when it clips into a
		//block. We can't seem to prevent this short of forcing the player to always fly when block collision is
		//disabled. Also, for some reason the client always reads flight state from this layer if present, even
		//though the player isn't in spectator mode.
		layers = append(layers, protocol.AbilityLayer{
			Type:      protocol.AbilityLayerTypeSpectator,
			Abilities: protocol.AbilityFlying,
			Values:    protocol.AbilityFlying,
		})
	}

	commandPermissions, playerPermissions := byte(protocol.CommandPermissionLevelAny), byte(packet.PermissionLevelMember)
	if isOp {
		commandPermissions, playerPermissions = byte(protocol.CommandPermissionLevelGameDirectors), byte(packet.PermissionLevelOperator)
	}
	s.SendDataPacket(&packet.UpdateAbilities{AbilityData: protocol.AbilityData{
		EntityUniqueID:     int64(p.GetID()),
		PlayerPermissions:  playerPermissions,
		CommandPermissions: commandPermissions,
		Layers:             layers,
	}})
}

// SyncAdventureSettings is a port of NetworkSession::syncAdventureSettings.
func (s *NetworkSession) SyncAdventureSettings() {
	if s.player == nil {
		panic("Cannot sync adventure settings for a player that is not yet created")
	}
	//everything except auto jump is handled via UpdateAbilitiesPacket
	s.SendDataPacket(&packet.UpdateAdventureSettings{
		NoPvM:          false,
		NoMvP:          false,
		ImmutableWorld: false,
		ShowNameTags:   true,
		AutoJump:       s.player.HasAutoJump(),
	})
}

// commandLister is SimpleCommandMap::getCommands.
type commandLister interface {
	GetCommands() map[string]command.CommandLike
}

// SyncAvailableCommands is a port of NetworkSession::syncAvailableCommands (with
// AvailableCommandsPacketAssembler::assemble's enum table building).
func (s *NetworkSession) SyncAvailableCommands() {
	pk := &packet.AvailableCommands{}
	lister, ok := s.server.GetCommandMap().(commandLister)
	if !ok {
		s.SendDataPacket(pk)
		return
	}
	commands := lister.GetCommands()
	keys := make([]string, 0, len(commands))
	for k := range commands {
		keys = append(keys, k)
	}
	slices.Sort(keys)

	enumValueIndexes := map[string]uint32{}
	addEnumValue := func(v string) uint32 {
		if i, ok := enumValueIndexes[v]; ok {
			return i
		}
		i := uint32(len(pk.EnumValues))
		pk.EnumValues = append(pk.EnumValues, v)
		enumValueIndexes[v] = i
		return i
	}

	seen := map[string]bool{}
	for _, k := range keys {
		cmd := commands[k]
		label := cmd.Label()
		if seen[label] || label == "help" || s.player == nil || !cmd.TestPermissionSilent(s.player, nil) {
			continue
		}
		seen[label] = true

		lname := strings.ToLower(label)
		aliases := cmd.Aliases()
		aliasesOffset := uint32(0xffffffff)
		if len(aliases) > 0 {
			if !slices.Contains(aliases, lname) {
				//work around a client bug which makes the original name not show when aliases are used
				aliases = append(aliases, lname)
			}
			enum := protocol.CommandEnum{Type: strings.ToUpper(label[:1]) + label[1:] + "Aliases"}
			for _, alias := range aliases {
				enum.ValueIndices = append(enum.ValueIndices, addEnumValue(alias))
			}
			aliasesOffset = uint32(len(pk.Enums))
			pk.Enums = append(pk.Enums, enum)
		}

		description := cmd.Description()
		var descriptionText string
		if t, ok := description.(*lang.Translatable); ok {
			descriptionText = s.player.GetLanguage().Translate(t)
		} else if description != nil {
			descriptionText = fmt.Sprint(description)
		}
		pk.Commands = append(pk.Commands, protocol.Command{
			Name:            lname, //TODO: commands containing uppercase letters in the name crash 1.9.0 client
			Description:     descriptionText,
			Flags:           0,
			PermissionLevel: byte(protocol.CommandPermissionLevelAny),
			AliasesOffset:   aliasesOffset,
			Overloads: []protocol.CommandOverload{{
				Chaining:   false,
				Parameters: []protocol.CommandParameter{{Name: "args", Type: protocol.CommandArgValid | protocol.CommandArgTypeRawText, Optional: true}},
			}},
		})
	}
	s.SendDataPacket(pk)
}

// PrepareClientTranslatableMessage is a port of NetworkSession::prepareClientTranslatableMessage:
// server-side ("pocketmine.") translations are resolved, the rest is left for the client.
func (s *NetworkSession) PrepareClientTranslatableMessage(message *lang.Translatable) (string, []string) {
	//we can't send nested translations to the client, so make sure they are always pre-translated by the server
	language := s.player.GetLanguage()
	parameters := make([]any, len(message.Parameters()))
	keys := make([]string, len(message.Parameters()))
	for i, p := range message.Parameters() {
		if t, ok := p.(*lang.Translatable); ok {
			parameters[i] = language.Translate(t)
		} else {
			parameters[i] = fmt.Sprint(p)
		}
		keys[i] = message.ParameterKey(i)
	}
	prefix := "pocketmine."
	translated, untranslatedParameterCount := language.TranslateStringKeyed(message.Text(), keys, parameters, &prefix)
	result := make([]string, 0, untranslatedParameterCount)
	for _, p := range parameters[:min(untranslatedParameterCount, len(parameters))] {
		result = append(result, p.(string))
	}
	return translated, result
}

// OnChatMessage is a port of NetworkSession::onChatMessage.
func (s *NetworkSession) OnChatMessage(message any) {
	if t, ok := message.(*lang.Translatable); ok {
		if !s.server.IsLanguageForced() {
			text, params := s.PrepareClientTranslatableMessage(t)
			s.SendDataPacket(&packet.Text{TextType: packet.TextTypeTranslation, NeedsTranslation: true, Message: text, Parameters: params})
		} else {
			s.SendDataPacket(&packet.Text{TextType: packet.TextTypeRaw, Message: s.player.GetLanguage().Translate(t)})
		}
		return
	}
	s.SendDataPacket(&packet.Text{TextType: packet.TextTypeRaw, Message: fmt.Sprint(message)})
}

// OnJukeboxPopup is a port of NetworkSession::onJukeboxPopup.
func (s *NetworkSession) OnJukeboxPopup(message any) {
	var params []string
	text := fmt.Sprint(message)
	if t, ok := message.(*lang.Translatable); ok {
		if !s.server.IsLanguageForced() {
			text, params = s.PrepareClientTranslatableMessage(t)
		} else {
			text = s.player.GetLanguage().Translate(t)
		}
	}
	s.SendDataPacket(&packet.Text{TextType: packet.TextTypeJukeboxPopup, NeedsTranslation: true, Message: text, Parameters: params})
}

// OnPopup is a port of NetworkSession::onPopup.
func (s *NetworkSession) OnPopup(message string) {
	s.SendDataPacket(&packet.Text{TextType: packet.TextTypePopup, Message: message})
}

// OnTip is a port of NetworkSession::onTip.
func (s *NetworkSession) OnTip(message string) {
	s.SendDataPacket(&packet.Text{TextType: packet.TextTypeTip, Message: message})
}

// OnFormSent is a port of NetworkSession::onFormSent.
func (s *NetworkSession) OnFormSent(id int, f form.Form) bool {
	data, err := json.Marshal(f)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to encode form %d: %v", id, err))
		return false
	}
	return s.sendDataPacket(&packet.ModalFormRequest{FormID: uint32(id), FormData: data}, false)
}

// OnCloseAllForms is a port of NetworkSession::onCloseAllForms.
func (s *NetworkSession) OnCloseAllForms() {
	s.SendDataPacket(&packet.ClientBoundCloseForm{})
}

// StartUsingChunk is a port of NetworkSession::startUsingChunk: the chunk is sent through the
// world's ChunkCache (in sub-chunk request mode, see sub_chunk_request.go), then onCompletion runs.
// Chunk serialization is synchronous here, so there's no promise to wait for.
func (s *NetworkSession) StartUsingChunk(chunkX, chunkZ int, onCompletion func()) {
	w := s.player.GetWorld()
	chunk, ok := w.GetChunk(chunkX, chunkZ)
	if !ok {
		s.logger.Debug(fmt.Sprintf("Tried to send no-longer-active chunk %d %d in world %s", chunkX, chunkZ, w.GetFolderName()))
		return
	}
	data := GetChunkCache(w).request(chunkX, chunkZ, chunk)
	s.SendDataPacket(levelChunkPacket(chunkX, chunkZ, data, s.blobCache))
	onCompletion()
}

// StopUsingChunk is a port of NetworkSession::stopUsingChunk (empty in PHP too).
func (s *NetworkSession) StopUsingChunk(chunkX, chunkZ int) {}

// OnEnterWorld is a port of NetworkSession::onEnterWorld.
func (s *NetworkSession) OnEnterWorld() {
	if s.player != nil {
		w := s.player.GetWorld()
		s.SyncWorldTime(int(w.GetTime()))
		s.SyncWorldDifficulty(w.GetDifficulty())
		s.SyncWorldSpawnPoint(w.GetSpawnLocation())
		//TODO: weather needs to be synced here (when implemented)
	}
}

// SyncWorldTime is a port of NetworkSession::syncWorldTime.
func (s *NetworkSession) SyncWorldTime(worldTime int) {
	s.SendDataPacket(&packet.SetTime{Time: int32(worldTime)})
}

// SyncWorldDifficulty is a port of NetworkSession::syncWorldDifficulty.
func (s *NetworkSession) SyncWorldDifficulty(worldDifficulty int) {
	s.SendDataPacket(&packet.SetDifficulty{Difficulty: uint32(worldDifficulty)})
}

// playerListEntry is PlayerListEntry::createAdditionEntry for p. The skin is the one p's client
// sent at login (see playerListSkin: PHP sends TypeConverter's LegacySkinAdapter::toSkinData,
// which 1.26.51 clients left with "Block" after); players without a session fall back to it.
func playerListEntry(p *player.Player) protocol.PlayerListEntry {
	skin := entity.SkinToNetwork(p.GetSkin())
	if session, ok := p.GetNetworkSession().(*NetworkSession); ok && session.playerListSkin != nil {
		skin = *session.playerListSkin
	}
	return protocol.PlayerListEntry{
		ActionType:     protocol.PlayerListActionAdd,
		UUID:           p.GetUniqueID(),
		EntityUniqueID: int64(p.GetID()),
		Username:       p.GetDisplayName(),
		XUID:           p.GetXuid(),
		BuildPlatform:  -1, // DeviceOS::UNKNOWN, PlayerListEntry::createAdditionEntry's default
		Skin:           skin,
	}
}

// SyncPlayerList is a port of NetworkSession::syncPlayerList.
func (s *NetworkSession) SyncPlayerList(players []*player.Player) {
	entries := make([]protocol.PlayerListEntry, 0, len(players))
	for _, p := range players {
		entries = append(entries, playerListEntry(p))
	}
	s.SendDataPacket(&packet.PlayerList{Entries: entries})
}

// OnPlayerAdded is a port of NetworkSession::onPlayerAdded.
func (s *NetworkSession) OnPlayerAdded(p *player.Player) {
	s.SendDataPacket(&packet.PlayerList{Entries: []protocol.PlayerListEntry{playerListEntry(p)}})
}

// OnPlayerRemoved is a port of NetworkSession::onPlayerRemoved.
func (s *NetworkSession) OnPlayerRemoved(p *player.Player) {
	if p != s.player {
		s.SendDataPacket(&packet.PlayerList{Entries: []protocol.PlayerListEntry{{ActionType: protocol.PlayerListActionRemove, UUID: p.GetUniqueID()}}})
	}
}

// OnTitle is a port of NetworkSession::onTitle.
func (s *NetworkSession) OnTitle(title string) {
	s.SendDataPacket(&packet.SetTitle{ActionType: packet.TitleActionSetTitle, Text: title})
}

// OnSubTitle is a port of NetworkSession::onSubTitle.
func (s *NetworkSession) OnSubTitle(subtitle string) {
	s.SendDataPacket(&packet.SetTitle{ActionType: packet.TitleActionSetSubtitle, Text: subtitle})
}

// OnActionBar is a port of NetworkSession::onActionBar.
func (s *NetworkSession) OnActionBar(actionBar string) {
	s.SendDataPacket(&packet.SetTitle{ActionType: packet.TitleActionSetActionBar, Text: actionBar})
}

// OnClearTitle is a port of NetworkSession::onClearTitle.
func (s *NetworkSession) OnClearTitle() {
	s.SendDataPacket(&packet.SetTitle{ActionType: packet.TitleActionClear})
}

// OnResetTitleOptions is a port of NetworkSession::onResetTitleOptions.
func (s *NetworkSession) OnResetTitleOptions() {
	s.SendDataPacket(&packet.SetTitle{ActionType: packet.TitleActionReset})
}

// OnTitleDuration is a port of NetworkSession::onTitleDuration.
func (s *NetworkSession) OnTitleDuration(fadeIn, stay, fadeOut int) {
	s.SendDataPacket(&packet.SetTitle{ActionType: packet.TitleActionSetDurations, FadeInDuration: int32(fadeIn), RemainDuration: int32(stay), FadeOutDuration: int32(fadeOut)})
}

// OnToastNotification is a port of NetworkSession::onToastNotification.
func (s *NetworkSession) OnToastNotification(title, body string) {
	s.SendDataPacket(&packet.ToastRequest{Title: title, Message: body})
}

// OnOpenSignEditor is a port of NetworkSession::onOpenSignEditor.
func (s *NetworkSession) OnOpenSignEditor(signPosition math.Vector3, frontSide bool) {
	s.SendDataPacket(&packet.OpenSign{Position: blockPos(signPosition), FrontSide: frontSide})
}

// OnItemCooldownChanged is a port of NetworkSession::onItemCooldownChanged.
func (s *NetworkSession) OnItemCooldownChanged(it item.Item, ticks int) {
	name, ok := convert.ItemTypeName(it)
	if !ok {
		return
	}
	s.SendDataPacket(&PlayerStartItemCooldown{ItemCategory: name, CooldownTicks: int32(ticks)})
}

// Tick is a port of NetworkSession::tick.
func (s *NetworkSession) Tick() {
	if !s.IsConnected() {
		s.dispose()
		return
	}

	if s.info == nil {
		if time.Since(s.connectTime) >= 10*time.Second {
			s.DisconnectWithError(lang.KnownTranslationFactory.PocketmineDisconnectErrorLoginTimeout(), nil)
		}
		return
	}

	if s.player != nil {
		s.player.DoChunkRequests()

		dirtyAttributes := s.player.GetAttributeMap().NeedSend()
		s.entityEventBroadcaster.SyncAttributes([]*NetworkSession{s}, s.player, dirtyAttributes)
		for _, attribute := range dirtyAttributes {
			//TODO: we might need to send these to other players in the future
			//if that happens, this will need to become more complex than a flag on the attribute itself
			attribute.MarkSynchronized(true)
		}
	}
	if s.invManager != nil {
		s.invManager.FlushPendingUpdates()
	}

	s.flushGamePacketQueue()
}

// CalculateUUIDFromXUID is a port of LoginPacketHandler::calculateUuidFromXuid.
func CalculateUUIDFromXUID(xuid string) uuid.UUID {
	hash := md5.Sum([]byte("pocket-auth-1-xuid:" + xuid))
	hash[6] = (hash[6] & 0x0f) | 0x30 // set version to 3
	hash[8] = (hash[8] & 0x3f) | 0x80 // set variant to RFC 4122
	return uuid.UUID(hash)
}

// NewPlayerInfoFromLogin is the PlayerInfo half of LoginPacketHandler::handleLogin +
// processLoginCommon, from the login data gophertunnel parsed. The returned error is the
// disconnect screen message (and reason) when the login data is invalid.
func NewPlayerInfoFromLogin(identity login.IdentityData, clientData login.ClientData, authenticated bool) (player.Info, any, any) {
	username := identity.DisplayName
	if !player.IsValidUserName(username) {
		return nil, lang.KnownTranslationFactory.DisconnectionScreenInvalidName(), nil
	}

	networkSkin, err := convert.ClientDataToSkinData(clientData)
	if err == nil {
		var skin *entity.Skin
		if skin, err = entity.SkinFromNetwork(networkSkin); err == nil {
			extraData := map[string]any{}
			if raw, err := json.Marshal(clientData); err == nil {
				_ = json.Unmarshal(raw, &extraData)
			}
			if authenticated && identity.XUID != "" {
				return player.NewXboxLivePlayerInfo(identity.XUID, username, CalculateUUIDFromXUID(identity.XUID).String(), skin, clientData.LanguageCode, extraData), nil, nil
			}
			legacyUUID := identity.Identity
			if _, err := uuid.Parse(legacyUUID); err != nil {
				return nil, "Invalid UUID string in self-signed certificate: " + legacyUUID, nil
			}
			return player.NewPlayerInfo(username, legacyUUID, skin, clientData.LanguageCode, extraData), nil, nil
		}
	}
	return nil, "Invalid skin: " + err.Error(), lang.KnownTranslationFactory.DisconnectionScreenInvalidSkin()
}

var (
	_ player.NetworkSession      = (*NetworkSession)(nil)
	_ network.Session            = (*NetworkSession)(nil)
	_ serverevent.NetworkSession = (*NetworkSession)(nil)
)
