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

Roughly **750–850 of PocketMine-MP's 1,498 PHP classes (~50–57%)** have a Go counterpart. The range depends on how classes that were merged or renamed in Go are counted. The server
"glue" is in place: `Server`, network sessions and packet handlers, events, the command map with
most default commands, the console, query and UPnP, crafting and enchanting. The big remaining gaps
are plugins, item NBT and container tiles.

| Area (PHP namespace, incl. sub-namespaces) | Ported / total PHP classes |
|---|---|
| `block` (incl. tile, inventory, utils) | 348 / 390 (all 19 block inventories) |
| `item` (incl. enchantment) | 137 / 154 (incl. enchanting table helper and registries) |
| `world` (all sub-namespaces) | 135 / 271 |
| ↳ `world/particle` | 38 / 38 |
| ↳ `world/sound` | 48 / 113 |
| ↳ `world/format/io` (LevelDB, region, upgraders) | LevelDB + level.dat only |
| `player` | 16 / 16 |
| `permission` | 9 / 14 |
| `command` | 43 / 53 (34 of 41 default commands) |
| `plugin` | 8 / 22 |
| `scheduler` | 14 / 15 (all but `DumpWorkerMemoryTask`, which needs PHP's MemoryDump) |
| `entity` (incl. effect, object, projectile, animation, attribute) | 77 / 77 |
| `event` | ~145 / 150 (every concrete event; fired where the ported code fires them) |
| `inventory` | 34 / 35 (incl. crafting and enchanting transactions; `json/CreativeGroupData` isn't needed: creative items come from the vendored 1.26.50 data) |
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
- [x] LevelDB world save/load, `level.dat`
- [x] Sky and block lighting
- [x] World tick: time, weather, scheduled and neighbour updates, random ticks
- [x] Chunk loading/unloading, chunk loaders, chunk listeners
- [x] Explosions
- [x] Multi-world manager (`WorldManager`, `worlds:` in pocketmine.yml)
- [x] Particles (all types) and sounds **(partial)**, 49 of 113 sound types
- [ ] Block-state / item upgraders (loading worlds from vanilla or older PMMP)
- [ ] Region formats (Anvil, McRegion, PMAnvil) and world conversion
- [x] Async chunk generation / population / lighting (worker goroutines, like PHP's `AsyncGeneratorExecutor`, `PopulationTask` and `LightPopulationTask`)

### World generation
- [x] Flat generator
- [x] Normal generator (noise terrain, all PMMP biomes)
- [x] Nether generator
- [x] Ores, tall grass, ground cover populators
- [x] Trees **(partial)**: oak, spruce, birch. Acacia, jungle, azalea and nether trees missing.

### Blocks
- [x] ~260 block classes ported with their behaviour (state, placement rules, drops, random ticks, ...)
- [x] 800 block type IDs
- [x] Most tiles (chest, furnace, hopper, sign, banner, bed, ...)
- [x] Survival block breaking with correct break times
- [ ] Block placing
- [ ] Vanilla block registry **(partial)**, ~55 blocks registered
- [x] Block ↔ network mappings: full `data/bedrock/block/convert` (`BlockObjectToStateSerializer`, `BlockStateToObjectDeserializer`, reader/writer, `VanillaBlockMappings`); all 799 `VanillaBlocks` (11,125 states) round-trip. The 1.26.50-only properties (`minecraft:corner` on stairs, `minecraft:connection_*` on fences/panes/bars/tripwire) are written with neutral values
- [x] Item ↔ network mappings: `data/bedrock/item` (`ItemSerializer`/`ItemDeserializer`, `ItemSerializerDeserializerRegistrar`, `BlockItemIdMap`) and the full `VanillaItems`
- [ ] Cauldrons, flower pot
- [ ] Block inventories for furnace, hopper, brewing stand, anvil, barrel, shulker box, ender chest
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
- [ ] Container tiles holding inventories (chest, barrel, furnace, hopper, brewing stand, shulker box): needs item NBT serialization to save their contents
- [ ] Furnace smelting and brewing ticks (the recipes are loaded), smithing

### Entities
All 77 classes under `pocketmine\entity` are ported, with their full logic.
- [x] Entity, Living, Human (movement/collision physics, fire, air supply, knockback, armor, death)
- [x] EntityFactory + entity NBT save/load (LevelDB `actorprefix` storage)
- [x] Item drops (`ItemEntity`) **(partial)**: item NBT serialization isn't ported, so dropped items aren't saved with the chunk
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
3. **Gameplay:** item NBT, container tiles, furnace/brewing ticks.
4. **Plugins** (design decision pending, see AGENTS.md §6 Phase 4).
5. **Everything else:** world upgraders, crash dumps.

Details in [AGENTS.md](AGENTS.md#6-plan--roadmap).

## Known issues

- Some testers report floating up into the sky right after spawning. Under investigation. See
  [AGENTS.md → Known issues](AGENTS.md#known-issues).
- Worlds from vanilla Bedrock or PHP PocketMine-MP won't load correctly yet.

## Credits

- [PocketMine-MP](https://github.com/pmmp/PocketMine-MP): the original project this is ported from (LGPL-3.0).
- [gophertunnel](https://github.com/sandertv/gophertunnel): Bedrock protocol implementation.
