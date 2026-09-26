# pocketmine-go

A port of [PocketMine-MP](https://github.com/pmmp/PocketMine-MP), the Minecraft: Bedrock Edition
server, from PHP to Go.

> **Status: early and incomplete. Not usable as a real server yet.**
> You can join, walk around a generated world, break blocks and hit other players. You can't
> place blocks, use items, chat or run commands yet. See the [checklist](#feature-checklist).

The goal is to keep PocketMine-MP's **game logic** (blocks, items, world, generation, lighting,
player rules) faithful to the original, one PHP class → one Go type. The **network protocol** is
the exception: it comes from [gophertunnel](https://github.com/sandertv/gophertunnel) and is not
ported by hand.

Port target: PocketMine-MP **5.44.4**. Client: Bedrock **1.26.50** (via gophertunnel v1.62.0). Upstream PocketMine-MP itself only supports 1.26.30 so far.

Contributing or using an AI agent? Read **[AGENTS.md](AGENTS.md)** first. It covers the plan,
architecture, porting conventions and known issues.

## Quick start

Requires Go (see `go.mod`).

```bash
go run ./cmd/pocketmine-go                 # data (server.properties, worlds/, players/) in the current directory
go run ./cmd/pocketmine-go --data=server   # or in another folder (PocketMine.php's --data)
```

Then add a server in Minecraft Bedrock pointing at your machine's IP, port `19132`.

Settings live in `server.properties`, created on first start with PocketMine-MP's defaults
(`server-port`, `motd`, `max-players`, `gamemode`, `difficulty`, `level-name`, `level-seed`,
`view-distance`, `xbox-auth`, ...). Any of them can be overridden on the command line with
`--key=value`, e.g. `--server-port=19133` or `--xbox-auth=false`. Like PocketMine-MP,
**`xbox-auth` is on by default**: players must be signed in to Xbox Live.

Stop with Ctrl+C or by typing `stop` in the console (players and worlds are saved). Run tests with `go test ./...`.

## Progress

Roughly **1,230–1,270 of PocketMine-MP's 1,498 PHP classes (~82–85%)** have a Go counterpart: 1,235
match by file or type name, and about 40 more were merged or renamed in Go (traits as `...Component`
structs, all biomes in one file, the tree types as `Tree` constructors). Many of the rest are out
of scope on purpose (PHP threads, RakLib and the protocol classes gophertunnel replaces, the
updater). The server "glue" is in place: `Server`, network sessions and packet handlers, events,
the command map with most default commands, the console, query and UPnP, crafting and enchanting.
World and blocks are complete, incl. world formats and tiles. The big remaining gaps are plugins,
crash dumps and 7 default commands.

| Area (PHP namespace, incl. sub-namespaces) | Ported / total PHP classes |
|---|---|
| `block` (incl. tile, inventory, utils) | 390 / 390 (every block, all tiles via `TileFactory`, all 19 block inventories; the 25 traits are embedded `...Component` structs) |
| `item` (incl. enchantment) | 143 / 154 (incl. item NBT, enchanting table helper and registries) |
| `world` (all sub-namespaces) | ~262 / 271 (missing: `GeneratorManager`/`GeneratorManagerEntry`/`InvalidGeneratorOptionsException`, `FlatGeneratorOptions`, `PopulationUtils`, `ChunkTicker`, `WorldTimings`, the two biome definition models) |
| ↳ `world/particle` | 38 / 38 |
| ↳ `world/sound` | 113 / 113 |
| ↳ `world/format` + `world/format/io` (LevelDB, region, conversion) | 76 / 76 (PHP exceptions are Go error types) |
| ↳ `world/generator` (incl. object, populator, noise) | ~34 / 39 (every generator, populator and tree) |
| `data` (block/item (de)serializers, upgraders, runtime) | 55 / 99 by name (the rest are mostly per-block `Model`/helper classes merged into bigger Go files) |
| `player` | 16 / 16 |
| `permission` | 9 / 14 |
| `command` | 43 / 53 (34 of 41 default commands) |
| `plugin` | 8 / 22 |
| `scheduler` | 14 / 15 (all but `DumpWorkerMemoryTask`, which needs PHP's MemoryDump) |
| `entity` (incl. effect, object, projectile, animation, attribute) | 77 / 77 (ItemEntity, Trident and Human inventories are saved) |
| `event` | ~145 / 150 (every concrete event; fired where the ported code fires them) |
| `inventory` | 34 / 35 (incl. crafting and enchanting transactions; `json/CreativeGroupData` isn't needed) |
| `network` (above the protocol layer) | ~40 / 85 (most of the rest are protocol-level classes gophertunnel replaces: compression, encryption, JWT/login, RakLib) |
| `console` | 2 / 5 (the rest is PHP child-process plumbing) |
| `resourcepacks`, `form` | 4 / 11 (the manifest classes are gophertunnel's) |
| `crafting` | 25 / 25 (recipes loaded from pmmp/BedrockData's recipe JSON) |
| `crash` | 0 |

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
- [ ] Crash dumps
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
- [ ] Redstone behaviour beyond what individual blocks implement

### Items
- [x] ~108 item classes (tools, armor, food, potions, buckets, books, records, ...)
- [x] 321 item type IDs
- [x] Item → network ID translation
- [ ] Vanilla item registry **(partial)**, ~76 items
- [ ] Bow, arrows, snowball, egg, ender pearl, spawn eggs, and other projectile items **(partial)**: item use is wired through the packet handlers; buckets, flint and steel and spawn eggs' block interactions aren't ported yet
- [x] Enchantments (all vanilla enchantments, protection/sharpness/knockback/fire aspect logic, armor EPF)
- [ ] `/give`-style item name parsing (`StringToItemParser`)

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
- [x] Creative inventory (1,215 entries, 727 of them block items)
- [x] Inventory transactions / item stack requests (moving, dropping, using items, crafting, enchanting)
- [x] Crafting (shaped/shapeless recipes, `CraftingDataPacket`, `CraftingTransaction`) (recipes with potions/unknown items are skipped, like PHP)
- [x] Enchanting table (options, bookshelves, `EnchantingTransaction`, lapis/XP cost)
- [x] Block inventories (all 19) and opening crafting table, enchanting table, anvil, loom, stonecutter, smithing/cartography table and ender chest windows
- [x] Container tiles holding inventories (chest, barrel, furnace, hopper, brewing stand, shulker box), saved with the world
- [x] Furnace smelting and brewing
- [ ] Smithing

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
- [ ] Default commands **(partial)**: 34 of 41 (`/give`, `/clear`, `/enchant`, `/effect`, `/particle`, `/timings`, `/dumpmemory` are missing)
- [x] Event system base (handlers, priorities, cancellable, parent events)
- [x] Concrete events (block, entity, player, inventory, world, server, plugin)
- [x] Permissions, attachments, ban lists, ops
- [x] `plugin.yml` parsing, API version checks
- [ ] Plugin loading and `PluginManager`. **Design undecided**: PHP plugins can't run in Go. See AGENTS.md §6 Phase 4.

## Roadmap

1. ~~**Make the world playable:** all block and item network mappings~~: done.
2. **Real server structure:** done (`Server`, network sessions and handlers, console, command map, events, permissions, query, UPnP, resource packs). Remaining: the 7 missing default commands.
3. ~~**Gameplay:** item NBT, container tiles, furnace/brewing ticks~~: done. Remaining: smithing, buckets/flint and steel/spawn eggs on blocks.
4. **Plugins** (design decision pending, see AGENTS.md §6 Phase 4).
5. **Everything else:** crash dumps.

Details in [AGENTS.md](AGENTS.md#6-plan--roadmap).

## Known issues

- Some testers report floating up into the sky right after spawning. Under investigation. See
  [AGENTS.md → Known issues](AGENTS.md#known-issues).
- Blocks that PocketMine-MP 5.44.4 itself doesn't implement (moss, kelp, seagrass, dripstone, ...)
  load as the "update!" block when a vanilla world is opened, exactly like in PHP.

## Credits

- [PocketMine-MP](https://github.com/pmmp/PocketMine-MP): the original project this is ported from (LGPL-3.0).
- [gophertunnel](https://github.com/sandertv/gophertunnel): Bedrock protocol implementation.
