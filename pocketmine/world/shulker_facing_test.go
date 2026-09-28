package world

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/math"
)

func TestShulkerBoxPlacedOnTheGroundFacesUp(t *testing.T) {
	w := newTestWorld()
	w.GetOrLoadChunk(0, 0)
	it := block.VanillaBlock("shulker_box").(interface{ AsItem() (block.Item, error) })
	bi, _ := it.AsItem()
	var held item.Item = bi.(item.Item)
	var returned []item.Item
	if !w.UseItemOn(math.NewVector3(4, 63, 4), held, math.Up, nil, nil, false, &returned) {
		t.Fatal("placing failed")
	}
	pks := w.CreateBlockUpdatePackets([]math.Vector3{math.NewVector3(4, 64, 4)})
	for _, pk := range pks {
		if d, ok := pk.(*packet.BlockActorData); ok {
			t.Logf("spawn data: %v", d.NBTData)
			if d.NBTData["facing"] != uint8(math.Up) {
				t.Errorf("facing = %v, want %d (up)", d.NBTData["facing"], math.Up)
			}
		}
	}
}

// The facing must survive a save and reload (sent again in the sub-chunk's tile data).
func TestShulkerBoxFacingPersists(t *testing.T) {
	dir := t.TempDir()
	w := newTestWorld()
	if err := w.OpenProvider(dir); err != nil {
		t.Fatal(err)
	}
	w.GetOrLoadChunk(0, 0)
	it := block.VanillaBlock("shulker_box").(interface{ AsItem() (block.Item, error) })
	bi, _ := it.AsItem()
	var returned []item.Item
	if !w.UseItemOn(math.NewVector3(4, 63, 4), bi.(item.Item), math.Up, nil, nil, false, &returned) {
		t.Fatal("placing failed")
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	w2 := newTestWorld()
	if err := w2.OpenProvider(dir); err != nil {
		t.Fatal(err)
	}
	defer w2.Close()
	got := w2.GetBlockAt(4, 64, 4)
	if s, ok := got.(*block.ShulkerBox); !ok || s.Facing != math.Up {
		t.Fatalf("after reload: %s facing %v, want up", got.GetName(), got)
	}
	for _, pk := range w2.CreateBlockUpdatePackets([]math.Vector3{math.NewVector3(4, 64, 4)}) {
		if d, ok := pk.(*packet.BlockActorData); ok && d.NBTData["facing"] != uint8(math.Up) {
			t.Errorf("after reload, spawn facing = %v", d.NBTData["facing"])
		}
	}
}

// TallGrassTrait::canBeReplaced: a block placed on short grass takes its place.
func TestPlacingOnShortGrassReplacesIt(t *testing.T) {
	w := newTestWorld()
	w.GetOrLoadChunk(0, 0)
	if err := w.SetBlock(block.NewPosition(6, 64, 6, w), block.VanillaTallGrass()); err != nil {
		t.Fatal(err)
	}
	stone, _ := block.VanillaStone().(interface{ AsItem() (block.Item, error) }).AsItem()
	var returned []item.Item
	// Clicking the grass itself (like the client does when aiming at it) places the stone there.
	if !w.UseItemOn(math.NewVector3(6, 64, 6), stone.(item.Item), math.Up, nil, nil, false, &returned) {
		t.Fatal("placing failed")
	}
	if got := w.GetBlockAt(6, 64, 6).GetTypeId(); got != block.STONE {
		t.Errorf("block at the grass = %s, want stone", w.GetBlockAt(6, 64, 6).GetName())
	}
	if got := w.GetBlockAt(6, 65, 6).GetTypeId(); got != block.AIR {
		t.Errorf("block above the grass = %s, want air", w.GetBlockAt(6, 65, 6).GetName())
	}
}

// Hanging sign tiles are TileSign subclasses in PHP: the text and the editor must go through them
// like a normal sign's (they used to be ignored, so hanging signs couldn't be written on).
func TestHangingSignKeepsItsTextAndEditor(t *testing.T) {
	w := newTestWorld()
	w.GetOrLoadChunk(0, 0)
	pos := block.NewPosition(3, 70, 3, w)
	sign := block.VanillaBlock("oak_wall_hanging_sign")
	if err := w.SetBlock(pos, sign); err != nil {
		t.Fatal(err)
	}
	got, ok := w.GetBlockAt(3, 70, 3).(interface {
		SetEditorEntityRuntimeID(id int, has bool)
		GetEditorEntityRuntimeID() (int, bool)
		SetText(text blockutils.SignText)
		GetText() blockutils.SignText
		block.Behavior
	})
	if !ok {
		t.Fatalf("%s isn't a sign", w.GetBlockAt(3, 70, 3).GetName())
	}
	got.SetEditorEntityRuntimeID(7, true)
	got.SetText(blockutils.NewSignText([]string{"hello"}, nil, false))
	if err := w.SetBlock(pos, got); err != nil {
		t.Fatal(err)
	}
	again := w.GetBlockAt(3, 70, 3).(interface {
		GetEditorEntityRuntimeID() (int, bool)
		GetText() blockutils.SignText
	})
	if id, has := again.GetEditorEntityRuntimeID(); !has || id != 7 {
		t.Errorf("editor = %d, %v; want 7", id, has)
	}
	if lines := again.GetText().GetLines(); lines[0] != "hello" {
		t.Errorf("text = %v, want hello", lines)
	}
}
