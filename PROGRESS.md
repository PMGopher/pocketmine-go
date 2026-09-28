# Porting progress

pocketmine-go reimplements PocketMine-MP **5.44.4** in Go, one upstream class at a time, with the
network protocol provided by [gophertunnel](https://github.com/sandertv/gophertunnel). This page
tracks how much of upstream has a Go counterpart. For the architecture and the porting
conventions, read [AGENTS.md](AGENTS.md); for what the server is and how to use it, the
[README](README.md).

## Progress

Roughly **1,260–1,300 of PocketMine-MP's 1,498 PHP classes (~84–87%)** have a Go counterpart: 1,260
match by file or type name, and about 40 more were merged or renamed in Go (traits as `...Component`
structs, all biomes in one file, the tree types as `Tree` constructors). Many of the rest are out
of scope on purpose (PHP threads, RakLib and the protocol classes gophertunnel replaces, the
updater). The server "glue" is in place: `Server`, network sessions and packet handlers, events,
the command map with most default commands, the console, query and UPnP, crafting and enchanting.
World and blocks are complete, incl. world formats and tiles, and plugins load as compiled-in Go
packages. Every area is complete: what isn't ported is PHP-runtime specific (threads, Phar plugins,
PHP heap dumps) or replaced by gophertunnel (the protocol layer).

