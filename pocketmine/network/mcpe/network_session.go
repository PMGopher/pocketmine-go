// Package mcpe is a port of pocketmine\network\mcpe: the connection-level code between the
// Bedrock protocol and the game. The protocol itself (packets, RakNet, login, encryption) is
// github.com/sandertv/gophertunnel instead of a port of pmmp/BedrockProtocol and RakLib (see
// AGENTS.md §1), so NetworkSession wraps a gophertunnel *minecraft.Conn, which has already
// completed the login sequence by the time a session is created.
package mcpe

import (
	"fmt"
	"net"
	"strconv"
	"sync"

	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

	"pocketmine-go/pocketmine/entity"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/log"
	"pocketmine-go/pocketmine/network/mcpe/convert"
	"pocketmine-go/pocketmine/player"
)

// Server is what NetworkSession needs from pocketmine\Server. It's declared here so the server
// package, which imports this one, satisfies it without an import cycle.
type Server interface {
	// Locker serialises packet handling with the server tick: PocketMine-MP does both on one
	// thread, and the world, entities and players aren't safe for concurrent use.
	sync.Locker
	GetLogger() log.Logger
	GetMotd() string
	// CreatePlayer is a port of Server::createPlayer (synchronous: this port generates chunks
	// inline, so World::requestSafeSpawn needs no promise).
	CreatePlayer(session *NetworkSession, playerInfo *player.XboxLivePlayerInfo) (*player.Player, error)
	// AddOnlinePlayer is a port of Server::addOnlinePlayer.
	AddOnlinePlayer(p *player.Player) bool
	// GetOnlinePlayers is a port of Server::getOnlinePlayers.
	GetOnlinePlayers() []*player.Player
	// GetAllowedViewDistance is a port of Server::getAllowedViewDistance.
	GetAllowedViewDistance(distance int) int
	// GetDefaultGameMode is Server::getGamemode as the network game mode ID.
	GetDefaultGameMode() int32
}

// NetworkSession is a port of pocketmine\network\mcpe\NetworkSession.
//
// Not ported: packet batching/compression/encryption and the send/receive buffers (gophertunnel
// does all of that), the rate limiter, and plugin-facing DataPacketSend/ReceiveEvents (packet
// events aren't ported).
type NetworkSession struct {
	server Server
	conn   *minecraft.Conn
	logger log.Logger

	player  *player.Player
	handler PacketHandler

	connected      bool
	disconnectOnce sync.Once
}

// NewNetworkSession creates the session for a connection gophertunnel has just accepted (login,
// encryption and resource packs already done).
func NewNetworkSession(server Server, conn *minecraft.Conn) *NetworkSession {
	s := &NetworkSession{server: server, conn: conn, connected: true}
	s.logger = log.NewPrefixedLogger(server.GetLogger(), s.getLogPrefix())
	return s
}

// getLogPrefix is a port of NetworkSession::getLogPrefix.
func (s *NetworkSession) getLogPrefix() string {
	return "NetworkSession: " + s.GetDisplayName()
}

func (s *NetworkSession) GetConn() *minecraft.Conn { return s.conn }
func (s *NetworkSession) GetLogger() log.Logger    { return s.logger }
func (s *NetworkSession) GetServer() Server        { return s.server }

// GetPlayer is a port of NetworkSession::getPlayer.
func (s *NetworkSession) GetPlayer() *player.Player { return s.player }

// IsConnected is a port of NetworkSession::isConnected.
func (s *NetworkSession) IsConnected() bool { return s.connected }

// GetIp is a port of NetworkSession::getIp.
func (s *NetworkSession) GetIp() string {
	host, _, err := net.SplitHostPort(s.conn.RemoteAddr().String())
	if err != nil {
		return s.conn.RemoteAddr().String()
	}
	return host
}

// GetPort is a port of NetworkSession::getPort.
func (s *NetworkSession) GetPort() int {
	_, port, err := net.SplitHostPort(s.conn.RemoteAddr().String())
	if err != nil {
		return 0
	}
	p, _ := strconv.Atoi(port)
	return p
}

