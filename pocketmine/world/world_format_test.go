package world

import (
	"bytes"
	"compress/zlib"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"os"
	"path/filepath"
	blockinventory "pocketmine-go/pocketmine/block/inventory"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/math"
	"testing"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/block/tile"
	"pocketmine-go/pocketmine/nbt"
	"pocketmine-go/pocketmine/world/format/io"
	"pocketmine-go/pocketmine/world/format/io/leveldb"
	"pocketmine-go/pocketmine/world/format/io/region"
)

// writeAnvilWorld writes a Java Edition (pre-1.13) Anvil world with one chunk at 0,0: stone at
// y=0..15 (except a birch planks block at 1,1,1) and a chest tile.
func writeAnvilWorld(t *testing.T, path string) {
	t.Helper()
	if err := region.GenerateRegionWorld(path, "anvilworld", io.WorldCreationOptions{GeneratorName: "normal", Seed: 7}, 19133); err != nil {
		t.Fatal(err)
	}
	blocks := make([]byte, 4096)
	data := make([]byte, 2048)
	for i := range blocks {
		blocks[i] = 1 // stone
	}
	idx := 1<<8 | 1<<4 | 1 // YZX order: (y << 8) | (z << 4) | x
	blocks[idx] = 5        // planks
	data[idx>>1] |= 2 << ((idx & 1) << 2)

	section := nbt.NewCompoundTag().SetByte("Y", 0).SetByteArray("Blocks", nbt.ByteArrayTag(blocks)).SetByteArray("Data", nbt.ByteArrayTag(data))
	sections, _ := nbt.NewListTag([]nbt.Tag{section}, nbt.TagCompound)
	chestTag := nbt.NewCompoundTag().SetString("id", "Chest").SetInt("x", 3).SetInt("y", 16).SetInt("z", 3)
	tiles, _ := nbt.NewListTag([]nbt.Tag{chestTag}, nbt.TagCompound)
	level := nbt.NewCompoundTag().
		SetInt("xPos", 0).SetInt("zPos", 0).
		SetByte("TerrainPopulated", 1).
		SetByteArray("Biomes", nbt.ByteArrayTag(make([]byte, 256))).
		SetTag("Sections", sections).
		SetTag("TileEntities", tiles)
	root, _ := nbt.NewTreeRoot(nbt.NewCompoundTag().SetTag("Level", level), "")
	raw, err := nbt.NewBigEndianSerializer().Write(root)
	if err != nil {
		t.Fatal(err)
	}
	var compressed bytes.Buffer
	zw := zlib.NewWriter(&compressed)
	_, _ = zw.Write(raw)
	_ = zw.Close()

	loader, err := region.CreateNew(filepath.Join(path, "region", "r.0.0.mca"))
	if err != nil {
		t.Fatal(err)
	}
	if err := loader.WriteChunk(0, 0, compressed.Bytes()); err != nil {
		t.Fatal(err)
	}
	_ = loader.Close()
}

// An Anvil world is detected, converted to LevelDB on load (with a backup of the original), and
// its legacy numeric blocks and tiles come out as the right modern blocks.
func TestLoadWorldConvertsAnvilToLevelDB(t *testing.T) {
	dataDir := t.TempDir()
	worldsDir := filepath.Join(dataDir, "worlds")
	writeAnvilWorld(t, filepath.Join(worldsDir, "anvilworld"))

	m := newTestWorldManager(t)
	m.dataPath = worldsDir
	if !m.IsWorldGenerated("anvilworld") {
		t.Fatal("the Anvil world isn't recognised")
	}
	if _, err := m.LoadWorld("anvilworld", false); err == nil {
		t.Fatal("loading a read-only format without autoUpgrade should fail")
	}
	w, err := m.LoadWorld("anvilworld", true)
	if err != nil {
		t.Fatalf("LoadWorld: %v", err)
	}
	if !leveldb.IsValid(filepath.Join(worldsDir, "anvilworld")) {
		t.Fatal("the world wasn't converted to LevelDB")
	}
	if _, err := os.Stat(filepath.Join(dataDir, "backups", "worlds", "anvilworld", "region", "r.0.0.mca")); err != nil {
		t.Errorf("no backup of the original world: %v", err)
	}
	if w.GetDisplayName() != "anvilworld" || m.GetSeed(w) != 7 {
		t.Errorf("level.dat not converted: name %q seed %d", w.GetDisplayName(), m.GetSeed(w))
	}

	if got := w.GetBlockAt(0, 0, 0).GetName(); got != "Stone" {
		t.Errorf("0,0,0 = %s, want Stone", got)
	}
	if got := w.GetBlockAt(1, 1, 1).GetName(); got != "Birch Planks" {
		t.Errorf("1,1,1 = %s, want Birch Planks", got)
	}
	if _, ok := w.GetTileAt(3, 16, 3); !ok {
		t.Errorf("the chest tile wasn't loaded")
	}
}