| Area (PHP namespace, incl. sub-namespaces) | Status | Ported / total PHP classes | Notes |
|---|---|---|---|
| `block` (incl. tile, inventory, utils) | ✅ Complete | 390 / 390 | Every block, all tiles (`TileFactory`), all 19 block inventories; the traits are `...Component` structs |
| `world` (all sub-namespaces) | ✅ Complete | ~265 / 271 | LevelDB, region formats + conversion, upgraders, generators, light, all 38 particles and 113 sounds. The rest are merged (`GeneratorManager` is `generator/manager.go`, biomes in one file) or PHP-only (`WorldTimings`) |
| `item` (incl. enchantment) | ✅ Complete | 145 / 154 | All items, enchantments, item NBT, string-to-item parsers. The rest are data tables (`VanillaItemsInputs`, `ItemTypeIds`...) merged into the registries, or traits |
| `entity` (incl. effect, object, projectile) | ✅ Complete | 75 / 77 | The 2 left are `InvalidSkinException` (a Go error) and `VanillaEffectsInputs` (merged into the effect registry) |
| `player` | ✅ Complete | 14 / 16 | The 2 left are exceptions (Go errors) |
| `crafting` | ✅ Complete | 25 / 25 | Recipes from pmmp/BedrockData's recipe JSON |
| `inventory` | ✅ Complete | 32 / 35 | The rest are 2 exceptions (Go errors) and `CreativeGroupData` (the creative list comes from the vanilla server's packet) |
| `command` | ✅ Complete | 50 / 53 | 40 of 41 default commands. `dumpmemory` needs PHP's `MemoryDump` (PHP heap dump); `CommandSender`/`CommandExecutor` are Go interfaces |
| `event` | ✅ Complete | 145 / 150 | Every concrete event; the rest are internal caches/tags merged into the event manager |
| `plugin` | ✅ Complete | 18 / 22 | Plugins are Go packages compiled into the server. `PharPluginLoader`/`ScriptPluginLoader` run PHP code, which Go can't |
| `network` (above the protocol layer) | ✅ Complete | ~40 / 85 | The rest is the protocol layer gophertunnel replaces: RakLib, compression, encryption, JWT/login, pthreads channels |
| `data` (serializers, upgraders, ID maps) | ✅ Complete | 56 / 99 | The rest are ID tables, state maps and exceptions merged into bigger Go files |
| `permission` | ✅ Complete | 9 / 14 | The rest are traits/internals merged into `Permissible` |
| `scheduler` | ✅ Complete | 12 / 15 | `DumpWorkerMemoryTask` needs PHP's `MemoryDump`; the rest are exceptions/internals |
| `console` | ✅ Complete | 2 / 5 | The rest is PHP's child-process console reader (Go reads stdin directly) |
| `resourcepacks`, `form` | ✅ Complete | 4 / 11 | The manifest classes are gophertunnel's |
| `utils`, `promise` | ✅ Complete | 25 / 36 | PHP traits (`SingletonTrait`, `EnumTrait`...) and thread classes; Go has generics and goroutines |
| `crash` | ✅ Complete | 5 / 5 | Crash dumps (`crashdumps/`) |
| `lang`, `timings`, `wizard` | ✅ Complete | 7 / 7 | |
| `thread`, `updater`, `stats` | ⛔ Out of scope | 0 / 17 | PHP-runtime specific (pthreads, PHP updater, PHP stats) |

_Counts are approximate: a class counts as ported if a Go file with its snake_case name or a Go
type with its name exists._

Legend for the checklist below: `[x]` done · `[ ]` not done. **(partial)** means started but incomplete.

## Feature checklist

### Networking & connection
- [x] RakNet listener, login, encryption, resource-pack handshake (via gophertunnel)
- [x] Server list entry (MOTD, player count) (`RakLibInterface::setName`)
- [x] StartGame, item table, abilities, spawn
- [x] Chunk streaming by view distance as the player moves (through `ChunkCache`)
- [x] Multiple players see each other (player list, spawn, movement)
- [x] Xbox Live authentication (`xbox-auth`, on by default)
- [x] `RakLibInterface` on gophertunnel: pre-login checks (`PlayerPreLoginEvent`: server full, whitelist, name/IP bans), duplicate-login and XUID checks, IP blocking, raw packet filters, bandwidth stats, ping
- [x] `NetworkSession` (full port above the wire), `NetworkSessionManager`, `Network`
- [x] Packet handlers: `PreSpawnPacketHandler`, `InGamePacketHandler` (movement, block breaking/placing, item use, inventory transactions, item stack requests, containers, signs, books, lecterns, forms, skins, commands, emotes), `DeathPacketHandler`
- [x] `InventoryManager` (window IDs, item stack IDs, predictions, container open/close), `ItemStackRequestExecutor`, creative inventory cache
- [x] Block changes sent to players (`World::changedBlocks`/`sendBlocks`)
- [x] Packet rate limiting, broadcasting (`StandardPacketBroadcaster`, `StandardEntityEventBroadcaster`), chunk cache
- [x] Query protocol (on the game port, or a dedicated interface), UPnP port forwarding
- [x] Resource packs (`resource_packs.yml`, `PlayerResourcePackOfferEvent`; delivery by gophertunnel)
- [x] Transfer server, forms, toasts, titles
- [x] `DataPacketSend/Receive/DecodeEvent`

### Server core
- [x] `Server` class: startup, `pocketmine.yml` + `server.properties`, language, ops/whitelist/ban lists, broadcast channels, player data, tick loop with TPS/load tracking, console title, query info regeneration, shutdown
- [x] `server.properties` + `pocketmine.yml` (`ServerConfigGroup`, `--key=value` overrides)
- [x] Console input and console command sender (`ConsoleReader`, `ConsoleCommandSender`, `BroadcastLoggerForwarder`)
- [x] Logger (`MainLogger` with `server.log` and log archive), text formatting, language/translation files (`lang`)
- [x] Config files (YAML/JSON/properties/enum) via `utils.Config`
- [x] Sync task scheduler
- [x] Async tasks / worker pool (goroutines)
- [x] Timings, memory manager
- [x] Crash dumps
- [x] Version info, `server.lock` (one server per data folder)

### World
- [x] Chunks, sub-chunks, paletted block storage, heightmaps
- [x] LevelDB world save/load (vanilla layout: `db/` folder, zlib-raw compression), `level.dat`: every chunk version, sub-chunk versions 0-9, legacy terrain, 2D/3D biomes, tiles and entities
- [x] Loading worlds made by vanilla Bedrock (tested with BDS 1.26.52) or PHP PocketMine-MP
- [x] Sky and block lighting
- [x] World tick: time, weather, scheduled and neighbour updates, random ticks
- [x] Chunk loading/unloading, chunk loaders, chunk listeners
- [x] Explosions
- [x] Multi-world manager (`WorldManager`, `worlds:` in pocketmine.yml)
- [x] Particles (all types) and sounds (all 111 sound types)
- [x] Block-state and item upgraders (`BlockDataUpgrader`, `ItemDataUpgrader` with pmmp's upgrade schemas): old block states and items are upgraded on load
- [x] Region formats (Anvil, McRegion, PMAnvil) and automatic conversion to LevelDB (`FormatConverter`, backup in `backups/worlds`)
- [x] Item NBT (de)serialization: containers, dropped items, tridents and player inventories are saved
- [x] Liquids flow (water, lava, `MinimumCostFlowCalculator`, obsidian/cobblestone/basalt forming), fire spreads and burns blocks
- [x] Async chunk generation / population / lighting (worker goroutines, like PHP's `AsyncGeneratorExecutor`, `PopulationTask` and `LightPopulationTask`)

### World generation
- [x] Flat generator
- [x] Normal generator (noise terrain, all PMMP biomes)
- [x] Nether generator
- [x] Ores, tall grass, ground cover populators
- [x] Trees: oak, spruce, birch, jungle, acacia, azalea, nether fungi (`TreeFactory`); saplings and bone meal grow them

### Blocks
- [x] All block classes ported with their behaviour (state, placement rules, drops with Fortune/Silk Touch, random ticks, bone meal, hoes/shovels/axes, ...)
- [x] 800 block type IDs
- [x] All tiles (`TileFactory`), saved and loaded with their chunk and sent to clients (in sub-chunks and with block updates)
- [x] Survival block breaking with correct break times
- [x] Block placing
- [x] Vanilla block registry (all 799 `VanillaBlocks`)
- [x] Block ↔ network mappings: full `data/bedrock/block/convert` (`BlockObjectToStateSerializer`, `BlockStateToObjectDeserializer`, reader/writer, `VanillaBlockMappings`); all 799 `VanillaBlocks` (11,125 states) round-trip. The 1.26.50-only properties (`minecraft:corner` on stairs, `minecraft:connection_*` on fences/panes/bars/tripwire) are written with neutral values
- [x] Item ↔ network mappings: `data/bedrock/item` (`ItemSerializer`/`ItemDeserializer`, `ItemSerializerDeserializerRegistrar`, `BlockItemIdMap`) and the full `VanillaItems`
- [x] Cauldrons (water, lava, potions, dyed water), flower pot
- [x] Container blocks keep their items (chest, double chest, barrel, furnace, hopper, brewing stand, shulker box, campfire, chiseled bookshelf), furnace smelting and brewing
- [x] Beds (sleeping), respawn anchor, dragon egg, jukebox, lectern, item frames, cake with candles, banners with patterns, bells
- [x] Redstone: what individual blocks implement, as in PocketMine-MP (which has no redstone circuits)

### Items
- [x] ~108 item classes (tools, armor, food, potions, buckets, books, records, ...)
- [x] 321 item type IDs
- [x] Item → network ID translation
- [x] Vanilla item registry (full `VanillaItems`)
- [x] Bow, arrows, snowball, egg, ender pearl, spawn eggs and other projectile items; buckets, flint and steel, spawn eggs, paintings and end crystals used on blocks
- [x] Enchantments (all vanilla enchantments, protection/sharpness/knockback/fire aspect logic, armor EPF)
- [x] `/give`-style item name parsing (`StringToItemParser`, `LegacyStringToItemParser`)

### Player
- [x] Player entity, game modes, skins, player info
- [x] Health, damage, knockback, PvP
- [x] Fall damage
- [x] Chat formatters
- [x] Player data file format
- [x] Held item / hotbar selection
- [x] Chat broadcast and commands
- [x] Player data saved and restored (`players/<name>.dat`: position, health, hunger, XP, game mode)
- [x] Death and respawn (death screen, `DeathPacketHandler`, respawn)
- [x] Hunger, saturation, experience, attributes (logic ported; attribute packets sent)
- [x] Changing game mode in-game (`/gamemode`)
- [x] Server-side movement checks (`Player::handleMovement`: moves over 15 blocks per tick are reverted). PocketMine-MP has no server-side movement physics or anti-cheat for players ("This is NOT an anti-cheat check"), so there is nothing more to port

### Inventory & crafting
- [x] Base inventory types
- [x] Initial inventory contents sent to the client
- [x] Player inventory, armor, offhand, ender chest
- [x] Cursor inventory, inventory network sync (InventoryManager)
- [x] Creative inventory (1,428 entries from the vanilla server's list, in its groups; the rest are items PocketMine-MP 5.44.4 doesn't have)
- [x] Inventory transactions / item stack requests (moving, dropping, using items, crafting, enchanting)
- [x] Crafting (shaped/shapeless recipes, `CraftingDataPacket`, `CraftingTransaction`) (recipes with potions/unknown items are skipped, like PHP)
- [x] Enchanting table (options, bookshelves, `EnchantingTransaction`, lapis/XP cost)
- [x] Block inventories (all 19) and opening crafting table, enchanting table, anvil, loom, stonecutter, smithing/cartography table and ender chest windows
- [x] Container tiles holding inventories (chest, barrel, furnace, hopper, brewing stand, shulker box), saved with the world
- [x] Furnace smelting and brewing
- [x] Smithing table window, as in PocketMine-MP (which has no smithing recipes)

### Entities
All 77 classes under `pocketmine\entity` are ported, with their full logic.
- [x] Entity, Living, Human (movement/collision physics, fire, air supply, knockback, armor, death)
- [x] EntityFactory + entity NBT save/load (LevelDB `actorprefix` storage)
- [x] Item drops (`ItemEntity`), saved with the chunk
- [x] Falling blocks, primed TNT, experience orbs, paintings, end crystals, firework rockets, area effect clouds
- [x] Projectiles (arrow, snowball, egg, ender pearl, XP bottle, ice bomb, splash potion, trident)
- [x] Effects (all 27 vanilla effects, EffectManager)
- [x] Animations
- [x] Mobs (zombie, villager, squid). PocketMine-MP has no AI, so neither does this port

### Commands, events, permissions, plugins
- [x] Command base classes and command map
- [x] Default commands: 40 of 41 (`/dumpmemory` dumps PHP's heap, which Go doesn't have)
- [x] Event system base (handlers, priorities, cancellable, parent events)
- [x] Concrete events (block, entity, player, inventory, world, server, plugin)
- [x] Permissions, attachments, ban lists, ops
- [x] `plugin.yml` parsing, API version checks
- [x] Plugin loading and `PluginManager`: plugins are Go packages compiled into the server (imported in `cmd/pocketmine-go/plugins.go`, registered with `plugin.RegisterGoPlugin`) and loaded like PocketMine-MP loads its plugins (`plugin.yml`, API version, dependencies and load order, commands, permissions, data folder). PHP plugins themselves can't run in Go.

## Roadmap

1. ~~**Make the world playable:** all block and item network mappings~~: done.
2. ~~**Real server structure:**~~ done (`Server`, network sessions and handlers, console, command map, events, permissions, query, UPnP, resource packs). All default commands but `/dumpmemory` (PHP heap dump).
3. ~~**Gameplay:** item NBT, container tiles, furnace/brewing ticks, item interactions~~: done.
4. ~~**Plugins**~~: done, as compiled-in Go plugins.
5. ~~**Everything else:** crash dumps~~: done.

**The port is complete.** What isn't ported is PHP-runtime specific (threads, Phar/script plugins,
PHP heap dumps, the PHP updater) or the protocol layer gophertunnel replaces.

Details in [AGENTS.md](AGENTS.md#6-plan--roadmap).

## Known issues

- Blocks that PocketMine-MP 5.44.4 itself doesn't implement (moss, kelp, seagrass, dripstone, ...)
  are kept as they are when a vanilla world is opened (shown and saved unchanged, like Dragonfly),
  but have no behaviour: they break instantly and drop nothing. PHP turns them into "update!".
- Beacons have no window or effects and note blocks don't play when clicked: PocketMine-MP 5.44.4
  has no logic for either (only the beacon's light and the note block's stored pitch).