// GetDisplayName is a port of NetworkSession::getDisplayName.
func (s *NetworkSession) GetDisplayName() string {
	if name := s.conn.IdentityData().DisplayName; name != "" {
		return name
	}
	return s.GetIp() + " " + strconv.Itoa(s.GetPort())
}

// GetHandler is a port of NetworkSession::getHandler.
func (s *NetworkSession) GetHandler() PacketHandler { return s.handler }

// SetHandler is a port of NetworkSession::setHandler.
func (s *NetworkSession) SetHandler(handler PacketHandler) error {
	if !s.connected { //TODO: this is fine since we can't handle anything from a disconnected session, but it might produce surprises in some cases
		return nil
	}
	s.handler = handler
	if handler != nil {
		return handler.SetUp()
	}
	return nil
}

// Run drives the session until the connection closes: NetworkSession::onServerLoginSuccess (the
// login itself was done by gophertunnel), the pre-spawn sequence, the spawn response and then
// the packet read loop. It returns once the client has disconnected.
func (s *NetworkSession) Run() {
	defer s.onClientDisconnect()

	if err := s.onServerLoginSuccess(); err != nil {
		s.logger.Warning(err.Error())
		return
	}

	for {
		pk, err := s.conn.ReadPacket()
		if err != nil {
			s.logger.Info(fmt.Sprintf("Connection closed: %v", err))
			return
		}
		s.server.Lock()
		s.handleDataPacket(pk)
		s.server.Unlock()
	}
}

// onServerLoginSuccess is a port of NetworkSession::onServerLoginSuccess + createPlayer +
// onPlayerCreated, followed by the spawn sequence (PreSpawnPacketHandler, then
// SpawnResponsePacketHandler's onClientSpawnResponse).
func (s *NetworkSession) onServerLoginSuccess() error {
	identity := s.conn.IdentityData()

	networkSkin, err := convert.ClientDataToSkinData(s.conn.ClientData())
	if err != nil {
		s.Disconnect("disconnectionScreen.invalidSkin")
		return fmt.Errorf("invalid skin: %w", err)
	}
	skin, err := entity.SkinFromNetwork(networkSkin)
	if err != nil {
		s.Disconnect("disconnectionScreen.invalidSkin")
		return fmt.Errorf("invalid skin: %w", err)
	}

	playerUUID := identity.Identity
	if _, err := uuid.Parse(playerUUID); err != nil {
		// Unauthenticated clients don't necessarily send a real identity UUID.
		playerUUID = uuid.NewString()
	}
	info := player.NewXboxLivePlayerInfo(identity.XUID, identity.DisplayName, playerUUID, skin, s.conn.ClientData().LanguageCode, nil)

	s.server.Lock()
	p, err := s.server.CreatePlayer(s, info)
	if err == nil {
		s.player = p
		p.SetNetworkSession(s)
		if !s.server.AddOnlinePlayer(p) {
			err = fmt.Errorf("player was not accepted")
		}
	}
	if err != nil {
		s.Disconnect("disconnectionScreen.noReason")
		s.server.Unlock()
		return fmt.Errorf("creating player: %w", err)
	}
	s.server.Unlock()

	if err := s.SetHandler(NewPreSpawnPacketHandler(s)); err != nil {
		return fmt.Errorf("pre-spawn: %w", err)
	}
	s.logger.Debug("Waiting for spawn chunks")

	s.server.Lock()
	defer s.server.Unlock()
	return s.onClientSpawnResponse()
}

// onClientSpawnResponse is a port of NetworkSession::onClientSpawnResponse: the client has
// loaded, so the player spawns into the world and the in-game handler takes over.
//
// gophertunnel's StartGame already waits for the client's RequestChunkRadius and
// SetLocalPlayerAsInitialised (the job of PreSpawnPacketHandler::handleRequestChunkRadius and
// SpawnResponsePacketHandler), so the view distance is applied and the first chunks are sent here.
func (s *NetworkSession) onClientSpawnResponse() error {
	s.logger.Debug("Received spawn response, entering in-game phase")
	p := s.player
	p.SetViewDistance(s.server.GetAllowedViewDistance(s.conn.ChunkRadius()))
	if err := s.SyncViewAreaCenterPoint(); err != nil {
		return err
	}
	if err := s.SetHandler(NewInGamePacketHandler(s)); err != nil {
		return err
	}
	p.DoFirstSpawn()
	return nil
}

