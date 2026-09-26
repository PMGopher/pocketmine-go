// Package handler is a port of pocketmine\network\mcpe\handler: the packet handlers a
// NetworkSession switches between as the connection goes through its phases.
package handler

import (
	"sync"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

	"pocketmine-go/pocketmine/data/bedrock"
	"pocketmine-go/pocketmine/network/mcpe"
	"pocketmine-go/pocketmine/network/mcpe/convert"
	"pocketmine-go/pocketmine/player"
	"pocketmine-go/pocketmine/world"
)

func init() {
	mcpe.NewPreSpawnPacketHandler = func(server mcpe.Server, p *player.Player, session *mcpe.NetworkSession, invManager *mcpe.InventoryManager) mcpe.PacketHandler {
		return NewPreSpawnPacketHandler(server, p, session, invManager)
	}
	mcpe.NewInGamePacketHandler = func(p *player.Player, session *mcpe.NetworkSession, invManager *mcpe.InventoryManager) mcpe.PacketHandler {
		return NewInGamePacketHandler(p, session, invManager)
	}
	mcpe.NewDeathPacketHandler = func(p *player.Player, session *mcpe.NetworkSession, invManager *mcpe.InventoryManager, deathMessage any) mcpe.PacketHandler {
		return NewDeathPacketHandler(p, session, invManager, deathMessage)
	}
}

// PreSpawnPacketHandler is a port of pocketmine\network\mcpe\handler\PreSpawnPacketHandler: sends
// StartGame and everything else the client needs before it can spawn.
//
// StartGame goes through gophertunnel's Conn.StartGame, which also sends ItemRegistry, answers the
// client's RequestChunkRadius itself and blocks until the client has spawned; the session then
// applies the requested radius (NetworkSession.OnClientRequestChunkRadius, the port of
// handleRequestChunkRadius).
type PreSpawnPacketHandler struct {
	server           mcpe.Server
	player           *player.Player
	session          *mcpe.NetworkSession
	inventoryManager *mcpe.InventoryManager
}

func NewPreSpawnPacketHandler(server mcpe.Server, p *player.Player, session *mcpe.NetworkSession, inventoryManager *mcpe.InventoryManager) *PreSpawnPacketHandler {
	return &PreSpawnPacketHandler{server: server, player: p, session: session, inventoryManager: inventoryManager}
}

// SetUp is a port of PreSpawnPacketHandler::setUp. It's called with the server lock held, which
// is released while gophertunnel waits for the client, so the world keeps ticking meanwhile.
func (h *PreSpawnPacketHandler) SetUp() {
	session, p := h.session, h.player
	logger := session.GetLogger()

	logger.Debug("Preparing StartGamePacket")
	data := h.startGameData()
	if conn := session.GetConn(); conn != nil {
		logger.Debug("Sending items")
		h.server.Unlock()
		err := conn.StartGame(data)
		h.server.Lock()
		if err != nil {
			if session.IsConnected() {
				session.DisconnectWithError("Failed to start game: "+err.Error(), nil)
			}
			return
		}
		if !session.IsConnected() {
			return
		}
	}

	logger.Debug("Sending actor identifiers")
	session.SendDataPacket(bedrock.AvailableActorIdentifiers())

	logger.Debug("Sending biome definitions")
	session.SendDataPacket(bedrock.BiomeDefinitionList())

	logger.Debug("Sending attributes")
	session.GetEntityEventBroadcaster().SyncAttributes([]*mcpe.NetworkSession{session}, p, p.GetAttributeMap().GetAll())

	logger.Debug("Sending available commands")
	session.SyncAvailableCommands()

	logger.Debug("Sending abilities")
	session.SyncAbilities(p)
	session.SyncAdventureSettings()

	logger.Debug("Sending effects")
	for _, effect := range p.GetEffects().All() {
		session.GetEntityEventBroadcaster().OnEntityEffectAdded([]*mcpe.NetworkSession{session}, p, effect, false)
	}

	logger.Debug("Sending actor metadata")
	p.SendData([]world.EntityViewer{p}, nil)

	logger.Debug("Sending inventory")
	h.inventoryManager.SyncAll()
	h.inventoryManager.SyncSelectedHotbarSlot()

	logger.Debug("Sending creative inventory data")
	h.inventoryManager.SyncCreative()

	logger.Debug("Sending crafting data")
	session.SendDataPacket(mcpe.GetCraftingDataCache().GetCache(h.server.GetCraftingManager()))

	logger.Debug("Sending player list")
	session.SyncPlayerList(h.server.GetOnlinePlayers())
}

