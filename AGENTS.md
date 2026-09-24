# AGENTS.md

Context for anyone (human or AI agent) picking up work on **pocketmine-go**. Read this before
changing code. The feature checklist lives in [README.md](README.md#feature-checklist). This file
covers the goal, how the code is organised, the conventions, what state things are in, and what to
do next.

_Last updated: 2026-09-24. Upstream reference: PocketMine-MP `5.44.4` (commit `6a7cc02`, July 2026)._

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
go run ./cmd/pocketmine-go          # starts a server on UDP :19132
go run ./cmd/pocketmine-go -port 19133 -seed 1234 -world-dir world -motd "test" -max-players 20
```

- Go version: see `go.mod` (`go 1.26.1`).
- Client protocol: whatever the pinned gophertunnel supports (v1.57.1 → Bedrock **1.26.30**,
  protocol 1001). Bumping gophertunnel means re-checking the vendored assets (see §3).
- Xbox Live auth is **disabled** (`AuthenticationDisabled: true` in `main.go`), so any client can join.
- World data is written to `-world-dir` (LevelDB + `level.dat`). Delete that directory to regenerate.

## 3. Repository layout

```
cmd/pocketmine-go/     The runnable server. main.go = startup, tick loop, packet read loop.
                       session.go = per-connection state + player registry (join/leave/move/
                       sound/particle broadcast). Currently acts as a stand-in for Server +
                       NetworkSession + InGamePacketHandler (none of which exist yet).
pocketmine/            One Go package per PHP namespace under pmmp/PocketMine-MP/src/.
  block/               ~260 of 270 block classes, tile/ (tiles), inventory/ (block inventories), utils/
  item/                ~110 of 136 item classes, VanillaItems (partial)
  world/               World, WorldManager, Explosion, tick loop, ChunkListener
    format/            Chunk, SubChunk, PalettedBlockArray, Height/LightArray
    format/io/         level.dat (WorldData), leveldb/ (chunk read/write)
    generator/         Flat, Normal, hell (Nether), noise, populators, trees/ores, biomeselector
    light/             sky + block light propagation
    biome/, particle/, sound/, utils/
  network/mcpe/convert/     BlockTranslator, BlockStateSerializer, ItemTranslator (internal ↔ network IDs)
  network/mcpe/serializer/  ChunkSerializer (LevelChunk payload)
  data/bedrock/        BlockStateDictionary, ItemTypeDictionary + vendored assets/
                       (canonical_block_states.nbt, required_item_list.json)
  data/runtime/        Runtime state bit-packing (describer/reader/writer)
  player/              Player, GameMode, chunk streaming, SurvivalBlockBreakHandler, combat,
                       fall damage, PlayerInfo, OfflinePlayer, DatFilePlayerDataProvider
  entity/              Entity, Living, Human, Skin, damage events
  command/, event/, permission/, plugin/, scheduler/, lang/, timings/, log/, promise/
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
**roughly 570–700 of 1,498 PHP classes (~40–47%)** have a Go counterpart (569 by strict filename match, 709 when also matching Go type names). Coverage is very uneven:

| Area | State |
|---|---|
| block (+ tiles, block inventories, utils) | ~87% of classes ported. **Only ~55 vanilla block singletons are registered and ~15 have network serializers**, so only a handful can actually appear in-game. |
| item | ~72% of classes. ~76 vanilla item singletons. Enchantments not started. |
| world core | World, ticking, scheduled/random updates, light, explosions, WorldManager, ChunkListener, level.dat, LevelDB save/load: done. |
| generators | Flat, Normal (all biomes), Nether done. Trees: only oak/spruce/birch. Generation is synchronous (no async executor). |
| player/entity | Player, Human, Living, chunk streaming, survival block breaking, PvP, fall damage. No attributes/hunger/XP/effects, no non-player entities. |
| framework libs | command (base only), event (base only, **no concrete events**), permission, scheduler (sync only), lang, timings, log, promise, plugin (description parsing only). |
| **server glue** | **Missing.** No `Server`, `ServerProperties`, `pocketmine.yml`, console, `NetworkSession`, packet handlers. `cmd/pocketmine-go/main.go` is a hand-written stand-in. |
| not started | crafting, inventory transactions, player/armor/creative inventories, entity effects/animations/projectiles/objects, enchantments, all concrete events, default commands, plugin loading, resource packs, query, crash dumps, region (Anvil/McRegion) world formats, block-state upgrader (old world compatibility). |

**Important: many ported packages are not used by the running server yet.** `command`, `event`,
`permission`, `scheduler`, `WorldManager` and `DatFilePlayerDataProvider` all exist and are tested,
but `main.go` doesn't call any of them. Wiring them up through a real `Server` type is the next
big structural step (Phase 2 below).

### What a player can do today

Connect (offline mode), spawn in a generated Normal world, walk around with chunks streaming in,
break blocks (with correct survival break times, bare hand only), see other players and their
movement, hit other players (damage + knockback), take fall damage. The world saves to LevelDB on
shutdown (Ctrl+C).

### What a player cannot do yet

Place blocks, use or move items, select a hotbar slot (everything acts as a bare hand), chat (text
is received but not broadcast), run commands, craft, open containers, eat, die/respawn properly,
change game mode.

### Known issues

- **Floating up / flying after spawn.** Reported by a tester (2026-09-24): right after spawning the
  player drifted up into the sky. Commit `697dfa7` fixed a related bug (client-requested flight was
  never denied, so the swim-up gesture enabled flying). If this still happens on current `main`, it
  is a separate bug. Not reproduced yet. Things to check: the spawn Y passed to `StartGame`
  (`computeSpawn` returns surface+1 as a block position) and whether the client is placed inside a
  block or in a chunk that hasn't arrived yet; that `UpdateAbilities` is received before movement
  starts; that there's no leftover `MayFly`/`Flying` state. There is no server-side physics or
  movement validation. The client's `PlayerAuthInput` position is trusted as-is.
- Blocks the generator can place but that have no network serializer can't be sent to the client.
  Keep `main.go`'s block list and `convert/vanilla_block_mappings.go` in sync until the full
  mappings are ported.
- Everything held is treated as a bare hand (`bareHandItem` in `main.go`).
- Worlds created by vanilla Bedrock or PocketMine-MP (PHP) will mostly fail to load: only the
  block states this port knows are recognised, and there is no block-state upgrader.

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

### Phase 2: Real server structure (replace the stand-in in `cmd/`)
1. Port `Server` (+ `ServerProperties`, `ServerConfigGroup`, `server.properties`/`pocketmine.yml`
   loading via the existing `utils.Config`).
2. Port `NetworkSession` + `handler/*` (`PreSpawnPacketHandler`, `InGamePacketHandler`,
   `DeathPacketHandler`, ...) on top of gophertunnel's `minecraft.Conn`, and `TypeConverter`.
3. Move `main.go`'s logic into those types. Use `WorldManager` instead of a single `world.New`.
   Drive the tick through `Server`. Use `scheduler.TaskScheduler`.
4. Console reader + `ConsoleCommandSender`, then the default commands (`command/defaults`: start
   with `stop`, `help`, `list`, `say`, `gamemode`, `tp`, `give`, `time`, `op`/`deop`, `kick`,
   `ban`, `whitelist`).
5. Port all concrete **events** (`event/block|entity|inventory|player|plugin|server|world`) and
   fire them from the places PHP fires them. Required before plugins make sense.
6. Player data persistence (wire `DatFilePlayerDataProvider`), permissions/ops wired to players.

### Phase 3: Gameplay systems
- Inventories: `PlayerInventory`, armor, offhand, cursor, crafting grid, ender chest, creative
  inventory + `CreativeInventoryCache`. Inventory transactions (`inventory/transaction/*`) and
  `ItemStackRequestExecutor`.
- Container blocks: remaining block inventories (furnace, hopper, brewing stand, anvil, barrel,
  shulker box, ender chest) and missing tiles (FlowerPot, Cauldron, furnace variants, TileFactory).
- Crafting (`crafting/*`: shaped/shapeless, furnace, brewing, smithing, loaded from pmmp's
  BedrockData recipe JSON).
- Entity systems: attributes (`AttributeMap`), hunger, XP, effects (`entity/effect`), animations,
  death/respawn, game-mode switching.
- Entities: `ItemEntity` (drops!), `FallingBlock`, `PrimedTNT`, `ExperienceOrb`, `Painting`,
  projectiles (arrow, snowball, egg, ender pearl, ...), and the few mobs PMMP has (Zombie,
  Villager, Squid).
- Enchantments (`item/enchantment`).
- Movement: server-side physics/collision checks (PMMP's `Entity::move`), anti-fly basics.
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
- `cmd/pocketmine-go/main.go` is temporary glue. Put new logic in the proper `pocketmine/...`
  package and keep `main.go` thin.
- Run `go vet ./... && go test ./...` before committing.
