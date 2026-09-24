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

Port target: PocketMine-MP **5.44.4**. Client: Bedrock **1.26.30** (via gophertunnel v1.57.1).

Contributing or using an AI agent? Read **[AGENTS.md](AGENTS.md)** first. It covers the plan,
architecture, porting conventions and known issues.

## Quick start

Requires Go (see `go.mod`).

```bash
go run ./cmd/pocketmine-go
```

Then add a server in Minecraft Bedrock pointing at your machine's IP, port `19132`.
Xbox Live authentication is off, so any account can join.

| Flag | Default | Meaning |
|---|---|---|
| `-port` | `19132` | UDP port |
| `-seed` | `0` | World seed (only used when the world is first created) |
| `-world-dir` | `world` | Where the LevelDB world is saved |
| `-motd` | `PocketMine-MP` | Server list name |
| `-max-players` | `20` | Advertised player limit |

Stop with Ctrl+C (the world is saved on shutdown). Run tests with `go test ./...`.

## Progress

Roughly **570–700 of PocketMine-MP's 1,498 PHP classes (~40–47%)** have a Go counterpart. The range depends on how classes that were merged or renamed in Go are counted. Most of that is
blocks, items and world code. Much of the server "glue" (the `Server` class, network sessions,
events, commands) isn't ported yet, and several ported packages aren't hooked into the running
server.

| Area (PHP namespace, incl. sub-namespaces) | Ported / total PHP classes |
|---|---|
| `block` (incl. tile, inventory, utils) | 341 / 390 |
| `item` (incl. enchantment) | 133 / 154 |
| `world` (all sub-namespaces) | 135 / 271 |
| ↳ `world/particle` | 38 / 38 |
| ↳ `world/sound` | 48 / 113 |
| ↳ `world/format/io` (LevelDB, region, upgraders) | LevelDB + level.dat only |
| `player` | 14 / 16 |
| `permission` | 9 / 14 |
| `command` | 9 / 53 (no default commands) |
| `plugin` | 8 / 22 |
| `scheduler` | 4 / 15 |
| `entity` (incl. effect, object, projectile, animation, attribute) | 77 / 77 |
| `event` | 43 / 150 (base classes + every `event/entity` event + 2 player events) |
| `inventory` | 10 / 35 (incl. player, armor, off-hand, ender inventories) |
| `network` (above the protocol layer) | 6 / 85 |
| `crafting`, `console`, `crash`, `resourcepacks`, `form` | 0 |

_Counts are approximate: a class counts as ported if a Go file with its snake_case name or a Go
type with its name exists._

Legend for the checklist below: `[x]` done · `[ ]` not done. **(partial)** means started but incomplete.

## Feature checklist

### Networking & connection
- [x] RakNet listener, login, encryption, resource-pack handshake (via gophertunnel)
- [x] Server list entry (MOTD, player count)
- [x] StartGame, item table, abilities, spawn
- [x] Chunk streaming by view distance as the player moves
- [x] Multiple players see each other (player list, spawn, movement)
- [ ] Xbox Live authentication (currently disabled)
- [ ] `NetworkSession` and packet handlers ported from PMMP (currently a stand-in in `cmd/`)
- [ ] Packet rate limiting, broadcast batching, chunk cache
- [ ] Query protocol, UPnP
- [ ] Resource packs
- [ ] Transfer server, forms

### Server core
- [ ] `Server` class
- [ ] `server.properties` / `pocketmine.yml`
- [ ] Console input and console command sender
- [x] Logger, text formatting, language/translation files (`lang`)
- [x] Config files (YAML/JSON/properties) via `utils.Config`
- [x] Sync task scheduler (not wired to the server yet)
- [ ] Async tasks / worker pool
- [x] Timings (not wired to the server yet)
- [ ] Crash dumps
- [x] Version info