// Tick is the per-tick work for this session (NetworkSession::tick plus Player::doChunkRequests):
// sends the chunks the player is waiting for.
func (s *NetworkSession) Tick() {
	if s.player == nil || !s.player.IsSpawned() {
		return
	}
	if err := s.doChunkRequests(); err != nil {
		s.logger.Warning(fmt.Sprintf("Failed to send chunks: %v", err))
		s.Disconnect("disconnectionScreen.noReason")
	}
}

// doChunkRequests is a port of Player::doChunkRequests's network half: every chunk that became
// ready is sent (NetworkSession::startUsingChunk) in sub-chunk request mode (see
// sub_chunk_request.go), then the view area centre is synced.
func (s *NetworkSession) doChunkRequests() error {
	p := s.player
	p.OrderChunks()
	ready := p.RequestChunks()
	if len(ready) == 0 {
		return nil
	}
	w := p.GetWorld()
	for _, c := range ready {
		chunk, ok := w.GetChunk(c[0], c[1])
		if !ok {
			continue
		}
		if err := s.conn.WritePacket(LevelChunkPacket(c[0], c[1], chunk)); err != nil {
			return err
		}
		p.MarkChunkSent(c[0], c[1])
	}
	return s.SyncViewAreaCenterPoint()
}

// SyncViewAreaCenterPoint is a port of NetworkSession::syncViewAreaCenterPoint.
func (s *NetworkSession) SyncViewAreaCenterPoint() error {
	pos := s.player.GetPosition()
	return s.conn.WritePacket(&packet.NetworkChunkPublisherUpdate{
		Position: protocol.BlockPos{int32(pos.FloorX()), int32(pos.FloorY()), int32(pos.FloorZ())},
		Radius:   uint32(s.player.GetViewDistance() * 16), //blocks, not chunks >.>
	})
}

// handleDataPacket is a port of NetworkSession::handleDataPacket: the packet goes to the current
// handler; anything it doesn't handle is logged at debug level (PHP's unhandledPacketDebug).
func (s *NetworkSession) handleDataPacket(pk packet.Packet) {
	if violation, ok := pk.(*packet.PacketViolationWarning); ok {
		// PHP 5.44 only reaches this through the unhandled-packet debug log; it's logged as a
		// warning here because it's the client telling us which packet of ours it rejected.
		s.logger.Warning(fmt.Sprintf("Client reported a packet violation: type=%d severity=%d packetID=%d context=%q",
			violation.Type, violation.Severity, violation.PacketID, violation.ViolationContext))
		return
	}
	if s.handler == nil || !s.handler.HandleDataPacket(pk) {
		s.logger.Debug(fmt.Sprintf("Unhandled %T: %+v", pk, pk))
	}
}

// SendDataPacket is a port of NetworkSession::sendDataPacket.
func (s *NetworkSession) SendDataPacket(pk packet.Packet) {
	if !s.connected {
		return
	}
	if err := s.conn.WritePacket(pk); err != nil {
		s.logger.Debug(fmt.Sprintf("Failed to send %T: %v", pk, err))
	}
}

// OnChatMessage is a port of NetworkSession::onChatMessage. Translatables are sent for the client
// to translate (prepareClientTranslatableMessage without server-side "pocketmine." translations:
// no Language is loaded yet).
func (s *NetworkSession) OnChatMessage(message any) {
	switch m := message.(type) {
	case *lang.Translatable:
		params := make([]string, 0, len(m.Parameters()))
		for _, param := range m.Parameters() {
			params = append(params, fmt.Sprint(param))
		}
		s.SendDataPacket(&packet.Text{TextType: packet.TextTypeTranslation, NeedsTranslation: true, Message: m.Text(), Parameters: params})
	default:
		s.SendDataPacket(&packet.Text{TextType: packet.TextTypeRaw, Message: fmt.Sprint(message)})
	}
}

