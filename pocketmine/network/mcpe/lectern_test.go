package mcpe

import (
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/math"
)

func TestLecternSendsItsBook(t *testing.T) {
	w := newTestWorld()
	w.GetOrLoadChunk(0, 0)
	pos := block.NewPosition(2, 70, 2, w)
	_ = w.SetBlock(pos, block.VanillaBlock("lectern"))
	book := item.VanillaWrittenBook().(*item.WrittenBook)
	book.SetPageText(0, "hello")
	book.SetPageText(1, "world")
	l := w.GetBlockAt(2, 70, 2).(*block.Lectern)
	if !l.OnInteract(book, math.Up, math.Vector3{}, nil, nil) {
		t.Fatal("interact failed")
	}
	got := w.GetBlockAt(2, 70, 2).(*block.Lectern)
	if got.GetBook() == nil {
		t.Fatal("book not on lectern")
	}
	found := false
	for _, pk := range w.CreateBlockUpdatePackets([]math.Vector3{pos.Vector3}) {
		if d, ok := pk.(*packet.BlockActorData); ok {
			found = true
			if d.NBTData["hasBook"] != uint8(1) || d.NBTData["totalPages"] != int32(2) || d.NBTData["book"] == nil {
				t.Errorf("lectern spawn data = %v, want the book with its 2 pages", d.NBTData)
			}
		}
	}
	if !found {
		t.Error("no BlockActorData for the lectern")
	}
}
