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
	"pocketmine-go/pocketmine/entity"
	"pocketmine-go/pocketmine/network/mcpe"
	"pocketmine-go/pocketmine/network/mcpe/convert"
	"pocketmine-go/pocketmine/player"
	"pocketmine-go/pocketmine/world"
)

func init() {
	mcpe.NewPreSpawnPacketHandler = func(session *mcpe.NetworkSession) mcpe.PacketHandler {
		return NewPreSpawnPacketHandler(session)
	}
	mcpe.NewInGamePacketHandler = func(session *mcpe.NetworkSession) mcpe.PacketHandler {
		return NewInGamePacketHandler(session)
	}
}

// PreSpawnPacketHandler is a port of pocketmine\network\mcpe\handler\PreSpawnPacketHandler: sends
// StartGame and everything else the client needs before it can spawn.
//
// handleRequestChunkRadius isn't needed: gophertunnel's Conn.StartGame answers RequestChunkRadius
// itself and waits for the client to finish spawning (see NetworkSession.onClientSpawnResponse,
// which applies the requested view distance).
type PreSpawnPacketHandler struct {
	session *mcpe.NetworkSession
	server  mcpe.Server
	player  *player.Player
}

func NewPreSpawnPacketHandler(session *mcpe.NetworkSession) *PreSpawnPacketHandler {
	return &PreSpawnPacketHandler{session: session, server: session.GetServer(), player: session.GetPlayer()}
}

// SetUp is a port of PreSpawnPacketHandler::setUp.
func (h *PreSpawnPacketHandler) SetUp() error {
	h.server.Lock()
	data := h.startGameData()
	h.server.Unlock()

	h.session.GetLogger().Debug("Sending start game packet")
	// StartGame + ItemRegistry, then blocks until the client has spawned. The server lock isn't
	// held meanwhile, so the world keeps ticking.
	if err := h.session.GetConn().StartGame(data); err != nil {
		return err
	}

	h.server.Lock()
	defer h.server.Unlock()
	session, p := h.session, h.player

	session.GetLogger().Debug("Sending actor identifiers")
	session.SendDataPacket(bedrock.AvailableActorIdentifiers())

	session.GetLogger().Debug("Sending biome definitions")
	session.SendDataPacket(bedrock.BiomeDefinitionList())

	session.GetLogger().Debug("Sending attributes")
	session.SendDataPacket(entity.SyncAttributesPacket(p.GetID(), p.GetAttributeMap().GetAll()))

	session.GetLogger().Debug("Sending available commands")
	// syncAvailableCommands: the command map isn't wired to players yet, so the list is empty.
	session.SendDataPacket(&packet.AvailableCommands{})

	session.GetLogger().Debug("Sending abilities")
	session.SyncAbilities(p)
	session.SyncAdventureSettings()

	session.GetLogger().Debug("Sending effects")
	for _, effect := range p.GetEffects().All() {
		session.SendDataPacket(entity.EntityEffectAddedPacket(p.GetID(), effect, false))
	}

	session.GetLogger().Debug("Sending actor metadata")
	p.SendData([]world.EntityViewer{p}, nil)

	session.GetLogger().Debug("Sending inventory")
	session.SyncAllInventories()
	session.SyncSelectedHotbarSlot()

	session.GetLogger().Debug("Sending creative inventory data")
	// InventoryManager::syncCreative: the 1.26.50 creative inventory (see bedrock.CreativeContent).
	session.SendDataPacket(bedrock.CreativeContent())

	session.GetLogger().Debug("Sending crafting data")
	// CraftingDataCache::getCache: the 1.26.50 recipes (see bedrock.CraftingData). CraftingManager
	// isn't ported, so crafting itself doesn't work yet.
	session.SendDataPacket(bedrock.CraftingData())

	session.GetLogger().Debug("Sending player list")
	session.SyncPlayerList(h.server.GetOnlinePlayers())
	return nil
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
		WorldGameMode:   h.server.GetDefaultGameMode(),
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

// HandleDataPacket handles nothing: see PreSpawnPacketHandler's doc comment.
func (h *PreSpawnPacketHandler) HandleDataPacket(pk packet.Packet) bool { return false }

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
