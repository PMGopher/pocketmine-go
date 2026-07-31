package timings

import (
	"fmt"
	"sync"

	"pocketmine-go/pocketmine/scheduler"
)

// Timings is a port of the portable subset of pocketmine\timings\Timings: the named handler
// registry.
//
// Deliberately NOT ported (yet): GetEntityTimings, GetTileEntityTimings, Get*DataPacketTimings,
// GetEventTimings/GetEventHandlerTimings, GetAsyncTask*Timings — all keyed by a concrete PHP
// class (Entity/Tile/ServerboundPacket/ClientboundPacket/Event/AsyncTask) that doesn't exist yet
// in this port. Once those packages land, each can be added the same way GetCommandDispatchTimings
// and GetScheduledTaskTimings already are here (a map keyed by name/reflect.Type, lazily filled).
var (
	FullTick                               *TimingsHandler
	ServerTick                             *TimingsHandler
	ServerInterrupts                       *TimingsHandler
	MemoryManager                          *TimingsHandler
	GarbageCollector                       *TimingsHandler
	TitleTick                              *TimingsHandler
	PlayerNetworkSend                      *TimingsHandler
	PlayerNetworkSendCompress              *TimingsHandler
	PlayerNetworkSendCompressBroadcast     *TimingsHandler
	PlayerNetworkSendCompressSessionBuffer *TimingsHandler
	PlayerNetworkSendEncrypt               *TimingsHandler
	PlayerNetworkSendInventorySync         *TimingsHandler
	PlayerNetworkSendPreSpawnGameData      *TimingsHandler
	PlayerNetworkReceive                   *TimingsHandler
	PlayerNetworkReceiveDecompress         *TimingsHandler
	PlayerNetworkReceiveDecrypt            *TimingsHandler
	PlayerChunkOrder                       *TimingsHandler
	PlayerChunkSend                        *TimingsHandler
	Connection                             *TimingsHandler
	Scheduler                              *TimingsHandler
	ServerCommand                          *TimingsHandler
	PermissibleCalculation                 *TimingsHandler
	PermissibleCalculationDiff             *TimingsHandler
	PermissibleCalculationCallback         *TimingsHandler
	EntityMove                             *TimingsHandler
	EntityMoveCollision                    *TimingsHandler
	ProjectileMove                         *TimingsHandler
	ProjectileMoveRayTrace                 *TimingsHandler
	PlayerCheckNearEntities                *TimingsHandler
	EntityBaseTick                         *TimingsHandler
	LivingEntityBaseTick                   *TimingsHandler
	ItemEntityBaseTick                     *TimingsHandler

	SchedulerSync  *TimingsHandler
	SchedulerAsync *TimingsHandler

	PlayerCommand            *TimingsHandler
	CraftingDataCacheRebuild *TimingsHandler

	SyncPlayerDataLoad *TimingsHandler
	SyncPlayerDataSave *TimingsHandler

	BroadcastPackets *TimingsHandler
	PlayerMove       *TimingsHandler

	asyncTaskProgressUpdateParent *TimingsHandler
	asyncTaskCompletionParent     *TimingsHandler
	asyncTaskErrorParent          *TimingsHandler
	AsyncTaskWorkers              *TimingsHandler

	initOnce sync.Once

	pluginTaskTimingMapMu sync.Mutex
	pluginTaskTimingMap   = map[string]*TimingsHandler{}

	commandTimingMapMu sync.Mutex
	commandTimingMap   = map[string]*TimingsHandler{}
)