### World
- [x] Chunks, sub-chunks, paletted block storage, heightmaps
- [x] LevelDB world save/load, `level.dat`
- [x] Sky and block lighting
- [x] World tick: time, weather, scheduled and neighbour updates, random ticks
- [x] Chunk loading/unloading, chunk loaders, chunk listeners
- [x] Explosions
- [x] Multi-world manager (`WorldManager`) (not wired to the server yet)
- [x] Particles (all types) and sounds **(partial)**, 49 of 113 sound types
- [ ] Block-state / item upgraders (loading worlds from vanilla or older PMMP)
- [ ] Region formats (Anvil, McRegion, PMAnvil) and world conversion
- [ ] Async chunk generation / population / lighting

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
- [ ] Block → network mappings **(partial)**, ~15 blocks can be sent to the client
- [ ] Cauldrons, flower pot
- [ ] Block inventories for furnace, hopper, brewing stand, anvil, barrel, shulker box, ender chest
- [ ] Redstone behaviour beyond what individual blocks implement

### Items
- [x] ~108 item classes (tools, armor, food, potions, buckets, books, records, ...)
- [x] 321 item type IDs
- [x] Item → network ID translation
- [ ] Vanilla item registry **(partial)**, ~76 items
- [ ] Bow, arrows, snowball, egg, ender pearl, spawn eggs, and other projectile items **(partial)**: the entities exist, but using the items needs the item-use packet handlers
- [x] Enchantments (all vanilla enchantments, protection/sharpness/knockback/fire aspect logic, armor EPF)
- [ ] `/give`-style item name parsing (`StringToItemParser`)

### Player
- [x] Player entity, game modes, skins, player info
- [x] Health, damage, knockback, PvP
- [x] Fall damage
- [x] Chat formatters
- [x] Player data file format (not wired to the server yet)
- [ ] Held item / hotbar selection (everything acts as a bare hand)
- [ ] Chat broadcast
- [ ] Death and respawn **(partial)**: death logic, drops and XP drop ported; respawn screen/packet flow isn't
- [x] Hunger, saturation, experience, attributes (logic ported; attribute packets sent)
- [ ] Changing game mode in-game
- [ ] Server-side movement checks / physics

### Inventory & crafting
- [x] Base inventory types
- [x] Initial inventory contents sent to the client
- [x] Player inventory, armor, offhand, ender chest (the classes; client sync is via InventoryContent only)
- [ ] Cursor inventory, inventory network sync (InventoryManager)
- [ ] Creative inventory
- [ ] Inventory transactions / item stack requests (moving, dropping, using items)
- [ ] Crafting, furnace smelting, brewing, smithing, enchanting

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
- [x] Command base classes and command map (not wired to the server yet)
- [ ] Default commands (`/stop`, `/help`, `/gamemode`, `/tp`, `/give`, `/time`, `/op`, `/ban`, ... 42 total)
- [x] Event system base (handlers, priorities, cancellable)
- [ ] Concrete events (block, entity, player, inventory, world, server, plugin: ~145) **(partial)**: all `event/entity` events are ported and fired
- [x] Permissions, attachments, ban lists (not wired to players yet)
- [x] `plugin.yml` parsing, API version checks
- [ ] Plugin loading and `PluginManager`. **Design undecided**: PHP plugins can't run in Go. See AGENTS.md §6 Phase 4.

## Roadmap

1. **Make the world playable:** all block mappings, block placing, held items, chat.
2. **Real server structure:** `Server`, `NetworkSession`, packet handlers, console, commands, events.
3. **Gameplay:** inventory transactions, crafting, item NBT, and wiring the ported entities into item use.
4. **Plugins.**
5. **Everything else:** resource packs, query, auth, world upgraders, crash dumps.

Details in [AGENTS.md](AGENTS.md#6-plan--roadmap).

## Known issues

- Some testers report floating up into the sky right after spawning. Under investigation. See
  [AGENTS.md → Known issues](AGENTS.md#known-issues).
- Worlds from vanilla Bedrock or PHP PocketMine-MP won't load correctly yet.

## Credits

- [PocketMine-MP](https://github.com/pmmp/PocketMine-MP): the original project this is ported from (LGPL-3.0).
- [gophertunnel](https://github.com/sandertv/gophertunnel): Bedrock protocol implementation.