// Tiles are saved with their chunk and come back, and placing a block creates its tile
// (Block::writeStateToWorld) with the block's tile-backed state.
func TestTilesPersistWithTheirChunk(t *testing.T) {
	dir := t.TempDir()
	w := newTestWorld()
	if err := w.OpenProvider(dir); err != nil {
		t.Fatal(err)
	}
	w.GetOrLoadChunk(0, 0)

	bed := block.VanillaBlock("bed").(*block.Bed)
	bed.Color = 4 // DyeColor 4
	pos := block.NewPosition(2, 70, 2, w)
	if err := w.SetBlock(pos, bed); err != nil {
		t.Fatal(err)
	}
	bedTile, ok := w.GetTileAt(2, 70, 2)
	if !ok {
		t.Fatal("placing a bed didn't create its tile")
	}
	if got := bedTile.(*tile.Bed).GetColor(); got != 4 {
		t.Errorf("bed tile colour = %v, want 4", got)
	}
	if got := w.GetBlockAt(2, 70, 2).(*block.Bed).Color; got != 4 {
		t.Errorf("bed read back with colour %v, want 4 (readStateFromWorld)", got)
	}

	// Replacing it with a block without a tile removes the tile.
	if err := w.SetBlock(block.NewPosition(3, 70, 3, w), block.VanillaBlock("chest")); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	w2 := newTestWorld()
	if err := w2.OpenProvider(dir); err != nil {
		t.Fatal(err)
	}
	defer w2.Close()
	if got := w2.GetBlockAt(2, 70, 2).(*block.Bed).Color; got != 4 {
		t.Errorf("after reload, bed colour = %v, want 4", got)
	}
	if _, ok := w2.GetTileAt(3, 70, 3); !ok {
		t.Errorf("after reload, the chest tile is missing")
	}
	if err := w2.SetBlock(block.NewPosition(3, 70, 3, w2), block.VanillaAir()); err != nil {
		t.Fatal(err)
	}
	if _, ok := w2.GetTileAt(3, 70, 3); ok {
		t.Errorf("replacing the chest with air didn't remove its tile")
	}
}

// Container tiles hold their inventory (created through block/inventory's hooks), save it with
// their chunk and load it back.
func TestContainerContentsPersist(t *testing.T) {
	dir := t.TempDir()
	w := newTestWorld()
	if err := w.OpenProvider(dir); err != nil {
		t.Fatal(err)
	}
	w.GetOrLoadChunk(0, 0)
	if err := w.SetBlock(block.NewPosition(5, 70, 5, w), block.VanillaBlock("chest")); err != nil {
		t.Fatal(err)
	}
	chestTile, ok := w.GetTileAt(5, 70, 5)
	if !ok {
		t.Fatal("no chest tile")
	}
	inv, ok := chestTile.(*tile.Chest).GetInventory().(*blockinventory.ChestInventory)
	if !ok {
		t.Fatalf("chest inventory is %T", chestTile.(*tile.Chest).GetInventory())
	}
	diamonds := item.VanillaItem("diamond")
	diamonds.SetCount(12)
	inv.SetItem(4, diamonds)
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	w2 := newTestWorld()
	if err := w2.OpenProvider(dir); err != nil {
		t.Fatal(err)
	}
	defer w2.Close()
	w2.GetOrLoadChunk(0, 0)
	chestTile2, ok := w2.GetTileAt(5, 70, 5)
	if !ok {
		t.Fatal("no chest tile after reload")
	}
	inv2 := chestTile2.(*tile.Chest).GetInventory().(*blockinventory.ChestInventory)
	if got := inv2.GetItem(4); got.GetName() != "Diamond" || got.GetCount() != 12 {
		t.Errorf("slot 4 after reload = %s x%d, want Diamond x12", got.GetName(), got.GetCount())
	}

	// Breaking the chest drops its contents and empties it.
	chestTile2.OnBlockDestroyed()
	if !inv2.GetItem(4).IsNull() {
		t.Errorf("the chest wasn't emptied when destroyed")
	}
}