// Init is a port of Timings::init(). Called lazily by the getters below, matching the PHP
// original's self::init() guards.
func Init() {
	initOnce.Do(func() {
		FullTick = NewTimingsHandler("Full Server Tick", nil, "")
		ServerTick = NewTimingsHandler("Server Tick Update Cycle", FullTick, "")
		ServerInterrupts = NewTimingsHandler("Server Mid-Tick Processing", FullTick, "")
		MemoryManager = NewTimingsHandler("Memory Manager", nil, "")
		GarbageCollector = NewTimingsHandler("Garbage Collector", MemoryManager, "")
		TitleTick = NewTimingsHandler("Console Title Tick", nil, "")

		Connection = NewTimingsHandler("Connection Handler", nil, "")

		PlayerNetworkSend = NewTimingsHandler("Player Network Send", Connection, "")
		PlayerNetworkSendCompress = NewTimingsHandler("Player Network Send - Compression", PlayerNetworkSend, "")
		PlayerNetworkSendCompressBroadcast = NewTimingsHandler("Player Network Send - Compression (Broadcast)", PlayerNetworkSendCompress, "")
		PlayerNetworkSendCompressSessionBuffer = NewTimingsHandler("Player Network Send - Compression (Session Buffer)", PlayerNetworkSendCompress, "")
		PlayerNetworkSendEncrypt = NewTimingsHandler("Player Network Send - Encryption", PlayerNetworkSend, "")
		PlayerNetworkSendInventorySync = NewTimingsHandler("Player Network Send - Inventory Sync", PlayerNetworkSend, "")
		PlayerNetworkSendPreSpawnGameData = NewTimingsHandler("Player Network Send - Pre-Spawn Game Data", PlayerNetworkSend, "")

		PlayerNetworkReceive = NewTimingsHandler("Player Network Receive", Connection, "")
		PlayerNetworkReceiveDecompress = NewTimingsHandler("Player Network Receive - Decompression", PlayerNetworkReceive, "")
		PlayerNetworkReceiveDecrypt = NewTimingsHandler("Player Network Receive - Decryption", PlayerNetworkReceive, "")

		BroadcastPackets = NewTimingsHandler("Broadcast Packets", PlayerNetworkSend, "")

		PlayerMove = NewTimingsHandler("Player Movement", nil, "")
		PlayerChunkOrder = NewTimingsHandler("Player Order Chunks", nil, "")
		PlayerChunkSend = NewTimingsHandler("Player Network Send - Chunks", PlayerNetworkSend, "")
		Scheduler = NewTimingsHandler("Scheduler", nil, "")
		ServerCommand = NewTimingsHandler("Server Command", nil, "")
		PermissibleCalculation = NewTimingsHandler("Permissible Calculation", nil, "")
		PermissibleCalculationDiff = NewTimingsHandler("Permissible Calculation - Diff", PermissibleCalculation, "")
		PermissibleCalculationCallback = NewTimingsHandler("Permissible Calculation - Callbacks", PermissibleCalculation, "")

		SyncPlayerDataLoad = NewTimingsHandler("Player Data Load", nil, "")
		SyncPlayerDataSave = NewTimingsHandler("Player Data Save", nil, "")

		EntityMove = NewTimingsHandler("Entity Movement", nil, "")
		EntityMoveCollision = NewTimingsHandler("Entity Movement - Collision Checks", EntityMove, "")

		ProjectileMove = NewTimingsHandler("Projectile Movement", EntityMove, "")
		ProjectileMoveRayTrace = NewTimingsHandler("Projectile Movement - Ray Tracing", ProjectileMove, "")

		PlayerCheckNearEntities = NewTimingsHandler("checkNearEntities", nil, "")
		EntityBaseTick = NewTimingsHandler("Entity Base Tick", nil, "")
		LivingEntityBaseTick = NewTimingsHandler("Entity Base Tick - Living", nil, "")
		ItemEntityBaseTick = NewTimingsHandler("Entity Base Tick - ItemEntity", nil, "")

		SchedulerSync = NewTimingsHandler("Scheduler - Sync Tasks", nil, "")

		SchedulerAsync = NewTimingsHandler("Scheduler - Async Tasks", nil, "")
		asyncTaskProgressUpdateParent = NewTimingsHandler("Async Tasks - Progress Updates", SchedulerAsync, "")
		asyncTaskCompletionParent = NewTimingsHandler("Async Tasks - Completion Handlers", SchedulerAsync, "")
		asyncTaskErrorParent = NewTimingsHandler("Async Tasks - Error Handlers", SchedulerAsync, "")

		AsyncTaskWorkers = NewTimingsHandler("Async Task Workers", nil, "")

		PlayerCommand = NewTimingsHandler("Player Command", nil, "")
		CraftingDataCacheRebuild = NewTimingsHandler("Build CraftingDataPacket Cache", nil, "")
	})
}

// GetScheduledTaskTimings is a port of Timings::getScheduledTaskTimings().
func GetScheduledTaskTimings(task *scheduler.TaskHandler, period int) *TimingsHandler {
	Init()
	name := "Task: " + task.GetTaskName()
	if period > 0 {
		name += fmt.Sprintf("(interval:%d)", period)
	} else {
		name += "(Single)"
	}

	pluginTaskTimingMapMu.Lock()
	defer pluginTaskTimingMapMu.Unlock()
	h, ok := pluginTaskTimingMap[name]
	if !ok {
		h = NewTimingsHandler(name, SchedulerSync, task.GetOwnerName())
		pluginTaskTimingMap[name] = h
	}
	return h
}

// GetCommandDispatchTimings is a port of Timings::getCommandDispatchTimings().
func GetCommandDispatchTimings(commandName string) *TimingsHandler {
	Init()
	commandTimingMapMu.Lock()
	defer commandTimingMapMu.Unlock()
	h, ok := commandTimingMap[commandName]
	if !ok {
		h = NewTimingsHandler("Command - "+commandName, nil, "")
		commandTimingMap[commandName] = h
	}
	return h
}
