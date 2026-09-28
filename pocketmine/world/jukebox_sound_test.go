package world

import (
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/item"
)

func soundTypes(pks []packet.Packet) []string {
	var out []string
	for _, pk := range pks {
		switch s := pk.(type) {
		case *packet.LevelSoundEvent:
			out = append(out, s.SoundType)
		case *packet.PlaySound:
			out = append(out, "play:"+s.SoundName)
		case *packet.StopSound:
			out = append(out, "stop:"+s.SoundName)
		}
	}
	return out
}

func hasSound(pks []packet.Packet, want string) bool {
	for _, s := range soundTypes(pks) {
		if s == want {
			return true
		}
	}
	return false
}

func TestJukeboxStopsItsRecord(t *testing.T) {
	w := newTestWorld()
	w.GetOrLoadChunk(0, 0)
	listener := &recordingListener{}
	w.RegisterChunkListener(listener, 0, 0)
	pos := block.NewPosition(3, 70, 3, w)
	if err := w.SetBlock(pos, block.VanillaBlock("jukebox")); err != nil {
		t.Fatal(err)
	}
	j := w.GetBlockAt(3, 70, 3).(*block.Jukebox)
	j.InsertRecord(item.VanillaItem("record_13").(block.Record))
	_ = w.SetBlock(pos, j)
	t.Logf("after insert: %v", soundTypes(listener.packets))
	if !hasSound(listener.packets, "play:record.13") {
		t.Errorf("inserting didn't play record.13 as a named sound (like BDS)")
	}

	listener.packets = nil
	j = w.GetBlockAt(3, 70, 3).(*block.Jukebox)
	if j.GetRecord() == nil {
		t.Fatal("record not kept after SetBlock")
	}
	j.EjectRecord()
	_ = w.SetBlock(pos, j)
	t.Logf("after eject: %v", soundTypes(listener.packets))
	if !hasSound(listener.packets, "stop:record.13") {
		t.Errorf("eject sent %v, want StopSound record.13", soundTypes(listener.packets))
	}

	j = w.GetBlockAt(3, 70, 3).(*block.Jukebox)
	j.InsertRecord(item.VanillaItem("record_13").(block.Record))
	_ = w.SetBlock(pos, j)
	listener.packets = nil
	w.UseBreakOn(pos.Vector3)
	t.Logf("after break: %v", soundTypes(listener.packets))
	if !hasSound(listener.packets, "stop:record.13") {
		t.Errorf("breaking the jukebox didn't stop the record")
	}
}
