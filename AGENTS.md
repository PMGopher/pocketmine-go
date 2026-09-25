# AGENTS.md

Context for anyone (human or AI agent) picking up work on **pocketmine-go**. Read this before
changing code. The feature checklist lives in [README.md](README.md#feature-checklist). This file
covers the goal, how the code is organised, the conventions, what state things are in, and what to
do next.

_Last updated: 2026-09-25. Upstream reference: PocketMine-MP `5.44.4` (commit `6a7cc02`, July 2026)._

---

## 1. Goal

Rewrite [PocketMine-MP](https://github.com/pmmp/PocketMine-MP) (a Minecraft: Bedrock Edition
server written in PHP) in Go, **keeping its game logic faithful to the original**.

- This is a **port**, not a new server design. Each Go type maps to a PHP class, and behaviour
  (block drops, break times, tick logic, generator noise, light propagation, ...) should match
  what PocketMine-MP does. When in doubt, read the PHP source and copy its behaviour, not your
  guess of what Minecraft does.
- **Exception: the network/protocol layer.** We do **not** port `pocketmine/network/mcpe/protocol`,
  RakLib, compression or encryption. We use
  [gophertunnel](https://github.com/sandertv/gophertunnel) (the library Dragonfly uses) for RakNet,
  login, the encryption handshake, resource-pack negotiation and every packet type. PocketMine-MP's
  own code *above* the wire (NetworkSession, packet handlers, InventoryManager, TypeConverter, ...)
  still needs porting, but on top of gophertunnel.
- **Out of scope** (PHP-runtime specific, no Go equivalent needed): `thread/`, pthreads channels in
  `network/mcpe/raklib`, `PharPluginLoader`, `GarbageCollectorManager`/`MemoryDump`, `updater/`,
  `stats/`. Go's goroutines replace the thread/async-worker model (see §6).

## 2. Build, run, test

```bash
go build ./...                      # builds everything
go test ./...                       # all packages currently pass
go run ./cmd/pocketmine-go                                   # server on UDP :19132, data in the current directory
go run ./cmd/pocketmine-go --data=srv --server-port=19133 --xbox-auth=false   # PocketMine.php options + server.properties/pocketmine.yml overrides
go run ./cmd/pocketmine-go --debug.level=2    # debug log lines (pocketmine.yml debug.level)
```

- Go version: see `go.mod` (`go 1.26.1`).
- Client protocol: whatever the pinned gophertunnel supports (v1.62.0 → Bedrock **1.26.50**,
  protocol 2193). This is ahead of upstream PocketMine-MP 5.44.4 (1.26.30).
- Bumping gophertunnel means updating the vendored assets in `pocketmine/data/bedrock/assets` to
  the same client version: block runtime IDs are positions in `canonical_block_states.nbt`, so an
  outdated palette makes every block after the first new one render wrong. pmmp/BedrockData has no
  1.26.50 release yet, so the 1.26.50 block palette, item list and `data_driven_blocks.nbt` come
  from df-mc/dragonfly (MIT, `assets/LICENSE-dragonfly`; same format). Replace them with
  BedrockData once pmmp publishes 1.26.50 data. Since 1.26.50 the client also needs the
  data-driven vanilla block definitions in StartGame (`bedrock.DataDrivenBlocks`).
- Chunks are sent in **sub-chunk request mode** (`pocketmine/network/mcpe/sub_chunk_request.go`): LevelChunk
  carries only biomes + `SubChunkLimit`, and `SubChunkRequest` is answered with `SubChunk`
  entries (sub-chunk format **version 9**). This matches the vanilla server and Dragonfly on
  1.26.50. PocketMine-MP 5.44.4 still sends full chunks with version 8 sub-chunks, but it only
  supports up to 1.26.30. A 1.26.51 client disconnected with "Block" (ClientDisconnection-90)
  while we sent full chunks with v8 sub-chunks; the switch to Dragonfly's proven path is the fix
  for that (not yet confirmed with a real client at the time of writing).
- Clients that enable the **client blob cache** (`ClientCacheStatus`, e.g. the Windows client) get
  their chunks through it (`network/mcpe/client_blob_cache.go`, as Dragonfly does): biomes and
  sub-chunks are sent as xxHash64 hashes and `ClientCacheBlobStatus` is answered with
  `ClientCacheMissResponse`. A 1.26.51 Windows client that played fine on Dragonfly disconnected
  with "Block" right after loading when our request-mode chunks came without the cache.
- The skin sent in `PlayerList` is built like Dragonfly's (`network/mcpe/client_skin.go`): the
  client's own login skin with persona flag, PlayFab ID and a compact geometry-only resource patch,
  no persona pieces/tints, `BuildPlatform` -1. PHP sends `LegacySkinAdapter::toSkinData`
  instead; the 1.26.51 client still left with "Block" right after loading while we did. On join the
  server logs `Client skin: persona=...` (debug), which also shows which skin the client used.
- StartGame must set `BaseGameVersion` and server-authoritative block breaking
  (`PlayerMovementSettings(0, true)` like PreSpawnPacketHandler).
- After StartGame, `PreSpawnPacketHandler.SetUp` sends the rest of PreSpawnPacketHandler's packets:
  `AvailableActorIdentifiers` (pmmp BedrockData 1.26.30 `entity_identifiers.nbt`),
  `BiomeDefinitionList` (1.26.50, captured from Dragonfly: `assets/biome_definitions.bin`, see
  `bedrock.BiomeDefinitionList`), `AvailableCommands` (empty), the player's own `SetActorData`,
  `PlayerHotBar` and an empty `CraftingData`. These were missing before.
- The 1.26.50 block palette was cross-checked against pmmp's official 1.26.30 file: Dragonfly's
  state order matches it on all 676 unchanged multi-state blocks (an altay/BedrockData 1.26.50 dump
  disagrees on 12, so don't use that one), and the 98 data-driven block definitions match altay's.
- `xbox-auth` is **on** by default, like PocketMine-MP. Test clients that can't sign in (e.g. a
  gophertunnel Dialer) need `--xbox-auth=false`.
- The data folder (`--data`, default: working directory) holds `server.properties`,
  `worlds/<level-name>/` (LevelDB + `level.dat`) and `players/<name>.dat`. Delete `worlds/` to
  regenerate.
- gophertunnel's own errors go to the server log (`ErrorLog`, see `server/slog_handler.go`), and a
  client's `PacketViolationWarning` is logged as a warning: check those first when a client
  disconnects. When the client closes the connection by itself (it sends no reason), the session
  logs its last ~80 sent/received packets (`network/mcpe/packet_trace.go`).
- gophertunnel's `Conn.StartGame` sends `PlayStatus(PLAYER_SPAWN)` as soon as the client asks for
  its chunk radius, while PocketMine-MP only sends it after `spawnThreshold` chunks are sent. To
  make up for it, a new world's spawn area is pre-generated at startup (radius 8, like
  `WorldManager::generateWorld`'s background generation; see `Server.generateSpawnTerrain`) and
  the spawn-threshold chunks are sent immediately on the spawn response
  (`NetworkSession.sendSpawnTerrain`). Before this, the first chunk arrived ~1 s after the client
  had already entered an empty world, and it disconnected ("Block") without ever requesting a
  sub-chunk (seen in the packet trace).
- `NetworkChunkPublisherUpdate` (NetworkSession::syncViewAreaCenterPoint) is only sent from
  `Player.OrderChunks`, which only runs when `nextChunkOrderRun` is due (`Player.DoChunkRequests`,
  like PHP: after a view distance change, a teleport, a changed chunk, and at most every 20 ticks
  while moving). It used to be sent after every 4-chunk batch, i.e. every tick; a real 1.26.51
  client with the blob cache enabled then never answered a single LevelChunk (no
  ClientCacheBlobStatus, no SubChunkRequest) and left after ~2 s.
- **Debugging a real client:** `tools/packetproxy` (gophertunnel's proxy example plus logging, not
  part of the port) sits between the game and a server started with `--xbox-auth=false` and
  writes every packet in both directions to `packets.log`. Run the same session through it
  against this server and against Dragonfly to see exactly what the client does differently.
- Go embedding has no virtual dispatch: a base `Entity`/`Living`/`Human` method that PHP calls as
  `$this->x()` must call `e.self.x()` / `l.lself.x()` / `h.hself.x()` when a subclass (usually
  `Player`) overrides `x`. This was missed for `sendData` (a player never got its own metadata
  updates), `setMotion` (no knockback on the hit player's own client), `broadcastSound`,
  `broadcastAnimation` and `canEat`. Check this for every new override.

## 3. Repository layout

```
cmd/pocketmine-go/     Entry point only (port of PocketMine.php): options, server.lock, MainLogger, then
                       server.New/Start.
tools/packetproxy/     Debugging proxy that logs a real client's packets (not part of the port).
pocketmine/            One Go package per PHP namespace under pmmp/PocketMine-MP/src/.
  server/              Server, ServerProperties, ServerConfigGroup (PHP's root-namespace classes;
                       own package because the root package is imported by entity/world)
  block/               ~260 of 270 block classes, tile/ (tiles), inventory/ (block inventories), utils/
  item/                ~110 of 136 item classes, VanillaItems (partial)
  world/               World, WorldManager, Explosion, tick loop, ChunkListener
    format/            Chunk, SubChunk, PalettedBlockArray, Height/LightArray
    format/io/         level.dat (WorldData), leveldb/ (chunk read/write)
    generator/         Flat, Normal, hell (Nether), noise, populators, trees/ores, biomeselector
    light/             sky + block light propagation
    biome/, particle/, sound/, utils/
  network/             Network, NetworkSessionManager, bandwidth stats; query/, upnp/
  network/mcpe/        NetworkSession (on gophertunnel's minecraft.Conn), InventoryManager, ChunkCache,
                       CreativeInventoryCache, broadcasters, rate limiter, sub-chunk requests
  network/mcpe/raklib/ RakLibInterface: gophertunnel listener + go-raknet socket (login phase, query
                       packets, IP blocks)
  network/mcpe/handler/     PreSpawn/InGame/Death packet handlers, ItemStackRequestExecutor
  network/mcpe/convert/     BlockTranslator, BlockStateSerializer, ItemTranslator, TypeConverter parts,
                            ClientDataToSkinDataHelper
  network/mcpe/serializer/  ChunkSerializer (LevelChunk payload)
  data/bedrock/        BlockStateDictionary, ItemTypeDictionary + vendored assets/
                       (canonical_block_states.nbt, required_item_list.json)
  data/runtime/        Runtime state bit-packing (describer/reader/writer)
  player/              Player, GameMode, chunk streaming, SurvivalBlockBreakHandler, combat,
                       fall damage, PlayerInfo, OfflinePlayer, DatFilePlayerDataProvider
  entity/              Entity, Living, Human, HungerManager, ExperienceManager, attributes,
                       EntityFactory, Skin, mobs (Zombie, Villager, Squid)
    effect/            All vanilla effects, EffectManager, EffectInstance (+ EntityEffect*Event,
                       kept here to avoid an import cycle)
    object/            ItemEntity, FallingBlock, PrimedTNT, ExperienceOrb, Painting, EndCrystal, ...
    projectile/        Arrow, Snowball, Egg, EnderPearl, SplashPotion, Trident, ...
    animation/, utils/ Entity animations (network), ExperienceUtils
  event/entity/        Every pocketmine\event\entity event (damage, death, explode, projectile, ...)
  event/block, player, inventory, world, server, plugin/   Every concrete event of those namespaces
  item/enchantment/    All vanilla enchantments, protection/sharpness/knockback/fire aspect logic
  command/ (+ defaults/), console/, event/, permission/, plugin/, scheduler/, lang/, timings/,
  log/, promise/, resourcepacks/, form/
  math/, nbt/, binaryutils/, color/, utils/   (ports of pmmp's math/nbt/binaryutils/color libs)
```

Go file names are the PHP class name in `snake_case` (`BaseBigDripleaf.php` →
`base_big_dripleaf.go`), in the package matching the PHP namespace. Keep that so the two trees can
be diffed mechanically (see §7).

## 4. Porting conventions (follow these)

1. **Doc comment points at the source.** Every ported type starts with
   `// X is a port of pocketmine\...\X.` Functions that port one specific PHP method say so
   (`// ... is a port of World::getSafeSpawn`).
2. **Document gaps where they are.** If part of a PHP method can't be ported yet because a
   dependency doesn't exist, port what you can and leave a comment saying exactly what's missing
   and why (e.g. `block/melon.go`: drops return nil until FortuneDropHelper + item construction
   exist). Grep for `unported`, `not yet ported`, `for now`, `isn't wired` to find these (~330 today).
   Don't invent behaviour to fill a gap.
3. **PHP inheritance → embedding + `self`.** PHP's `Block` base class calls overridable methods on
   `$this`. Go embedding doesn't dispatch virtually, so concrete blocks call `Init(self)` in their
   constructor and base methods call `b.self.X()`. See the long comment on `block.Behavior`
   ([pocketmine/block/behavior.go](pocketmine/block/behavior.go)). Every concrete block also
   implements `Clone()` and calls `rebind`. Copy an existing block (e.g. `melon.go`) when adding one.
4. **PHP traits → embedded helper structs** (e.g. `CancellableTrait`, `BlockInventoryTrait`).
5. **Registries** (`VanillaBlocks`, `VanillaItems`) are lazily built singletons whose getters
   return a `Clone()`. Break info, tags, etc. are copied from `VanillaBlocksInputs.php` /
   `VanillaItemsInputs.php`, not guessed.
6. **Break import cycles with small interfaces** (e.g. `block.Item`, `block.Player`, `block.World`
   are interfaces the `block` package declares and `item`/`player`/`world` satisfy). Don't merge
   packages to get around a cycle.
7. **Tests.** Each ported feature gets a `_test.go` next to it that checks behaviour against what
   the PHP code does. `go test ./...` must stay green.
8. **Commits** are small and describe what was ported (`Port X`, `Port X::y's branch`, `Wire X into
   cmd/pocketmine-go`). One logical unit per commit.

## 5. Current status (summary)

Measured by mapping every PHP class in upstream `src/` to a Go file/type (see §7):
**roughly 750–850 of 1,498 PHP classes (~50–57%)** have a Go counterpart. Coverage is very uneven:

| Area | State |
|---|---|
| block (+ tiles, block inventories, utils) | ~87% of classes ported. **Only ~55 vanilla block singletons are registered and ~15 have network serializers**, so only a handful can actually appear in-game. |
| item | 133/154 classes incl. enchantments. ~85 vanilla item singletons. **Item NBT (de)serialization isn't ported** (blocks saving dropped items/tridents and Human inventories). |
| world core | World, ticking, scheduled/random updates, light, explosions, WorldManager (incl. `worlds:` from pocketmine.yml), ChunkListener, level.dat, LevelDB save/load: done. |
| generators | Flat, Normal (all biomes), Nether done. Trees: only oak/spruce/birch. Generation is synchronous (no async executor). |
| entity | **All 77 `pocketmine\entity` classes ported with their logic.** Item use (bows, throwables, spawn eggs) is wired through the packet handlers, limited by which items have network mappings. |
| player | **Complete** (`Player` with every PHP method that has its dependencies: chunk streaming, block interaction, item use, combat, death/respawn, forms, titles, game modes, permissions, broadcast channels, player data). |
| event | **Every concrete event** (block, entity, player, inventory, world, server, plugin), fired where the ported code fires them. Event inheritance: `event.DeclareParent` (see `event/parents.go`). |
| network | **Complete above the wire** except crafting data: `NetworkSession`, `RakLibInterface` (on gophertunnel), `InventoryManager`, `ChunkCache`, `CreativeInventoryCache`, broadcasters, rate limiter, `PreSpawn`/`InGame`/`Death` packet handlers, `ItemStackRequestExecutor`, query, UPnP, resource packs, `DataPacket*Event`. |
| server core | **Complete except crash dumps**: `Server` (pocketmine.yml, language, ops/whitelist/bans, broadcast channels, TPS, title tick, query regeneration, memory manager, async pool, shutdown), console reader/sender, `MainLogger` + `server.log`, `server.lock`. |
| command | Command map + **34 of 41 default commands** (`command/defaults`). Missing: give, clear, enchant, effect (need `StringToItemParser`/string-to-enchantment/effect parsers), particle, timings, dumpmemory. |
| not started | crafting (`crafting/*`, CraftingManager, crafting/enchanting transactions: craft requests fail like unknown recipes), plugin loading, crash dumps, region (Anvil/McRegion) world formats, block-state upgrader (old world compatibility), item NBT serialization. |

### How gophertunnel changes the login/spawn flow

gophertunnel runs RakNet, login, encryption and resource packs itself, so the PHP handlers for those
phases (SessionStart, Login, Handshake, ResourcePacks) have no Go counterpart. Their game logic lives
where gophertunnel calls back into us (`network/mcpe/raklib/rak_lib_interface.go`):
`ListenConfig.Allow` builds the `PlayerInfo` and fires `PlayerPreLoginEvent` (whitelist, bans,
server full); `FetchResourcePacks` fires `PlayerResourcePackOfferEvent`; after `Accept`,
`NetworkSession.Login` runs `setAuthenticationStatus` (duplicate login, XUID checks) and creates
the player. `PreSpawnPacketHandler.SetUp` calls `Conn.StartGame`, which answers
`RequestChunkRadius` and sends `PlayStatus(PLAYER_SPAWN)` itself; the session then sends the spawn
chunks at once (`NetworkSession.OnClientRequestChunkRadius`) and `NotifyTerrainReady` completes the
spawn. The server lock is released while `StartGame` waits for the client.

With `xbox-auth=false`, gophertunnel doesn't verify identity tokens, so every client is treated as
unauthenticated and its XUID is discarded (PHP would still keep a verified XUID).

### What a player can do today

Connect (online or offline mode), spawn in a generated Normal world, walk around with chunks
streaming in, break and place blocks (those with network mappings), use items, move items in the
inventory, open containers, use the creative inventory (items with network mappings), chat, run
commands (op/whitelist/ban/gamemode/tp/time/...), see other players, PvP, take fall damage, die and
respawn, regenerate health from food, lose hunger. The world and players save on shutdown and on
autosave; `stop` or Ctrl+C shuts down cleanly.

### What a player cannot do yet

Craft, smelt, brew or enchant; hold or place blocks/items that have no network mapping; use plugins.

### Known issues

- **Floating up / flying after spawn.** Reported by a tester (2026-09-24): right after spawning the
  player drifted up into the sky. Commit `697dfa7` fixed a related bug (client-requested flight was
  never denied, so the swim-up gesture enabled flying). **Probable cause found (not yet confirmed
  in a client):** `StartGame` was sent the *feet* position and `PlayerAuthInput`'s position (the
  client's *eye* position) was stored as the feet position, so every player was 1.62 blocks too
  high server-side (and to other players). `PreSpawnPacketHandler` now sends `Human::getOffsetPosition` (feet +
  1.621) in `StartGame` and subtracts 1.62 from `PlayerAuthInput`, like PHP. If the bug is still
  seen, check `UpdateAbilities` ordering and leftover `MayFly`/`Flying` state.
- Movement is still client-authoritative: `Player.HandleMovement` (port of
  `Player::handleMovement`) only rejects moves > 15 blocks per tick; no anti-fly.
- Blocks the generator can place but that have no network serializer can't be sent to the client.
  Keep `server.knownBlocks` and `convert/vanilla_block_mappings.go` in sync until the full
  mappings are ported.
- Population writes behave like PHP's async `PopulationTask` (`World.populationWrites`): no
  neighbour updates or changed-block tracking, then `setChunk` semantics (`OnChunkChanged`).
  Before this, populated ores lit up and leaves kept re-checking decay forever.
- Item NBT isn't serialized, so dropped items (`ItemEntity`) and tridents don't save with the
  chunk (`CanSaveWithChunk` returns false), and Human inventories aren't saved in entity NBT.
- Architecture note for the entity port: `world.Entity` is the polymorphic entity interface;
  PHP's `$this` virtual calls go through the exported `entity.Hooks`/`LivingHooks`/`HumanHooks`
  interfaces (same idea as `block.Behavior`). Packages below `entity` get behaviour through small
  function hooks set in `init()` (e.g. `block.SpawnFallingBlockFunc`, `world.DropItemFunc`,
  `world.LoadEntityFunc`); grep for `Func =` to find them.
- Worlds created by vanilla Bedrock or PocketMine-MP (PHP) will mostly fail to load: only the
  block states this port knows are recognised, and there is no block-state upgrader.
- The creative inventory is now built from the core `CreativeInventory`
  (`CreativeInventoryCache`, like PHP) instead of the vendored 1.26.50 packet, so it only shows
  items with a network mapping until the item/block mappings are complete (the network IDs have to
  match the core inventory, or creative item requests would create the wrong item).
- Threading: packet handling and the tick share `Server`'s lock (`mcpe.Server` embeds
  `sync.Locker`). Code called from a packet handler or the tick already holds it; never call
  `Server.Lock` from there. `Server.Shutdown` is safe either way (the `stop` command calls it from
  inside the tick).

## 6. Plan / roadmap

Ordered by what gets us to a **playable survival server** fastest. Each phase should end with
something a person can see working in the client.

### Phase 1: Make the existing world playable
1. **All block ↔ network mappings.** Port `data/bedrock/block/convert/*` in full
   (`BlockObjectToStateSerializer`, `BlockStateToObjectDeserializer`, `BlockStateReader/Writer`,
   property helpers) and the whole `VanillaBlockMappings`. Then register every block in
   `VanillaBlocks`. This unblocks everything else visual.
2. **Finish `VanillaBlocks` / `VanillaItems`** from the `*Inputs.php` files, plus
   `StringToItemParser` (needed for `/give` and plugins).
3. **Held item + hotbar**: handle `MobEquipment`, replace `bareHandItem`.
4. **Block placing** via `InventoryTransaction` `UseItem` (click-block) → `World.UseItemOn`.
5. **Chat broadcast** (`Text` packet → all players, with the `chat` formatters already ported).
6. **Investigate the spawn/floating issue** above.

### Phase 2: Real server structure
Done: `Server` (pocketmine.yml, ops/whitelist/bans, broadcast channels, async pool, memory
manager, query, UPnP), `NetworkSession`/`RakLibInterface`/`InventoryManager` and every packet
handler above the wire, console reader + `ConsoleCommandSender`, all concrete events, permissions
and ops wired to players, 34 of 41 default commands. Remaining: `give`, `clear`, `enchant`,
`effect` (need the string-to-item/enchantment/effect parsers), `particle`, `timings` (report
upload), crash dumps.

### Phase 3: Gameplay systems
- Inventories: `PlayerInventory`, armor, offhand, cursor, crafting grid, ender chest, creative
  inventory + `CreativeInventoryCache`. Inventory transactions (`inventory/transaction/*`) and
  `ItemStackRequestExecutor`.
- Container blocks: remaining block inventories (furnace, hopper, brewing stand, anvil, barrel,
  shulker box, ender chest) and missing tiles (FlowerPot, Cauldron, furnace variants, TileFactory).
- Crafting (`crafting/*`: shaped/shapeless, furnace, brewing, smithing, loaded from pmmp's
  BedrockData recipe JSON).
- ~~Entity systems, entities, enchantments~~: done (see §5). Remaining: item NBT serialization
  (then turn `CanSaveWithChunk` back on for ItemEntity/Trident and save Human inventories),
  death/respawn packet flow, game-mode switching, wiring item use (bow, throwables, spawn eggs,
  buckets) to the ported entities, and the player events PHP fires from `Player`.
- Missing trees (acacia, jungle, azalea, nether) and `TreeFactory`. Async generation (goroutine
  pool instead of `AsyncGeneratorExecutor`).

### Phase 4: Plugins
- **Open design decision:** PHP plugins can't run in Go. Options are (a) compile-time Go plugins
  registered via an interface (simplest, like Dragonfly), (b) Go's `plugin` package (`.so`, Linux
  only, fragile), (c) an embedded scripting/WASM runtime. Decide before porting `PluginManager`
  and `PluginBase`. `PluginDescription`/`ApiVersion` parsing (`plugin.yml`) is already ported and
  usable in all three.

### Phase 5: Everything else
Resource packs, query protocol, UPnP, Xbox Live auth (gophertunnel supports it, just turn it on +
config), crash dumps, timings reports, block-state/item upgraders (load old/vanilla worlds),
region formats (Anvil/McRegion → conversion), forms.

## 7. Measuring progress against upstream

The status numbers above come from mapping upstream class names to Go file names. To redo it:

```bash
git clone --depth 1 https://github.com/pmmp/PocketMine-MP.git /tmp/pmmp
# For each src/<ns>/<Class>.php check whether pocketmine/<ns>/<snake_case(Class)>.go exists.
# Classes merged into one file (all sounds in world/sound/sound.go, all particles in
# world/particle/particle.go, all biomes in world/biome/vanilla_biomes.go) won't match by
# filename, so grep for `type <Class>` or `New<Class>` for those.
```

When you finish a chunk of work, update the checklist in `README.md` and §5 of this file.

## 8. Tips for agents

- Always have the PHP source open for whatever you're porting. Behaviour must match it.
- Check whether a helper is already ported before writing one (`grep -rn "is a port of" pocketmine`).
- Don't add features PocketMine-MP doesn't have.
- `cmd/pocketmine-go/main.go` is only the entry point (PocketMine.php). Put new logic in the proper
  `pocketmine/...` package.
- Run `go vet ./... && go test ./...` before committing.