// startGameData is the StartGamePacket PreSpawnPacketHandler::setUp builds, as gophertunnel's
// GameData.
func (h *PreSpawnPacketHandler) startGameData() minecraft.GameData {
	p := h.player
	location := p.GetLocation()
	w := p.GetWorld()
	spawn := w.GetSpawnLocation()
	// The client's position is its eye position: $this->player->getOffsetPosition($location).
	eyePos := p.GetOffsetPosition(location.Vector3)

	return minecraft.GameData{
		WorldName:       h.server.GetMotd(),
		WorldSeed:       -1,
		Difficulty:      int32(w.GetDifficulty()),
		EntityUniqueID:  int64(p.GetID()),
		EntityRuntimeID: uint64(p.GetID()),
		PlayerGameMode:  convert.CoreGameModeToProtocol(int(p.GetGamemode())),
		PlayerPosition:  mgl32.Vec3{float32(eyePos.X), float32(eyePos.Y), float32(eyePos.Z)},
		Pitch:           float32(location.Pitch),
		Yaw:             float32(location.Yaw),
		WorldSpawn:      protocol.BlockPos{int32(spawn.FloorX()), int32(spawn.FloorY()), int32(spawn.FloorZ())},
		WorldGameMode:   convert.CoreGameModeToProtocol(int(h.server.GetGamemode())),
		Time:            int64(w.GetTime()),
		GameRules: []protocol.GameRule{
			{Name: "naturalregeneration", Value: false}, //Hack for client side regeneration
			{Name: "locatorbar", Value: false},          //Disable client-side tracking of nearby players
		},
		// The protocol library's LevelSettings default; without it the 1.26.50 client doesn't
		// load the current vanilla block definitions.
		BaseGameVersion:              protocol.CurrentVersion,
		PlayerMovementSettings:       protocol.PlayerMovementSettings{RewindHistorySize: 0, ServerAuthoritativeBlockBreaking: true},
		ServerAuthoritativeInventory: true,
		PlayerPermissions:            packet.PermissionLevelMember,
		ChunkRadius:                  int32(h.server.GetAllowedViewDistance(1 << 30)),
		Items:                        itemRegistryEntries(),
		// Since Bedrock 1.26.50 the client needs the data-driven vanilla block definitions to
		// complete its block palette (see bedrock.DataDrivenBlocks).
		CustomBlocks: dataDrivenBlockEntries(),
	}
}

// CanHandle reports the packets PreSpawnPacketHandler handles (PacketHandlerInspector).
func (h *PreSpawnPacketHandler) CanHandle(pk packet.Packet) bool {
	_, ok := pk.(*packet.RequestChunkRadius)
	return ok
}

// HandleDataPacket dispatches pk to handleRequestChunkRadius.
func (h *PreSpawnPacketHandler) HandleDataPacket(pk packet.Packet) (bool, error) {
	if pk, ok := pk.(*packet.RequestChunkRadius); ok {
		return h.handleRequestChunkRadius(pk), nil
	}
	return false, nil
}

// handleRequestChunkRadius is a port of PreSpawnPacketHandler::handleRequestChunkRadius.
func (h *PreSpawnPacketHandler) handleRequestChunkRadius(pk *packet.RequestChunkRadius) bool {
	h.player.SetViewDistance(int(pk.ChunkRadius))
	return true
}

var (
	itemRegistryOnce  sync.Once
	itemRegistryCache []protocol.ItemEntry
)

// itemRegistryEntries is TypeConverter::getItemTypeDictionary()->getEntries() as ItemRegistry
// entries. It's the same for every player, so it's built once.
func itemRegistryEntries() []protocol.ItemEntry {
	itemRegistryOnce.Do(func() {
		types := bedrock.ItemTypes()
		itemRegistryCache = make([]protocol.ItemEntry, len(types))
		for i, t := range types {
			itemRegistryCache[i] = protocol.ItemEntry{
				Name:           t.Name,
				RuntimeID:      int16(t.RuntimeID),
				ComponentBased: t.ComponentBased,
				Version:        t.Version,
				Data:           t.Data,
			}
		}
	})
	return itemRegistryCache
}

// dataDrivenBlockEntries is StartGame's block list (see bedrock.DataDrivenBlocks).
func dataDrivenBlockEntries() []protocol.BlockEntry {
	blocks := bedrock.DataDrivenBlocks()
	entries := make([]protocol.BlockEntry, len(blocks))
	for i, b := range blocks {
		entries[i] = protocol.BlockEntry{Name: b.Name, Properties: b.Components}
	}
	return entries
}