// Item frames and lecterns keep their item (Item::nbtSerialize in the tile's save data).
func TestTileItemsPersist(t *testing.T) {
	dir := t.TempDir()
	w := newTestWorld()
	if err := w.OpenProvider(dir); err != nil {
		t.Fatal(err)
	}
	w.GetOrLoadChunk(0, 0)
	if err := w.SetBlock(block.NewPosition(6, 64, 6, w), block.VanillaStone()); err != nil {
		t.Fatal(err)
	}
	frame := block.VanillaBlock("item_frame").(*block.ItemFrame)
	frame.Facing = 1 // up, supported by the stone below
	frame.SetFramedItem(item.VanillaItem("diamond"))
	frame.ItemRotation = 3
	if err := w.SetBlock(block.NewPosition(6, 65, 6, w), frame); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	w2 := newTestWorld()
	if err := w2.OpenProvider(dir); err != nil {
		t.Fatal(err)
	}
	defer w2.Close()
	got, ok := w2.GetBlockAt(6, 65, 6).(*block.ItemFrame)
	if !ok {
		t.Fatalf("after reload, block = %s, want an item frame", w2.GetBlockAt(6, 65, 6).GetName())
	}
	if got.FramedItem == nil || got.FramedItem.GetTypeId() != item.VanillaItem("diamond").GetTypeId() {
		t.Errorf("after reload, framed item = %v, want a diamond", got.FramedItem)
	}
	if got.ItemRotation != 3 {
		t.Errorf("after reload, rotation = %d, want 3", got.ItemRotation)
	}
}

// World::createBlockUpdatePackets sends a tile's spawn compound after the block update, and a
// flower pot gets the render-update workaround state first.
func TestBlockUpdatePacketsIncludeTileData(t *testing.T) {
	w := newTestWorld()
	w.GetOrLoadChunk(0, 0)
	if err := w.SetBlock(block.NewPosition(4, 70, 4, w), block.VanillaBlock("chest")); err != nil {
		t.Fatal(err)
	}
	pks := w.CreateBlockUpdatePackets([]math.Vector3{math.NewVector3(4, 70, 4)})
	if len(pks) != 2 {
		t.Fatalf("got %d packets for a chest, want UpdateBlock + BlockActorData", len(pks))
	}
	data, ok := pks[1].(*packet.BlockActorData)
	if !ok || data.NBTData["id"] != "Chest" {
		t.Errorf("second packet = %#v, want a Chest BlockActorData", pks[1])
	}

	if err := w.SetBlock(block.NewPosition(5, 70, 5, w), block.VanillaBlock("flower_pot")); err != nil {
		t.Fatal(err)
	}
	pks = w.CreateBlockUpdatePackets([]math.Vector3{math.NewVector3(5, 70, 5)})
	if len(pks) != 3 {
		t.Fatalf("got %d packets for a flower pot, want fake UpdateBlock + UpdateBlock + BlockActorData", len(pks))
	}
	fake, real := pks[0].(*packet.UpdateBlock), pks[1].(*packet.UpdateBlock)
	if fake.NewBlockRuntimeID == real.NewBlockRuntimeID {
		t.Error("the render-update workaround state is the same as the real state")
	}
}