// Disconnect is a port of NetworkSession::disconnect / Player::disconnect: the client is shown
// reason, the connection is closed and the player goes through Player::onPostDisconnect (quit
// message, save, removal). The server lock must be held (PHP runs this on the main thread).
func (s *NetworkSession) Disconnect(reason string) {
	s.disconnectOnce.Do(func() {
		s.connected = false
		_ = s.conn.WritePacket(&packet.Disconnect{Message: reason})
		_ = s.conn.Close()
		if s.player != nil {
			s.player.OnPostDisconnect()
		}
	})
}

// onClientDisconnect is a port of NetworkSession::onClientDisconnect: the client closed the
// connection, which ends the same way as Disconnect.
func (s *NetworkSession) onClientDisconnect() {
	s.server.Lock()
	defer s.server.Unlock()
	s.disconnectOnce.Do(func() {
		s.connected = false
		_ = s.conn.Close()
		if s.player != nil {
			s.player.OnPostDisconnect()
		}
	})
}

// playerListEntry is the PlayerListEntry for p (PlayerListPacket::add with
// TypeConverter::getSkinAdapter()->toSkinData($player->getSkin())).
func playerListEntry(p *player.Player) protocol.PlayerListEntry {
	return protocol.PlayerListEntry{
		ActionType:     protocol.PlayerListActionAdd,
		UUID:           p.GetUniqueID(),
		EntityUniqueID: int64(p.GetID()),
		Username:       p.GetName(),
		XUID:           p.GetXuid(),
		Skin:           entity.SkinToNetwork(p.GetSkin()),
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

// SyncAbilities is a port of NetworkSession::syncAbilities. Operator status and the teleport
// permission are always false: players aren't permissibles yet (permissions aren't wired to
// players).
func (s *NetworkSession) SyncAbilities(p *player.Player) {
	isOp := false

	//ALL of these need to be set for the base layer, otherwise the client will cry
	boolAbilities := map[uint32]bool{
		protocol.AbilityMayFly:            p.GetAllowFlight(),
		protocol.AbilityFlying:            p.IsFlying(),
		protocol.AbilityNoClip:            !p.HasBlockCollision(),
		protocol.AbilityOperatorCommands:  isOp,
		protocol.AbilityTeleport:          false,
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
	//everything except auto jump is handled via UpdateAbilitiesPacket
	s.SendDataPacket(&packet.UpdateAdventureSettings{
		NoPvM:          false,
		NoMvP:          false,
		ImmutableWorld: false,
		ShowNameTags:   true,
		AutoJump:       s.player.HasAutoJump(),
	})
}

// SyncAllInventories is InventoryManager::syncAll for the player's permanent windows (inventory,
// armor, off-hand). InventoryManager itself isn't ported: there are no other windows yet, and no
// container/slot tracking for ItemStackRequests.
func (s *NetworkSession) SyncAllInventories() {
	p := s.player
	windows := []struct {
		windowID  uint32
		container byte
		inventory interface {
			GetSize() int
			GetItem(int) item.Item
		}
	}{
		{protocol.WindowIDInventory, protocol.ContainerCombinedHotBarAndInventory, p.GetInventory()},
		{protocol.WindowIDArmour, protocol.ContainerArmor, p.GetArmorInventory()},
		{protocol.WindowIDOffHand, protocol.ContainerOffhand, p.GetOffHandInventory()},
	}
	for _, w := range windows {
		content := make([]protocol.ItemInstance, w.inventory.GetSize())
		for i := range content {
			content[i] = convert.ItemStackWrapperLegacy(w.inventory.GetItem(i))
		}
		s.SendDataPacket(&packet.InventoryContent{
			WindowID:  w.windowID,
			Content:   content,
			Container: protocol.FullContainerName{ContainerID: w.container},
		})
	}
}

// SyncSelectedHotbarSlot is a port of InventoryManager::syncSelectedHotbarSlot.
func (s *NetworkSession) SyncSelectedHotbarSlot() {
	s.SendDataPacket(&packet.PlayerHotBar{
		SelectedHotBarSlot: uint32(s.player.GetInventory().GetHeldItemIndex()),
		WindowID:           byte(protocol.WindowIDInventory),
		SelectHotBarSlot:   true,
	})
}
