package world

import (
	"testing"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/world/format"
)

type recordingChunkListener struct {
	changed, loaded, unloaded, populated []struct {
		x, z  int
		chunk *format.Chunk
	}
	blockChanges []math.Vector3
}

func (l *recordingChunkListener) OnChunkChanged(chunkX, chunkZ int, chunk *format.Chunk) {
	l.changed = append(l.changed, struct {
		x, z  int
		chunk *format.Chunk
	}{chunkX, chunkZ, chunk})
}

func (l *recordingChunkListener) OnChunkLoaded(chunkX, chunkZ int, chunk *format.Chunk) {
	l.loaded = append(l.loaded, struct {
		x, z  int
		chunk *format.Chunk
	}{chunkX, chunkZ, chunk})
}

func (l *recordingChunkListener) OnChunkUnloaded(chunkX, chunkZ int, chunk *format.Chunk) {
	l.unloaded = append(l.unloaded, struct {
		x, z  int
		chunk *format.Chunk
	}{chunkX, chunkZ, chunk})
}

func (l *recordingChunkListener) OnChunkPopulated(chunkX, chunkZ int, chunk *format.Chunk) {
	l.populated = append(l.populated, struct {
		x, z  int
		chunk *format.Chunk
	}{chunkX, chunkZ, chunk})
}

func (l *recordingChunkListener) OnBlockChanged(pos math.Vector3) {
	l.blockChanges = append(l.blockChanges, pos)
}

func TestRegisterChunkListenerReceivesOnChunkLoadedAndOnChunkPopulated(t *testing.T) {
	w := newTestWorld()
	l := &recordingChunkListener{}
	w.RegisterChunkListener(l, 0, 0)

	w.GetOrLoadChunk(0, 0)

	if len(l.loaded) != 1 || l.loaded[0].x != 0 || l.loaded[0].z != 0 {
		t.Errorf("OnChunkLoaded calls = %+v, want exactly one call for (0,0)", l.loaded)
	}
	if len(l.populated) != 1 || l.populated[0].x != 0 || l.populated[0].z != 0 {
		t.Errorf("OnChunkPopulated calls = %+v, want exactly one call for (0,0)", l.populated)
	}
}

func TestRegisterChunkListenerOnlyReceivesCallbacksForItsOwnChunk(t *testing.T) {
	w := newTestWorld()
	l := &recordingChunkListener{}
	w.RegisterChunkListener(l, 0, 0)

	// Far enough away that loading it doesn't also generate (0,0) as one of its 8 neighbours (see
	// ensurePopulated's own doc comment on why a chunk's neighbours get generated too).
	w.GetOrLoadChunk(100, 100)

	if len(l.loaded) != 0 {
		t.Errorf("OnChunkLoaded fired for a chunk this listener never registered on: %+v", l.loaded)
	}
}

func TestSetBlockFiresOnBlockChangedForTheContainingChunksListeners(t *testing.T) {
	w := newTestWorld()
	l := &recordingChunkListener{}
	w.RegisterChunkListener(l, 0, 0)

	if err := w.SetBlock(block.NewPosition(5, 90, 5, w), block.VanillaStone()); err != nil {
		t.Fatal(err)
	}

	if len(l.blockChanges) != 1 || l.blockChanges[0] != math.NewVector3(5, 90, 5) {
		t.Errorf("OnBlockChanged calls = %+v, want exactly one call for (5,90,5)", l.blockChanges)
	}
}

func TestUnloadChunkFiresOnChunkUnloaded(t *testing.T) {
	w := newTestWorld()
	l := &recordingChunkListener{}
	w.RegisterChunkListener(l, 2, 2)
	w.GetOrLoadChunk(2, 2)

	if !w.unloadChunk(2, 2, false) {
		t.Fatal("unloadChunk reported failure")
	}

	if len(l.unloaded) != 1 || l.unloaded[0].x != 2 || l.unloaded[0].z != 2 {
		t.Errorf("OnChunkUnloaded calls = %+v, want exactly one call for (2,2)", l.unloaded)
	}
}

func TestUnregisterChunkListenerStopsFurtherCallbacks(t *testing.T) {
	w := newTestWorld()
	l := &recordingChunkListener{}
	w.RegisterChunkListener(l, 0, 0)
	w.UnregisterChunkListener(l, 0, 0)

	w.GetOrLoadChunk(0, 0)

	if len(l.loaded) != 0 {
		t.Errorf("OnChunkLoaded fired after UnregisterChunkListener: %+v", l.loaded)
	}
	if got := w.GetChunkListeners(0, 0); got != nil {
		t.Errorf("GetChunkListeners(0,0) after unregister = %v, want nil", got)
	}
}

func TestUnregisterChunkListenerFromAllRemovesFromEveryRegisteredChunk(t *testing.T) {
	w := newTestWorld()
	l := &recordingChunkListener{}
	w.RegisterChunkListener(l, 0, 0)
	w.RegisterChunkListener(l, 1, 1)

	w.UnregisterChunkListenerFromAll(l)

	if got := w.GetChunkListeners(0, 0); len(got) != 0 {
		t.Errorf("GetChunkListeners(0,0) after UnregisterChunkListenerFromAll = %v, want empty", got)
	}
	if got := w.GetChunkListeners(1, 1); len(got) != 0 {
		t.Errorf("GetChunkListeners(1,1) after UnregisterChunkListenerFromAll = %v, want empty", got)
	}
}

func TestGetChunkListenersReturnsNilWhenNoneAreRegistered(t *testing.T) {
	w := newTestWorld()
	if got := w.GetChunkListeners(5, 5); got != nil {
		t.Errorf("GetChunkListeners(5,5) with nothing registered = %v, want nil", got)
	}
}
