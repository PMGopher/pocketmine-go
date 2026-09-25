package player

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

	"pocketmine-go/pocketmine/command"
	"pocketmine-go/pocketmine/entity"
	"pocketmine-go/pocketmine/form"
	"pocketmine-go/pocketmine/inventory"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/log"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
	"pocketmine-go/pocketmine/permission"
	"pocketmine-go/pocketmine/world"
)

// Server is what Player needs from pocketmine\Server (PHP's $this->server). It's declared here so
// the server package, which imports this one, satisfies it without an import cycle.
type Server interface {
	command.Server

	GetLogger() log.Logger
	// BroadcastMessage is a port of Server::broadcastMessage: message is a string or
	// *lang.Translatable; nil recipients means the BROADCAST_CHANNEL_USERS subscribers.
	BroadcastMessage(message any, recipients []command.Sender) int
	// DispatchCommand is a port of Server::dispatchCommand.
	DispatchCommand(sender command.Sender, commandLine string, internal bool) bool
	SubscribeToBroadcastChannel(channelID string, subscriber command.Sender)
	UnsubscribeFromBroadcastChannel(channelID string, subscriber command.Sender)
	UnsubscribeFromAllBroadcastChannels(subscriber command.Sender)

	GetTick() int64
	GetOnlinePlayers() []*Player
	RemoveOnlinePlayer(p *Player)
	SaveOfflinePlayerData(name string, data *nbt.CompoundTag)
	GetWorldManager() *world.WorldManager

	IsOp(name string) bool
	GetNameBans() *permission.BanList
	IsHardcore() bool
	GetGamemode() GameMode
	GetForceGamemode() bool
	// GetAllowedViewDistance is a port of Server::getAllowedViewDistance.
	GetAllowedViewDistance(distance int) int
	// GetPropertyInt is ServerConfigGroup::getPropertyInt (pocketmine.yml).
	GetPropertyInt(variable string, defaultValue int) int
	// GetConfigBool is ServerConfigGroup::getConfigBool (server.properties).
	GetConfigBool(variable string, defaultValue bool) bool
}

// NetworkSession is what Player needs from pocketmine\network\mcpe\NetworkSession (PHP's
// $this->networkSession).
type NetworkSession interface {
	SendDataPacket(pk packet.Packet)
	IsConnected() bool
	GetIp() string
	GetPort() int

	OnChatMessage(message any)
	OnJukeboxPopup(message any)
	OnPopup(message string)
	OnTip(message string)
	OnTitle(title string)
	OnSubTitle(subtitle string)
	OnActionBar(actionBar string)
	OnClearTitle()
	OnResetTitleOptions()
	OnTitleDuration(fadeIn, stay, fadeOut int)
	OnToastNotification(title, body string)
	OnFormSent(id int, f form.Form) bool
	OnCloseAllForms()
	OnOpenSignEditor(signPosition math.Vector3, frontSide bool)
	OnItemCooldownChanged(it item.Item, ticks int)

	SyncViewAreaRadius(distance int)
	SyncViewAreaCenterPoint(pos math.Vector3, viewDistance int)
	SyncPlayerSpawnPoint(newSpawn math.Vector3)
	SyncGameMode(mode GameMode, isRollback bool)
	SyncAbilities(p *Player)
	SyncAdventureSettings()
	// SyncMovement is NetworkSession::syncMovement: nil yaw/pitch keep the player's own.
	SyncMovement(pos math.Vector3, yaw, pitch *float64, mode byte)
	OnEnterWorld()

	// StartUsingChunk is NetworkSession::startUsingChunk: the chunk is sent, then onCompletion
	// is called.
	StartUsingChunk(chunkX, chunkZ int, onCompletion func())
	StopUsingChunk(chunkX, chunkZ int)
	NotifyTerrainReady()

	OnServerDeath(deathMessage any)
	OnServerRespawn()
	OnPlayerDestroyed(reason, disconnectScreenMessage any)
	Transfer(ip string, port int, reason any)
	DisconnectWithError(reason, disconnectScreenMessage any)

	// GetInvManager is NetworkSession::getInvManager (nil before spawning and after
	// disconnecting).
	GetInvManager() InventoryManager
}

// InventoryManager is what Player needs from pocketmine\network\mcpe\InventoryManager.
type InventoryManager interface {
	inventory.InventorySyncer
	OnCurrentWindowChange(inv inventory.Inventory)
	OnCurrentWindowRemove()
	SyncCreative()
}

// Info is pocketmine\player\PlayerInfo as Player uses it: *PlayerInfo and *XboxLivePlayerInfo.
type Info interface {
	GetUsername() string
	GetUUID() string
	GetLocale() string
	GetExtraData() map[string]any
	GetSkin() *entity.Skin
}
